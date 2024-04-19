package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/varungujarathi9/job-queue/internal/models"
)

const (
	consumerHeader = "QUEUE_CONSUMER"
	QUEUED         = "QUEUED"
	IN_PROGRESS    = "IN_PROGRESS"
	CONCLUDED      = "CONCLUDED"
)

var (
	queue                        = models.JobQueue{}
	mutex                        = &sync.Mutex{}
	nextID                       = 1
	jobStore map[int]*models.Job = make(map[int]*models.Job)
)

func EnqueueService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var job models.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job.ID = nextID
	nextID++
	job.Status = QUEUED
	queue.Insert(&job)
	jobStore[job.ID] = &job

	fmt.Fprintf(w, strconv.Itoa(job.ID))
}

func DequeueService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	if job := queue.Poll(); job != nil {
		job.Status = IN_PROGRESS
		queueConsumer, err := strconv.Atoi(r.Header.Get("QUEUE_CONSUMER"))
		if err != nil {
			http.Error(w, "Invalid QUEUE_CONSUMER", http.StatusBadRequest)
			return
		}
		job.ConsumedBy = queueConsumer
		json.NewEncoder(w).Encode(job)
	} else {
		http.Error(w, "No job available", http.StatusNotFound)
	}
}

func ConcludeService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["job_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if job, exists := jobStore[id]; exists {
		switch job.Status {
		case QUEUED:
			http.Error(w, "Dequeue job first in order to conclude", http.StatusBadRequest)
		case CONCLUDED:
			http.Error(w, "Job already concluded", http.StatusBadRequest)
		default:
			job.Status = CONCLUDED
			fmt.Fprintf(w, "Job concluded successfully")
		}
	} else {
		http.Error(w, "Job not found", http.StatusNotFound)
	}

}

func JobService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["job_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if job, exists := jobStore[id]; exists {
		json.NewEncoder(w).Encode(job)
	} else {
		http.Error(w, "Job not found", http.StatusNotFound)
	}

}
