package services

import (
	"encoding/json"
	"net/http"
	"sync"

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

	json.NewEncoder(w).Encode(job)
}
