# graphqlclient

Simple client for sending graphql request.

Example:

```go
    graphqlClient := graphqlclient.New("http://graphql-server:8080/v1/graphql", nil, http.Header{})
	var resp EmployeesResp
	err := graphqlClient.Do(ctx, Request{ Query: `employees { id name }` }, &resp)
    if err != nil { .... }
```