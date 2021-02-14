package graphqlfixture

import (
	"context"
	"github.com/gmm1900/gopointer"
	"github.com/gmm1900/graphqlclient"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type testCase struct {
	name                     string
	givenFixtures            Fixtures
	givenMockServer          graphqlclient.MockGraphqlServer
	expectedCapturedRequests []map[string]interface{}
	expectedSetupResult      Fixtures // the fields other than the `Fixtures`, i.e., parsed, parseErr, captured ... etc.
}

func TestSetup(t *testing.T) {
	// function  for creating baseline test case.
	newBaselineCase := func() testCase {
		return testCase{
			name: "successful setup",
			givenFixtures: Fixtures{
				Fixtures: []Fixture{
					{
						// a valid graphql
						Setup: `mutation { insert_abc(objects: { name: "abc1"}) } { returning { id alias }}`,
						Captors: map[string]string{
							"abc_id":    "/data/insert_abc/returning/0/id",
							"abc_alias": "/data/insert_abc/returning/0/alias",
						},
					},
					{
						// a valid graphql with variable
						Setup: `mutation ($abc_id: Int!) { insert_xyz(objects: { name: "xyz2", parent_id: $abc_id }) } { returning { id }}`,
						Captors: map[string]string{
							"xyz_id": "/data/insert_xyz/returning/0/id",
						},
					},
				},
			},
			givenMockServer: graphqlclient.MockGraphqlServer{
				MockedRespBody: [][]byte{
					// response 1
					[]byte(`{ "data": { "insert_abc": { "returning": [ { "id": 13, "alias": "abc1_alias" } ] } } }`),
					// response 2
					[]byte(`{ "data": { "insert_xyz": { "returning": [ { "id": 21 } ] } } }`),
				},
			},
			expectedCapturedRequests: []map[string]interface{}{
				{ // request 1
					"query": `mutation { insert_abc(objects: { name: "abc1"}) } { returning { id alias }}`,
				},
				{ // request 2
					"query":     `mutation ($abc_id: Int!) { insert_xyz(objects: { name: "xyz2", parent_id: $abc_id }) } { returning { id }}`,
					"variables": map[string]interface{}{"abc_id": 13.0},
				},
			},
			expectedSetupResult: Fixtures{
				parsed:   true,
				parseErr: nil,
				captured: map[string]interface{}{ // from mocked graphql response
					"abc_id":    13.0,
					"abc_alias": "abc1_alias",
					"xyz_id":    21.0,
				},
				setupUntilIdx:    gopointer.OfInt(1),
				teardownUntilIdx: nil,
				logs: []string{
					"fixture[0].setup: completed",
					"fixture[0].captors: completed with 2 capture(s)",
					"fixture[1].setup: completed",
					"fixture[1].captors: completed with 1 capture(s)",
				},
			},
		}
	}

	testCases := []testCase{
		// 1. a successful case
		newBaselineCase(),

		// 2. fail at setup
		func() testCase {
			tc := newBaselineCase()
			tc.name = "fail at setup"
			// edit from baseline to make this an error case
			tc.givenMockServer.MockedRespBody[1] = []byte(`{ "errors": { "extensions": {} } }`)
			delete(tc.expectedSetupResult.captured, "xyz_id")
			tc.expectedSetupResult.setupUntilIdx = gopointer.OfInt(0) // complete the setup the first one only
			tc.expectedSetupResult.logs = append(tc.expectedSetupResult.logs[0:2],
				"fixture[1].setup failed: graphql response contains error: map[extensions:map[]]")
			return tc
		}(),

		// 3. fail at captures
		func() testCase {
			tc := newBaselineCase()
			tc.name = "fail at captors"
			// edit from baseline to make this an error case
			tc.givenFixtures.Fixtures[1].Captors = map[string]string{
				"xyz_id": "/data/insert_xyz/returning/id", // a wrong json path
			}
			delete(tc.expectedSetupResult.captured, "xyz_id")
			tc.expectedSetupResult.setupUntilIdx = gopointer.OfInt(1)
			tc.expectedSetupResult.logs[3] = "fixture[1].captors failed: xyz_id (/data/insert_xyz/returning/id) not found: failed to resolve path segment '3': found array but segment value 'id' could not be parsed into array index: strconv.Atoi: parsing \"id\": invalid syntax"
			return tc
		}(),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T){
			// GIVEN
			tc.givenMockServer.Start(t)
			defer tc.givenMockServer.Close()

			ctx := context.Background()
			graphqlClient := graphqlclient.New(tc.givenMockServer.URL, nil, http.Header{})

			// parse should result in no error: this test is about testing "setup" and should give valid, parsable fixtures
			tc.givenFixtures.Parse()
			assert.NoError(t, tc.givenFixtures.parseErr)

			// WHEN
			tc.givenFixtures.Setup(ctx, graphqlClient)

			// THEN
			assert.Equal(t, tc.expectedCapturedRequests, tc.givenMockServer.CapturedReqBody)
			cmpOpts := []cmp.Option{
				cmpopts.IgnoreFields(Fixtures{}, "Fixtures"),
				cmp.AllowUnexported(Fixtures{}),
			}
			want, got := tc.expectedSetupResult, tc.givenFixtures
			if !cmp.Equal(want, got, cmpOpts...) {
				t.Errorf("Setup result mismatched %v", cmp.Diff(want, got, cmpOpts...))
			}
		})
	}
}

