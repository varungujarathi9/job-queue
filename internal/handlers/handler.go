package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/varungujarathi9/job-queue/internal/services"
)

func Init() {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/jobs").Subrouter()
	subrouter.HandleFunc("/enqueue", services.EnqueueService).Methods("POST")
	subrouter.HandleFunc("/dequeue", services.DequeueService).Methods("GET")
	subrouter.HandleFunc("/{job_id}/conclude", services.ConcludeService).Methods("PUT")
	subrouter.HandleFunc("/{job_id}", services.JobService).Methods("GET")
	http.ListenAndServe("localhost:8080", subrouter)
}
