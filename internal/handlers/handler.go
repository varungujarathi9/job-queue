package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/varungujarathi9/job-queue/docs"
	"github.com/varungujarathi9/job-queue/internal/services"
	"github.com/varungujarathi9/job-queue/internal/utils"
)

func Init() {
	utils.Logger.Info("Starting REST API server")
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/jobs").Subrouter()
	subrouter.HandleFunc("/enqueue", services.EnqueueService).Methods("POST")
	subrouter.HandleFunc("/dequeue", services.DequeueService).Methods("GET")
	subrouter.HandleFunc("/{job_id}/conclude", services.ConcludeService).Methods("PUT")
	subrouter.HandleFunc("/{job_id}", services.JobService).Methods("GET")
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	utils.Logger.Info("Started server at port 8080")
	http.ListenAndServe("localhost:8080", router)
}
