# graphqlfixture

A test utility for fixture creation and teardown using GraphQL. 

When creating fixtures for tests, we often need to know the exact DB ID that's generated in order to:

- use that DB ID in subsequent fixture creations (e.g., subsequent fixture uses that DB ID as foreign key).
- use that DB ID in tests (e.g., query by ID)
- assert that DB ID as part of object comparison in tests (optional)

This utility tries to address this need, especially in the context of [hasura](https://hasura.io/) or GraphQL tests.  One particular limitation is the fixtures created are actually persisted into (cannot take advantage of the DB rollback), thus requires explicit "teardown" step.

# Example

See [example code](https://github.com/gmm1900/graphqlfixture/blob/main/example/main.go) and [steps to run it](https://github.com/gmm1900/graphqlfixture/blob/main/example/README.md).