package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Init() {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/jobs").Subrouter()
	subrouter.HandleFunc("/enqueue", func(w http.ResponseWriter, r *http.Request) { fmt.Println("Got enqueue") }).Methods("POST")
	subrouter.HandleFunc("/dequeue", func(w http.ResponseWriter, r *http.Request) { fmt.Println("Got dequeue") }).Methods("GET")
	subrouter.HandleFunc("/{job_id}/conclude", func(w http.ResponseWriter, r *http.Request) { fmt.Println("Got conclude") }).Methods("PUT")
	subrouter.HandleFunc("/{job_id}", func(w http.ResponseWriter, r *http.Request) { fmt.Println("Got info") }).Methods("GET")
	http.ListenAndServe(":8080", subrouter)
}
