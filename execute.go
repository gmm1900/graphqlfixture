package graphqlfixture

import (
	"context"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/gmm1900/graphqlclient"
)

// Setup calls each fixture's Setup (graphql call) in sequence, and captures the values from the responses.
func (fs *Fixtures) Setup(ctx context.Context, graphqlClient *graphqlclient.Client) error {
	if !fs.parsed {
		fs.Parse()
	}
	if fs.parsed && fs.parseErr != nil {
		return fmt.Errorf("parse error: %w", fs.parseErr)
	}
	if fs.setupUntilIdx != nil {
		return fmt.Errorf("setup has already been attempted until fixture[%d]", *fs.setupUntilIdx)
	}

	// reach here: can attempt setups
	// early exit on the first encountered error.
	// if the setup is aborted due to error:
	// - fs.setupUntilIdx records the last successful setupUntilIdx
	fs.captured = map[string]interface{}{}

	for fIdx, f := range fs.Fixtures {
		fixtureName := fmt.Sprintf("fixture[%d]", fIdx)

		// 1. execute setup
		jsonParsedResp, err := doGraphqlRequest(ctx, graphqlClient, f.Setup, f.setupVariables, fs.captured)
		if err != nil {
			return fs.logAndReturnError("%s.setup failed: %w", fixtureName, err)
		}
		// reach here: the setup is done (if the graphql is mutation, the data is already persisted)
		// then teardown needs to start at least from this fixture.
		fs.logs = append(fs.logs, fmt.Sprintf("%s.setup: completed", fixtureName))
		setupUntilIdx := fIdx // make a copy
		fs.setupUntilIdx = &setupUntilIdx

		// 2. captures from response
		if len(f.Captors) == 0 {
			fs.logs = append(fs.logs, fmt.Sprintf("%s.captors: not exist", fixtureName))
			return nil
		}
		// reach here: there are captures to handle
		for captorName, captorPath := range f.Captors {
			// captorVal can be single value, or map, or array.
			capturedGabsObj, err := jsonParsedResp.JSONPointer(captorPath)
			if err != nil {
				return fs.logAndReturnError("%s.captors failed: %s (%s) not found: %w", fixtureName, captorName, captorPath, err)
			}
			fs.captured[captorName] = capturedGabsObj.Data()
		}
		// reach here: captures are done
		fs.logs = append(fs.logs, fmt.Sprintf("%s.captors: completed with %d capture(s)", fixtureName, len(f.Captors)))
	}

	return nil
}

// Teardown calls each fixture's Teardown (graphql call) in reverse sequence.
func (fs *Fixtures) Teardown(ctx context.Context, graphqlClient *graphqlclient.Client) error {
	// only do teardown if it has been setup before (even partial), and has not been torn down before.
	// having setup before means the parsing is already passed.
	if fs.setupUntilIdx == nil {
		return errors.New("setup hasn't been attempted")
	}
	if fs.teardownUntilIdx != nil {
		return fmt.Errorf("teardown has already been attempted until fixture[%d]", *fs.teardownUntilIdx)
	}

	// reach here: can attempt teardown in reverse order
	for fIdx := *fs.setupUntilIdx; fIdx >= 0; fIdx-- {
		f := fs.Fixtures[fIdx]
		fixtureName := fmt.Sprintf("fixture[%d]", fIdx)

		if f.Teardown == nil { // this fixture doesn't have teardown step
			fs.logs = append(fs.logs, fmt.Sprintf("%s.teardown: not exist", fixtureName))
			continue
		}

		// execute teardown
		_, err := doGraphqlRequest(ctx, graphqlClient, *f.Teardown, f.teardownVariables, fs.captured)
		if err != nil {
			return fs.logAndReturnError("%s.teardown failed: %w", fixtureName, err)
		}
		fs.logs = append(fs.logs, fmt.Sprintf("%s.teardown: completed", fixtureName))
		teardownUntilIdx := fIdx // making a copy
		fs.teardownUntilIdx = &teardownUntilIdx
	}

	return nil
}

func (fs *Fixtures) logAndReturnError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fs.logs = append(fs.logs, err.Error())
	return err
}

// doGraphqlRequest composes the variables (if applicable), send the graphql request,
// and parse the graphql response for errors
// Used in both Setup and Teardown.
func doGraphqlRequest(ctx context.Context, graphqlClient *graphqlclient.Client,
	graphqlQueryStr string, varNames []string, captured map[string]interface{}) (*gabs.Container, error) {
	// 1. prepare request variables
	var variables map[string]interface{}
	if len(varNames) > 0 {
		variables = map[string]interface{}{}
		for _, varName := range varNames {
			varVal, found := captured[varName]
			if !found { // shouldn't happen, since the fixtures should have passed parsing
				return nil, fmt.Errorf("cannot find variable %s in captured", varName)
			}
			variables[varName] = varVal
		}
	}

	// 2. call graphql server
	var resp []byte
	err := graphqlClient.Do(
		ctx,
		graphqlclient.Request{
			Query: graphqlQueryStr,
			Variables: variables,
		},
		&resp,
	)
	if err != nil {
		return nil, fmt.Errorf("graphql request failed: %w", err)
	}
	// examine if errors exist in the response
	jsonParsedResp, err := gabs.ParseJSON(resp)
	if err != nil {
		return nil, fmt.Errorf("graphql response is not json: %w", err)
	}
	errorsGabsObj, err := jsonParsedResp.JSONPointer("/errors")
	if err == nil { // means "errors" found
		return nil, fmt.Errorf("graphql response contains error: %v", errorsGabsObj.Data())
	}

	return jsonParsedResp, nil
}

