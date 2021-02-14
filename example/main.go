package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gmm1900/gopointer"
	"github.com/gmm1900/graphqlclient"
	"github.com/gmm1900/graphqlfixture"
	"log"
	"net/http"
	"syscall"
	"time"
)

func main() {
	type Subject struct {
		ID int `json:"id"`
		Name string `json:"name"`
	}
	type Instructor struct {
		ID int `json:"id"`
		Name string `json:"name"`
		Teaches []struct{
			Subject Subject `json:"subject"`
		} `json:"teaches"`
	}
	type Student struct {
		ID int `json:"id"`
		Name string `json:"name"`
		Studies []struct{
			Subject Subject `json:"subject"`
		} `json:"studies"`
	}

	fixtures := graphqlfixture.Fixtures{
		Fixtures: []graphqlfixture.Fixture{
			{
				Setup: `mutation {
				  	insert_subjects (objects: [
						{ name: "CS101" }
						{ name: "CS102" }
				  	]) {
						returning { id }
				  	}
				}`,
				Captors: map[string]string {
					"subject_cs101_id": "/data/insert_subjects/returning/0/id",
					"subject_cs102_id": "/data/insert_subjects/returning/1/id",
				},
				Teardown: gopointer.OfString(`mutation ($subject_cs101_id: Int!, $subject_cs102_id: Int!) {
					delete_subjects ( where: { id: { _in: [$subject_cs101_id, $subject_cs102_id] }} ) {
    					affected_rows
  					}	
				}`),
			},
			{
				Setup: `mutation ($subject_cs101_id: Int!, $subject_cs102_id: Int!) {
				  	insert_instructors (objects: [
						{ 
							name: "Murphy"
							teaches: { data: { subject_id: $subject_cs101_id } } 
						} { 
							name: "Evans"
							teaches: { data: { subject_id: $subject_cs102_id } } 
						} { 
							name: "Beck"
							teaches: { data: { subject_id: $subject_cs102_id } } 
						}
				  	]) {
						returning { id name teaches { subject { id name } } }
				  	}
				}`,
				Captors: map[string]string {
					"instructor_murphy_id": "/data/insert_instructors/returning/0/id",
					"instructor_murphy": "/data/insert_instructors/returning/0",
					"instructor_evans_id": "/data/insert_instructors/returning/1/id",
					"instructor_evans": "/data/insert_instructors/returning/1",
					"instructor_beck_id": "/data/insert_instructors/returning/2/id",
					"instructor_beck": "/data/insert_instructors/returning/2",
				},
				Teardown: gopointer.OfString(`mutation ($instructor_murphy_id: Int!, $instructor_evans_id: Int!, $instructor_beck_id: Int!) {
					delete_instructors ( where: { id: { _in: [$instructor_murphy_id, $instructor_evans_id, $instructor_beck_id] }} ) {
    					affected_rows
  					}	
				}`),
			},
			{
				Setup: `mutation ($subject_cs101_id: Int!, $subject_cs102_id: Int!) {
				  	insert_students (objects: [
						{ 
							name: "Bryan"
							studies: { data: { subject_id: $subject_cs101_id } } 
						} { 
							name: "Avery"
							studies: { data: { subject_id: $subject_cs101_id } } 
						} { 
							name: "Erik"
							studies: { data: { subject_id: $subject_cs102_id } } 
						} { 
							name: "Derek"
							studies: { data: { subject_id: $subject_cs102_id } } 
						}
				  	]) {
						returning { id name studies { subject { id name } } }
				  	}
				}`,
				Captors: map[string]string {
					"student_bryan_id": "/data/insert_students/returning/0/id",
					"student_bryan": "/data/insert_students/returning/0",
					"student_avery_id": "/data/insert_students/returning/1/id",
					"student_avery": "/data/insert_students/returning/1",
					"student_erik_id": "/data/insert_students/returning/2/id",
					"student_erik": "/data/insert_students/returning/2",
					"student_derek_id": "/data/insert_students/returning/3/id",
					"student_derek": "/data/insert_students/returning/3",
				},
				Teardown: gopointer.OfString(`mutation ($student_bryan_id: Int!, $student_avery_id: Int!, $student_erik_id: Int!, $student_derek_id: Int!) {
					delete_students ( where: { id: { _in: [$student_bryan_id, $student_avery_id, $student_erik_id, $student_derek_id] }} ) {
    					affected_rows
  					}	
				}`),
			},
		},
	}

	graphqlClient := graphqlclient.New(
		"http://hasura-server:8080/v1/graphql",
		nil,
		http.Header{"x-hasura-admin-secret": []string{"adminsecret"}})

	ctx := context.Background()

	// wait for max 2min for the hasura-server to be up
	ctx2Min, cancel := context.WithTimeout(ctx, 2 * time.Minute)
	defer cancel()

WAIT_FOR_HASURA_SERVER:
	for {
		select {
		case <-ctx2Min.Done():
			log.Fatal("2 minutes reached; still unable to reach hasura-server")
		default:
			// send an empty request: just to test if the hasura-server is up
			var anyResp string
			err := graphqlClient.Do(ctx, graphqlclient.Request{ Query: `query { students { id } } `}, &anyResp)
			if err == nil {
				break WAIT_FOR_HASURA_SERVER
			}
			if errors.Is(err, syscall.ECONNREFUSED) {
				log.Println("hasura-server is not yet up; retry in 5 sec...")
				time.Sleep(5 * time.Second)
			}
		}
	}

	// call fixture setup
	err := fixtures.Setup(ctx, graphqlClient)
	if err != nil {
		log.Fatal(err)
	}

	// example of getting the captured
	fmt.Println("******** getting the captured **********")
	for _, instructorName:= range []string{"murphy", "evans", "beck"} {
		var instructor Instructor
		err = fixtures.GetAndParse("instructor_" + instructorName, &instructor)
		if err != nil { log.Fatal(err) }
		fmt.Printf("%v:\t%+v\n", instructorName, instructor)
	}
	for _, studentName:= range []string{"bryan", "avery", "erik", "derek"} {
		var student Student
		err = fixtures.GetAndParse("student_" + studentName, &student)
		if err != nil { log.Fatal(err) }
		fmt.Printf("%v:\t%+v\n", studentName, student)
	}

	// call fixture teardown
	err = fixtures.Teardown(ctx, graphqlClient)
	if err != nil {
		log.Fatal(err)
	}
}