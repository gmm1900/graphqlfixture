# About the example

- The `hasura-server` is the graphql server, which persists data into the postgres db. 
- The `main.go` tries to set up a list of fixtures via hasura server and then tear them down. 
- At the end of the run of `main.go`, the DB should be clean (no records left). But in subsequent runs one shall see different (increased) IDs.    

# Run the example

```bash
docker-compose run example
```

(There may be delays in waiting for the hausra-server to be up running)

Sample output:

```bash
Creating network "example_graphql-fixture-example" with the default driver
Creating example_db_1 ... done
Creating example_hasura-server_1 ... done
Creating example_example_run     ... done
2021/02/14 16:40:28 hasura-server is not yet up; retry in 5 sec...
2021/02/14 16:40:33 hasura-server is not yet up; retry in 5 sec...
2021/02/14 16:40:38 hasura-server is not yet up; retry in 5 sec...
2021/02/14 16:40:43 hasura-server is not yet up; retry in 5 sec...
2021/02/14 16:40:48 hasura-server is not yet up; retry in 5 sec...
2021/02/14 16:40:53 hasura-server is not yet up; retry in 5 sec...
******** getting the captured **********
murphy:	{ID:1 Name:Murphy Teaches:[{Subject:{ID:1 Name:CS101}}]}
evans:	{ID:2 Name:Evans Teaches:[{Subject:{ID:2 Name:CS102}}]}
beck:	{ID:3 Name:Beck Teaches:[{Subject:{ID:2 Name:CS102}}]}
bryan:	{ID:1 Name:Bryan Studies:[{Subject:{ID:1 Name:CS101}}]}
avery:	{ID:2 Name:Avery Studies:[{Subject:{ID:1 Name:CS101}}]}
erik:	{ID:3 Name:Erik Studies:[{Subject:{ID:2 Name:CS102}}]}
derek:	{ID:4 Name:Derek Studies:[{Subject:{ID:2 Name:CS102}}]}
```

Run it again (note the DB or `hasura-server` were not restarted): 

```bash
docker-compose run example
```

(Note that although the same fixtures are created, IDs are different) 

```bash
Creating example_example_run ... done
******** getting the captured **********
murphy:	{ID:4 Name:Murphy Teaches:[{Subject:{ID:3 Name:CS101}}]}
evans:	{ID:5 Name:Evans Teaches:[{Subject:{ID:4 Name:CS102}}]}
beck:	{ID:6 Name:Beck Teaches:[{Subject:{ID:4 Name:CS102}}]}
bryan:	{ID:5 Name:Bryan Studies:[{Subject:{ID:3 Name:CS101}}]}
avery:	{ID:6 Name:Avery Studies:[{Subject:{ID:3 Name:CS101}}]}
erik:	{ID:7 Name:Erik Studies:[{Subject:{ID:4 Name:CS102}}]}
derek:	{ID:8 Name:Derek Studies:[{Subject:{ID:4 Name:CS102}}]}
```

To clean up:

```bash
docker-compose down
```