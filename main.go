package main

import (
	"log"
	"net/http"

	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	// Check env var for DB.
	if "" == os.Getenv("TASK_DB") {
		log.Fatalln("Env var TASK_DB is not define!")
	}

	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome on Task API"))
	})
	r.HandleFunc("/task/", SearchTaskAPI).Methods(http.MethodGet)
	r.HandleFunc("/task/{query}", ReadTaskAPI).Methods(http.MethodGet)
	r.HandleFunc("/task/", CreateTaskAPI).Methods(http.MethodPost)
	r.HandleFunc("/task/", UpdateTaskAPI).Methods(http.MethodPatch)
	r.HandleFunc("/task/{sid}", DeleteTaskAPI).Methods(http.MethodDelete)

	// Define the logger system.
	loggerRouter := handlers.LoggingHandler(os.Stdout, r)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", loggerRouter))
}
