package graphqlfixture

import (
	"fmt"
	gqlast "github.com/graphql-go/graphql/language/ast"
	gqlparser "github.com/graphql-go/graphql/language/parser"
	gqlsource "github.com/graphql-go/graphql/language/source"
	"github.com/hashicorp/go-multierror"
	"strings"
)

// Parse does a few validations and parsing in the list of fixtures:
// - setup and teardown graphql can pass the graphql parser (at least syntactically correct),
//     - and extract the graphql variables (whose values will be filled from captures) at the same time
// - no duplicates in captor names across all fixtures
// - captor name used in a fixture's setup must already be "captured" in previous fixture's captors
// - captor name used in a fixture's teardown must already be "captured" in previous + current fixture's captors
// The result of parsing is in fs.parsed and fs.parseErr
func (fs *Fixtures) Parse() {
	if fs.parsed {
		return // no need to parse again
	}

	// all validation errors to be collected; no early exit upon error
	var multierr *multierror.Error

	// key = captor name, int = the index to fixtures on which fixture declares this captor name
	captors := map[string]int{}

	for fIdx, f := range fs.Fixtures {
		fixtureName := fmt.Sprintf("fixture[%d]", fIdx)

		// examine the setup graphql BEFORE gathering the corresponding captors
		// as those captors are meant for extracting from setup results, they cannot be used in setup query itself.
		variables, err := parseGraphqlForVariables(f.Setup)
		if err != nil {
			multierr = multierror.Append(multierr,
				fmt.Errorf("%s.setup: is invalid. %w", fixtureName, err))
		} else if containsAll, missed := captorsContainsAllKeys(captors, variables); !containsAll {
			multierr = multierror.Append(multierr,
				fmt.Errorf("%s.setup: captors not available: %s", fixtureName, strings.Join(missed, ", ")))
		} else {
			fs.Fixtures[fIdx].setupVariables = variables
		}

		// gather the fixture's captors
		if len(f.Captors) > 0 {
			for captorName, _ := range f.Captors {
				if existingFIdx, found := captors[captorName]; found {
					multierr = multierror.Append(multierr,
						fmt.Errorf("%s.captors: duplicate captor name: %s is already used by fixture[%d]",
							fixtureName, captorName, existingFIdx))
				} else {
					// collect the captor name
					captors[captorName] = fIdx
				}
			}
		}

		// examine the teardown template AFTER gathering the corresponding captors.
		if f.Teardown != nil {
			variables, err := parseGraphqlForVariables(*f.Teardown)
			if err != nil {
				multierr = multierror.Append(multierr,
					fmt.Errorf("%s.teardown: is invalid. %w", fixtureName, err))
			} else if containsAll, missed := captorsContainsAllKeys(captors, variables); !containsAll {
				multierr = multierror.Append(multierr,
					fmt.Errorf("%s.teardown: captors not available: %s", fixtureName, strings.Join(missed, ", ")))
			} else {
				fs.Fixtures[fIdx].teardownVariables = variables
			}
		}
	}

	fs.parsed = true
	fs.parseErr = multierr.ErrorOrNil()
}


// parseGraphqlForVariables parses the graphql str (hence validate its syntax) and
// extract out the variables used in the query
func parseGraphqlForVariables(graphqlStr string) ([]string, error) {
	strippedStr := strings.ReplaceAll(strings.ReplaceAll(graphqlStr, "\n", " "), "\t", " ")
	doc, err := gqlparser.Parse(gqlparser.ParseParams{
		Source: &gqlsource.Source{
			Body: []byte(strippedStr),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("parse graphql error: %w", err)
	}

	var variables []string
	for _, def := range doc.Definitions {
		switch node := def.(type) {
		case *gqlast.OperationDefinition:
			if len(node.VariableDefinitions) > 0 {
				for _, vDef := range node.VariableDefinitions {
					variables = append(variables, vDef.Variable.Name.Value)
				}
			}
		}
	}

	return variables, nil
}

func captorsContainsAllKeys(captors map[string]int, keys []string) (bool, []string) {
	if len(keys) == 0 {
		return true, nil
	}
	if len(captors) == 0 {
		return false, keys
	}

	var missed []string
	for _, key := range keys {
		if _, found := captors[key]; !found {
			missed = append(missed, key)
		}
	}
	return len(missed) == 0, missed
}
