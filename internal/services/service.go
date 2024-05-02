package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/varungujarathi9/job-queue/internal/models"
	"github.com/varungujarathi9/job-queue/internal/utils"
)

const (
	consumerHeader = "QUEUE_CONSUMER"
	QUEUED         = "QUEUED"
	IN_PROGRESS    = "IN_PROGRESS"
	CONCLUDED      = "CONCLUDED"
)

var (
	queue                              = models.JobQueue{}
	mutex                              = &sync.Mutex{}
	nextID                             = 1
	jobStore       map[int]*models.Job = make(map[int]*models.Job)
	enqueueTimeout                     = 60000
	dequeueTimeout                     = 30000
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

	utils.Logger.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL,
	}).Info("Enqueue request received")

	// marshal incoming request body to models.Job
	var job models.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		utils.Logger.Error("Error in decoding body flow: " + err.Error())
		http.Error(w, `{"status" : "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// request body validation
	if job.Type == "" || job.Status == "" {
		utils.Logger.Info("Missing required fields")
		http.Error(w, `{"status" : "Missing required fields"}`, http.StatusBadRequest)
		return
	}

	// field Type validation
	if job.Type != "TIME_CRITICAL" && job.Type != "NOT_TIME_CRITICAL" {
		utils.Logger.Info("Invalid Type value")
		http.Error(w, `{"status" : "Invalid Type value"}`, http.StatusBadRequest)
		return
	}

	// add job to the linked list and give it an ID
	job.ID = nextID
	nextID++
	job.EnqueueTime = time.Now()
	job.Status = QUEUED
	queue.Insert(&job)
	jobStore[job.ID] = &job
	utils.Logger.Info("Returned response after enqueueing")
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
	utils.Logger.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL,
	}).Info("Dequeue request received")

	// get the next job from the queue
	if job := queue.Poll(); job != nil {
		for {
			// check if job is not nill
			if job == nil {
				utils.Logger.Info("No job available")
				http.Error(w, `{"status" : "No job available"}`, http.StatusBadRequest)
				return
			}

			// calculate elapsed time from job was enqueued
			elapsed := time.Now().Sub(job.EnqueueTime)
			if job.Cancel || elapsed > time.Duration(enqueueTimeout) {
				job = queue.Poll()
			} else {
				break
			}
		}
		job.Status = IN_PROGRESS
		queueConsumer, err := strconv.Atoi(r.Header.Get("QUEUE_CONSUMER"))
		if err != nil {
			utils.Logger.Info("Invalid QUEUE_CONSUMER: " + r.Header.Get("QUEUE_CONSUMER"))
			http.Error(w, `{"status" : "Invalid QUEUE_CONSUMER"}`, http.StatusBadRequest)
			return
		}
		job.ConsumedBy = queueConsumer
		utils.Logger.Info("Returned response after dequeueing job")
		json.NewEncoder(w).Encode(job)
	} else {
		utils.Logger.Info("No job available")
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
	utils.Logger.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL,
	}).Info("Conclude request received")

	// get job ID from URI path
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["job_id"])
	if err != nil {
		utils.Logger.Error("Error in converting job_id: " + err.Error())
		http.Error(w, `{"status" : "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// check if job of this ID was created and if so conclude according to the flow
	if job, exists := jobStore[id]; exists {
		if job.Cancel {
			utils.Logger.Info("Job already cancelled so cannot conclude")
			http.Error(w, `{"status" : "Job already cancelled so cannot conclude"}`, http.StatusBadRequest)
		}
		switch job.Status {
		case QUEUED:
			utils.Logger.Info("Conclude requested before dequeue")
			http.Error(w, `{"status" : "Dequeue job first in order to conclude"}`, http.StatusBadRequest)
		case CONCLUDED:
			utils.Logger.Info("Job already concluded")
			http.Error(w, `{"status" : "Job already concluded"}`, http.StatusBadRequest)
		default:
			job.Status = CONCLUDED
			utils.Logger.Info("Job concluded successfully")
			fmt.Fprintf(w, `{"status" : "Job concluded successfully"}`)
		}
	} else {
		utils.Logger.Info("Job not found")
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
	utils.Logger.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL,
	}).Info("Job Info request received")

	//  get job ID from URI path
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["job_id"])
	if err != nil {
		utils.Logger.Error("Error in converting job_id: " + err.Error())
		http.Error(w, `{"status" : "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	//  check if a job of this ID was created, if so return its data
	if job, exists := jobStore[id]; exists {
		utils.Logger.Info("Response returned for job info")
		json.NewEncoder(w).Encode(job)
	} else {
		utils.Logger.Info("Job not found")
		http.Error(w, `{"status" : "Job not found"}`, http.StatusBadRequest)
	}

}

func CancelService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	utils.Logger.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL,
	}).Info("Job cancel request received")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["job_id"])

	if err != nil {
		utils.Logger.Error("Error in canceling job_id: " + err.Error())
		http.Error(w, `{"status" : "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if job, exists := jobStore[id]; exists {
		job.Cancel = true
		fmt.Fprintf(w, `{"status" : "Job cancelled successfully"}`)
	} else {
		http.Error(w, `{"status" : "Job not found"}`, http.StatusBadRequest)
	}
}

func RetryService(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["job_id"])

	if err != nil {
		utils.Logger.Error("Error in retrying job_id: " + err.Error())
		http.Error(w, `{"status" : "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if job, exists := jobStore[id]; exists {
		if job.Cancel {
			utils.Logger.Info("Job already cancelled so cannot retry")
			http.Error(w, `{"status" : "Job already cancelled so cannot retry"}`, http.StatusBadRequest)
			return
		}

		job.Status = QUEUED
		job.EnqueueTime = time.Now()

		queue.Insert(job)

		fmt.Fprintf(w, `{"status" : "Job enqueued for retry"}`)
	}
}
