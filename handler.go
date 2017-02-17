package main

import (
	"net/http"

	validator "gopkg.in/go-playground/validator.v9"

	"fmt"

	"io"

	"strconv"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
)

// populateTask create a task object with json properties.
func populateTask(body io.ReadCloser, w http.ResponseWriter) (*Task, error) {

	task := new(Task)
	if err := jsonapi.UnmarshalPayload(body, task); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{
			Title:  "Json Unmarshal Payload Error",
			Detail: err.Error(),
			Status: "500",
		}}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return nil, err
	}
	return task, nil
}

// validateTask check the values's task.
func validateTask(task *Task, w http.ResponseWriter) error {

	if err := task.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		var eos []*jsonapi.ErrorObject
		for _, err := range err.(validator.ValidationErrors) {
			eos = append(eos, &jsonapi.ErrorObject{
				Title:  "Validation Error",
				Detail: fmt.Sprintf("%s", err),
				Status: "400",
				Meta:   &map[string]interface{}{"field": err.Field(), "error": err.Tag(), "expected": err.Type(), "received": err.Value()},
			})
		}

		if err := jsonapi.MarshalErrors(w, eos); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return err
	}
	return nil
}

// CreateTaskAPI create a new task with jsonapi params.
func CreateTaskAPI(w http.ResponseWriter, r *http.Request) {
	// Set the header content-type.
	w.Header().Set("Content-Type", jsonapi.MediaType)

	task, err := populateTask(r.Body, w)
	if err != nil {
		return
	}

	if err := validateTask(task, w); err != nil {
		return
	}

	// Save the task.
	if err := task.Save(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{
			Title:  "Save Error",
			Detail: err.Error(),
			Status: "500",
		}}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

		}
		return
	}

	// Set header status code.
	w.WriteHeader(http.StatusCreated)

	// Write the response.
	jsonapi.MarshalOnePayload(w, task)

}

// UpdateTaskAPI bring up to date a specific Task.
func UpdateTaskAPI(w http.ResponseWriter, r *http.Request) {

	// Set the header content-type.
	w.Header().Set("Content-Type", jsonapi.MediaType)

	task, err := populateTask(r.Body, w)
	if err != nil {
		return
	}

	if err := validateTask(task, w); err != nil {
		return
	}

	// Update the task.
	if err := task.Update(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{
			Title:  "Update Error",
			Detail: err.Error(),
			Status: "500",
		}}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

		}
		return
	}

	// Set header status code.
	w.WriteHeader(http.StatusOK)

	// Write the response.
	jsonapi.MarshalOnePayload(w, task)
}

// DeleteTaskAPI remove a task and return a 204 (no-content) response
func DeleteTaskAPI(w http.ResponseWriter, r *http.Request) {
	
	vars := mux.Vars(r)
	
	sid := vars["sid"]
	if sid == "" {
		w.WriteHeader(http.StatusInternalServerError)
		if err := jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{
			Title:  "Delete Error",
			Detail: "ID parameter is required",
			Status: "500",
		}}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := DeleteTask(sid); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if err := jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{{
			Title:  "Delete Error",
			Detail: err.Error(),
			Status: "500",
		}}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Set header status code.
	w.WriteHeader(http.StatusNoContent)
}

// ReadTaskAPI return a response with tasks encoding to json
func ReadTaskAPI(w http.ResponseWriter, r *http.Request) {

	// Set the header defaults.
	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	task := &Task{}
	vars := mux.Vars(r)
	if vars["query"] != "" {
		task, _ = SelectTask(vars["query"])
	}

	jsonapi.MarshalOnePayload(w, task)
}

// SearchTaskAPI return a response with tasks encoding to json
func SearchTaskAPI(w http.ResponseWriter, r *http.Request) {

	// Set the header defaults.
	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)

	// Defaults params.
	var err error
	page := 1
	limit := 10

	// Get all params.
	v := r.URL.Query()

	if p := v.Get("page"); p != "" {
		page, err = strconv.Atoi(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if l := v.Get("limit"); l != "" {
		limit, err = strconv.Atoi(l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Query task.
	var tasks []*Task
	var n int
	q := v.Get("query")
	d := v.Get("done")

	if d != "" {
		bd, _ := strconv.ParseBool(d)
		tasks, n, err = SearchTask(q, bd, false, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		tasks, n, err = SearchTask(q, false, true, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	jsonapi.MarshalManyPayload(w, tasks, n)
}
