package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/varungujarathi9/job-queue/docs"
	"github.com/varungujarathi9/job-queue/internal/services"
)

func Init() {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/jobs").Subrouter()
	subrouter.HandleFunc("/enqueue", services.EnqueueService).Methods("POST")
	subrouter.HandleFunc("/dequeue", services.DequeueService).Methods("GET")
	subrouter.HandleFunc("/{job_id}/conclude", services.ConcludeService).Methods("PUT")
	subrouter.HandleFunc("/{job_id}", services.JobService).Methods("GET")
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	http.ListenAndServe("localhost:8080", router)
}
