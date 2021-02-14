package graphqlfixture

// Fixture contains the setup, teardown logic for a piece of fixtures, and the data needs to be extracted (captured) from the fixture, e.g., IDs.
type Fixture struct {
	Setup string // the graphql to seed the fixture (expect mutation.. could be query too? to just get some existing data, e.g., max of something)
	Captors map[string]string // directives for capturing data from the setup response: key = captor name, the "logical name" of the captured value, value = the jsonpath ino the response to extract the value
	Teardown *string // the graphql to remove the seeded fixture (expect delete mutation). optional, if no new fixture is created during setup.

	// internal: variable names parsed from graphql (== captor names)
	setupVariables []string
	teardownVariables []string
}

type Fixtures struct {
	Fixtures []Fixture // a list of fixtures, to be setup in this sequence, and torn down in the reverse sequence

	// internal: parsing
	parsed bool // if false, Fixtures need to go through the Parse() step first.
	parseErr error // if parsed = true && parseErr != nil, these fixtures are not ready for setup (hint: test case not written correctly).

	// internal: execution
	captured map[string]interface{} // key = captor name, value = extracted value from the setup graphql response
	setupUntilIdx *int // the index to the last fixture that was successfully set up. Nil if not setup before.
	teardownUntilIdx *int // the index to the last fixture that was successfully torn down. Nil if not torndown before.
	logs []string // track info on setup and teardown (success or failure). Since this is a rather fragile fixture-gen (not db transaction, cannot rollback), an unsuccessful execution will require manual intervention (e.g., delete data from db)
}