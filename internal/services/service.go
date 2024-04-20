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

// EnqueueService godoc
// @Summary      Enqueue Job
// @Description  Enqueue Job by ID
// @Accept       json
// @Param        job   body   models.Job   true   "Job object"
// @Success      200  string  models.Job.ID
// @Failure      400  string  http.StatusBadRequest
// @Router       /enqueue [post]
func EnqueueService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var job models.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		http.Error(w, `{"status" : "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if job.Type != "TIME_CRITICAL" && job.Type != "NOT_TIME_CRITICAL" {
		http.Error(w, `{"status" : "Invalid Type value"}`, http.StatusBadRequest)
		return
	}

	job.ID = nextID
	nextID++
	job.Status = QUEUED
	queue.Insert(&job)
	jobStore[job.ID] = &job

	fmt.Fprintf(w, `{"id" : `+strconv.Itoa(job.ID)+`}`)
}

// DequeueService godoc
// @Summary      Dequeue Job
// @Description  Dequeues a Job from the queue
// @Produce      json
// @Param        QUEUE_CONSUMER   header   int     true   "Queue Consumer ID"
// @Success      200  {object}     models.Job
// @Failure      400  string       http.StatusBadRequest
// @Failure      404  string       http.StatusNotFound
// @Router       /dequeue [get]
func DequeueService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	if job := queue.Poll(); job != nil {
		job.Status = IN_PROGRESS
		queueConsumer, err := strconv.Atoi(r.Header.Get("QUEUE_CONSUMER"))
		if err != nil {
			http.Error(w, `{"status" : "Invalid QUEUE_CONSUMER"}`, http.StatusBadRequest)
			return
		}
		job.ConsumedBy = queueConsumer
		json.NewEncoder(w).Encode(job)
	} else {
		http.Error(w, `{"status" : "No job available"}`, http.StatusBadRequest)
	}
}

// ConcludeService godoc
// @Summary      Conclude Job
// @Description  Concludes a Job by ID
// @Produce      plain
// @Param        job_id   path      int  true  "Job ID"
// @Success      200  string  "Job concluded successfully"
// @Failure      400  string  http.StatusBadRequest
// @Failure      404  string  http.StatusNotFound
// @Router       /{job_id}/conclude [put]
func ConcludeService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["job_id"])
	if err != nil {
		http.Error(w, `{"status" : "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if job, exists := jobStore[id]; exists {
		switch job.Status {
		case QUEUED:
			http.Error(w, `{"status" : "Dequeue job first in order to conclude"}`, http.StatusBadRequest)
		case CONCLUDED:
			http.Error(w, `{"status" : "Job already concluded"}`, http.StatusBadRequest)
		default:
			job.Status = CONCLUDED
			fmt.Fprintf(w, `{"status" : "Job concluded successfully"}`)
		}
	} else {
		http.Error(w, `{"status" : "Job not found"}`, http.StatusBadRequest)
	}

}

// JobService godoc
// @Summary      Get Job by ID
// @Description  Retrieves a Job by ID
// @Produce      json
// @Param        job_id   path      int  true  "Job ID"
// @Success      200  {object}  models.Job
// @Failure      400  string   http.StatusBadRequest
// @Failure      404  string   http.StatusNotFound
// @Router       /{job_id} [get]
func JobService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["job_id"])
	if err != nil {
		http.Error(w, `{"status" : "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if job, exists := jobStore[id]; exists {
		json.NewEncoder(w).Encode(job)
	} else {
		http.Error(w, `{"status" : "Job not found"}`, http.StatusBadRequest)
	}

}
