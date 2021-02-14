package graphqlfixture

import (
	"errors"
	"github.com/gmm1900/gopointer"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct{
		name          string
		givenFixtures Fixtures
		expectedErr   error
	} {
		{
			name: "no error",
			givenFixtures: Fixtures{
				Fixtures: []Fixture{
					{
						Setup: `mutation {
							insert_abc(objects: [
								{ name: "abc1"}
								{ name: "abc2"}
							])} {
								returning { id }
							}`,
						Captors: map[string]string{
							"id1": "data/insert_abc/0/id",
							"id2": "data/insert_abc/1/id",
						},
						Teardown: gopointer.OfString(`mutation ($id1: Int, $id2: Int) {
							delete_abc(where: 
								{ id: { _in: [ $id1, $id2 ] } } 
							)} {
								affected_rows
							}`),
					},
					{
						Setup: `mutation ($id2: int) {
							insert_xyz(objects: {
								abc_id: $id2
								name: "xyz1"
							})} {
								returning { id }
							}`,
						Captors: map[string]string{
							"id3": "data/insert_xyz/0/id",
						},
						Teardown: gopointer.OfString(`mutation ($id3: Int) {
							delete_xyz(where: 
								{ id: { _eq: $id3 } } 
							)} {
								affected_rows
							}`),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "with errors: missing captors, syntax errors",
			givenFixtures: Fixtures{
				Fixtures: []Fixture{
					{
						Setup: `mutation {
							insert_abc(objects: [
								{ name: "abc1"}
								{ name: "abc2"}
							])} {
								returning { id }
							}`,
						Captors: map[string]string{
							"id1": "data/insert_abc/0/id",
							"id2": "data/insert_abc/1/id",
						},
						Teardown: gopointer.OfString(`mutation ($id11: Int, $id12: Int) {
							delete_abc(where: 
								{ id: { _in: [ $id1, $id2 ] } } 
							)} {
								affected_rows
							}`),
					},
					{
						Setup: `mutation ($id22: int) {
							insert_xyz(objects: {
								abc_id: $id22
								name: "xyz1"
							})} {
								returning { id }
							}`,
						Captors: map[string]string{
							"id3": "data/insert_xyz/0/id",
						},
						Teardown: gopointer.OfString(`mutation ($id3: Int) {
							delete_xyz(where:
								{ id: _eq: $id3 }
							)} {
								affected_rows
							}`),
					},
				},
			},
			expectedErr: multierror.Append(
				errors.New("fixture[0].teardown: captors not available: id11, id12"),
				errors.New("fixture[1].setup: captors not available: id22"),
				errors.New("fixture[1].teardown: is invalid. parse graphql error: Syntax Error  (1:66) Expected Name, found :\n\n1: mutation ($id3: Int) {        delete_xyz(where:         { id: _eq: $id3 }        )} {         affected_rows        }\n                                                                    ^\n"),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T){
			// WHEN
			tc.givenFixtures.Parse()
			err := tc.givenFixtures.parseErr
			// THEN
			if tc.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr.Error())
			}
		})
	}
}

func TestParseGraphqlForVariables(t *testing.T) {
	testCases := []struct{
		name string
		givenGraphqlStr string
		expectedVars []string
		expectedErr error
	} {
		{
			name: "no variables",
			givenGraphqlStr: `
				mutation {
					insert_abc(objects: [
						{ name: "abc1"}
						{ name: "abc2"}
					])
				} {
					returning { id }
				}
			`,
			expectedVars: nil,
			expectedErr: nil,
		},
		{
			name: "has variables",
			givenGraphqlStr: `
				mutation ($name1: String, $name2: String) {
					insert_abc(objects: [
						{ name: $name1}
						{ name: $name2}
					])
				} {
					returning { id }
				}
			`,
			expectedVars: []string{"name1", "name2"},
			expectedErr: nil,
		},
		{
			name: "graphql syntax error",
			givenGraphqlStr: `
				mutation ($name1: String, $name2: String) {
					insert_abc(objects: [
						{ name: $name1}
						{ name: $name2}}}
					])
				} {
					returning { id }
				}
			`,
			expectedVars: nil,
			expectedErr: errors.New(`parse graphql error: Syntax Error  (1:120) Unexpected }

1:      mutation ($name1: String, $name2: String) {      insert_abc(objects: [       { name: $name1}       { name: $name2}}}      ])     } {      returning { id }     }    
                                                                                                                          ^
`),
		},
	}

	for _, tc := range testCases {
		// WHEN
		vars, err := parseGraphqlForVariables(tc.givenGraphqlStr)
		// THEN
		assert.Equal(t, tc.expectedVars, vars)
		if tc.expectedErr == nil {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, tc.expectedErr.Error())
		}

	}
}