func TestTeardown(t *testing.T) {
	// function for creating baseline test case.
	newBaselineCase := func() testCase {
		return testCase{
			name: "successful teardown",
			givenFixtures: Fixtures{
				// pre-populate successful parse and setup steps
				parsed:   true,
				parseErr: nil,
				captured: map[string]interface{}{ // from mocked graphql response
					"abc_id":    13.0,
					"abc_alias": "abc1_alias",
					"xyz_id":    21.0,
				},
				setupUntilIdx:    gopointer.OfInt(1),
				teardownUntilIdx: nil,
				logs: []string{
					"some existing setup logs",
				},
				Fixtures: []Fixture{
					{
						Setup:    `doesnt matter for this test`,
						Teardown: gopointer.OfString(`mutation ($abc_id: int!) { delete_abc( where: { id: { _eq: $abc_id } } ) { affected_rows }`),
						teardownVariables: []string{"abc_id"}, // mock this value to mimic successful parse result
					},
					{
						Setup:    `doesnt matter for this test`,
						Teardown: gopointer.OfString(`mutation ($xyz_id: int!) { delete_xyz( where: { id: { _eq: $xyz_id } } ) { affected_rows }`),
						teardownVariables: []string{"xyz_id"}, // mock this value to mimic successful parse result
					},
				},
			},
			givenMockServer: graphqlclient.MockGraphqlServer{
				MockedRespBody: [][]byte{
					// response 1
					[]byte(`{ "data": { "delete_xyz": { "affected_rows": 1 } } }`),
					// response 2
					[]byte(`{ "data": { "delete_abc": { "affected_rows": 1 } } }`),
				},
			},
			expectedCapturedRequests: []map[string]interface{}{
				{ // teardown request 1
					"query": `mutation ($xyz_id: int!) { delete_xyz( where: { id: { _eq: $xyz_id } } ) { affected_rows }`,
					"variables": map[string]interface{}{"xyz_id": 21.0},
				},
				{ // teardown request 2
					"query": `mutation ($abc_id: int!) { delete_abc( where: { id: { _eq: $abc_id } } ) { affected_rows }`,
					"variables": map[string]interface{}{"abc_id": 13.0},
				},
			},
			expectedSetupResult: Fixtures{
				teardownUntilIdx: gopointer.OfInt(0),
				logs: []string{
					"some existing setup logs",
					"fixture[1].teardown: completed",
					"fixture[0].teardown: completed",
				},
			},
		}
	}

	testCases := []testCase{
		// 1. a successful case
		newBaselineCase(),

		// 2. fail at teardown
		func() testCase {
			tc := newBaselineCase()
			tc.name = "fail at teardown"
			// edit from baseline to make this an error case
			tc.givenMockServer.MockedRespBody[1] = []byte(`{ "errors": { "extensions": {} } }`)
			tc.expectedSetupResult.teardownUntilIdx = gopointer.OfInt(1)
			tc.expectedSetupResult.logs[2] = "fixture[0].teardown failed: graphql response contains error: map[extensions:map[]]"
			return tc
		}(),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T){
			// GIVEN
			tc.givenMockServer.Start(t)
			defer tc.givenMockServer.Close()

			ctx := context.Background()
			graphqlClient := graphqlclient.New(tc.givenMockServer.URL, nil, http.Header{})

			// WHEN
			tc.givenFixtures.Teardown(ctx, graphqlClient)

			// THEN
			// assert.Equal(t, tc.expectedCapturedRequests, tc.givenMockServer.CapturedReqBody)
			cmpOpts := []cmp.Option{
				cmpopts.IgnoreFields(Fixtures{}, "Fixtures", "parsed", "parseErr", "captured", "setupUntilIdx"),
				cmp.AllowUnexported(Fixtures{}),
			}
			want, got := tc.expectedSetupResult, tc.givenFixtures
			if !cmp.Equal(want, got, cmpOpts...) {
				t.Errorf("Teardown result mismatched %v", cmp.Diff(want, got, cmpOpts...))
			}
		})
	}
}
