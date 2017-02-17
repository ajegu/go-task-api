package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
)

func TestResetDatabase(t *testing.T) {
	// Connection to mongodb server.
	host := "localhost"
	session, err := mgo.Dial(host)
	if err != nil {
		t.Errorf("can't to connect to mongodb server at %v (%v)", host, err)
	}

	// Close session at the end function.
	defer session.Close()

	// Select the collection.
	c := session.DB("test").C("tasks")

	// Clear all tasks.
	c.RemoveAll(nil)

}

func TestMockTask(t *testing.T) {
	// Create moking task.
	for i := 0; i < 100; i++ {
		task := createTaskOrFatal(t, "search task number "+fmt.Sprintf("%02d", i))
		if i%2 == 0 {
			task.Done = true
			task.Update()
		}
	}
}

const url = "task"

/*
func TestHandlerCreateTask(t *testing.T) {

	// Create a request.
	title := "handler testing task"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(fmt.Sprintf("title=%v", title)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	// Create a response recorder.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateTask)

	// Serve the request.
	handler.ServeHTTP(rr, req)

	// Test status code.
	if rr.Code != http.StatusOK {
		t.Errorf("%v", rr.Body.String())
	}

	// Test the errors.

	// Transform the response to a task.
	task := &Task{}
	json.Unmarshal(rr.Body.Bytes(), task)

	if task.Title != title {
		t.Errorf("expected title '%v', got '%v'", title, task.Title)
	}
}
*/
func TestHandlerCreateTaskJsonApi(t *testing.T) {
	// Create a request.
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("{\"data\": {\"type\": \"task\", \"attributes\": {\"title\": \"create from jsonapi\",\"done\":false}}}"))
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateTaskAPI)

	// Serve the request.
	handler.ServeHTTP(rr, req)

	// Test status code.
	if rr.Code != http.StatusCreated {
		t.Errorf("Code : %v, Error : %v", rr.Code, rr.Body.String())
	}
}

func TestHandlerUpdateTaskAPI(t *testing.T) {
	task, err := NewTask("handler task will be updated")
	if err := task.Save(); err != nil {
		t.Errorf("unexpected error (%v)", err)
	}

	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader("{\"data\": {\"type\": \"task\", \"id\":\""+task.SID+"\", \"attributes\": {\"title\": \"handler task have been updated\",\"done\":true}}}"))
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateTaskAPI)

	// Serve the request.
	handler.ServeHTTP(rr, req)

	// Test status code.
	if rr.Code != http.StatusOK {
		t.Errorf("Code : %v, Error : %v", rr.Code, rr.Body.String())
	}
}

func oldTestHandlerDeleteTaskAPI(t *testing.T) {
	task, err := NewTask("handler task will be deleted")
	if err := task.Save(); err != nil {
		t.Errorf("unexpected error (%v)", err)
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v?id=%v", url, task.SID), strings.NewReader(""))
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DeleteTaskAPI)

	// Serve the request.
	handler.ServeHTTP(rr, req)

	// Test status code.
	if rr.Code != http.StatusNoContent {
		t.Errorf("Code : %v, Error : %v", rr.Code, rr.Body.String())
	}
}

// Test Read handler with no query
func TestDeleteTaskAPI(t *testing.T) {

	task, err := NewTask("test delete task api")
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	if err := task.Save(); err != nil {
		t.Fatalf("unexpected error : %v", err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/task/{sid}", DeleteTaskAPI)

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/task/"+task.SID, strings.NewReader(""))

	m.ServeHTTP(rr, req)

	// Test status code.
	if rr.Code != http.StatusNoContent {
		t.Errorf("Code : %v, Error : %v", rr.Code, rr.Body.String())
	}
}

// Test Read handler with no query
func TestReadTaskAPI(t *testing.T) {

	task, err := NewTask("test read task api")
	if err != nil {
		t.Fatalf("unexpected error : %v", err)
	}
	if err := task.Save(); err != nil {
		t.Fatalf("unexpected error : %v", err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/task/{query}", ReadTaskAPI)

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/task/"+task.SID, strings.NewReader(""))

	m.ServeHTTP(rr, req)

	// Test status code.
	if rr.Code != http.StatusOK {
		t.Errorf("Code : %v, Error : %v", rr.Code, rr.Body.String())
	}

	// Test the result data.
	find := new(Task)
	if err := jsonapi.UnmarshalPayload(rr.Body, find); err != nil {
		t.Errorf("unexpected error (%v)", err)
	}

	if find.SID != task.SID {
		t.Errorf("expected task '%v', got '%v'", task.SID, find.SID)
	}
}

// Test Read handler with all done task
func TestReadWithDoneQueryTaskAPI(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, url+"?done=true", strings.NewReader(""))
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SearchTaskAPI)

	// Serve the request.
	handler.ServeHTTP(rr, req)

	// Test status code.
	if rr.Code != http.StatusOK {
		t.Errorf("Code : %v, Error : %v", rr.Code, rr.Body.String())
	}

	// Test the result data.
	data, err := jsonapi.UnmarshalManyPayload(rr.Body, reflect.TypeOf(&Task{}))
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}

	if len(data) == 0 {
		t.Errorf("expected data len, got %v", len(data))
	}

	// Iter on data and Convert reflect.Type to Task
	for _, row := range data {
		task := row.(*Task)
		if !task.Done {
			t.Errorf("expected done task, got %v", task.Done)
			return
		}
	}

}

// Test Read handler with query title
func TestReadWithSearchQueryTaskAPI(t *testing.T) {
	query := "search"
	req, err := http.NewRequest(http.MethodDelete, url+"?query="+query, strings.NewReader(""))
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SearchTaskAPI)

	// Serve the request.
	handler.ServeHTTP(rr, req)

	// Test status code.
	if rr.Code != http.StatusOK {
		t.Errorf("Code : %v, Error : %v", rr.Code, rr.Body.String())
	}

	// Test the result data.
	data, err := jsonapi.UnmarshalManyPayload(rr.Body, reflect.TypeOf(&Task{}))
	if err != nil {
		t.Errorf("unexpected error (%v)", err)
	}

	if len(data) == 0 {
		t.Errorf("expected data len, got %v", len(data))
	}

	// Iter on data and check the query exist in title
	re := regexp.MustCompile(query)
	for _, row := range data {
		task := row.(*Task)
		if "" == re.FindString(task.Title) {
			t.Errorf("expected task with match title with '%v', got '%v'", query, task.Title)
			return
		}
	}

}
