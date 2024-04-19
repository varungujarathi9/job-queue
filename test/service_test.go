package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/varungujarathi9/job-queue/internal/models"
	"github.com/varungujarathi9/job-queue/internal/services"
)

func TestEnqueueService(t *testing.T) {
	payload := []byte(`{
        "Type": "TIME_CRITICAL",
        "Status": "IN_PROGRESS"
       }`)
	req, err := http.NewRequest("POST", "/jobs/enqueue", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	services.EnqueueService(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	expectedBody := "1"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected response body %q, got %q", expectedBody, rr.Body.String())
	}

}

func TestDequeueService(t *testing.T) {
	// Create a mock job and queue
	job := models.Job{
		ID:     1,
		Type:   "TIME_CRITICAL",
		Status: "IN_PROGRESS",
	}
	queue := models.JobQueue{}
	queue.Insert(&job)

	req, err := http.NewRequest("GET", "/jobs/dequeue", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("QUEUE_CONSUMER", strconv.Itoa(1))

	rr := httptest.NewRecorder()

	services.DequeueService(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var responseJob models.Job
	err = json.NewDecoder(rr.Body).Decode(&responseJob)
	if err != nil {
		t.Fatal(err)
	}

	if responseJob.ID != job.ID {
		t.Errorf("expected job type %q, got %q", job.Type, responseJob.Type)
	}

	if responseJob.Type != job.Type {
		t.Errorf("expected job type %q, got %q", job.Type, responseJob.Type)
	}

	if responseJob.Status != job.Status {
		t.Errorf("expected job status %q, got %q", job.Status, responseJob.Status)
	}

	if responseJob.ConsumedBy != 1 {
		t.Errorf("expected job consumed by %d, got %d", job.ConsumedBy, responseJob.ConsumedBy)
	}
}

func TestConcludeService(t *testing.T) {
	// Create a mock job and add it to the jobStore
	job := models.Job{
		ID:     1,
		Type:   "TIME_CRITICAL",
		Status: "IN_PROGRESS",
	}
	var jobStore map[int]*models.Job = make(map[int]*models.Job)
	jobStore[1] = &job

	req, err := http.NewRequest("PUT", "/jobs/1/conclude", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	vars := map[string]string{
		"job_id": "1",
	}

	req = mux.SetURLVars(req, vars)

	services.ConcludeService(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	expectedBody := "Job concluded successfully"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected response body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestConcludeService_InvalidJobID(t *testing.T) {
	req, err := http.NewRequest("GET", "/jobs/2/conclude", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	vars := map[string]string{
		"job_id": "ID3",
	}

	req = mux.SetURLVars(req, vars)

	services.ConcludeService(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}

	expectedBody := "strconv.Atoi: parsing \"ID3\": invalid syntax\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected response body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestConcludeService_JobNotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/jobs/2/conclude", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	vars := map[string]string{
		"job_id": "2",
	}

	req = mux.SetURLVars(req, vars)

	services.ConcludeService(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, rr.Code)
	}

	expectedBody := "Job not found\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected response body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestConcludeService_AlreadyConcluded(t *testing.T) {
	job := models.Job{
		ID:     1,
		Type:   "TIME_CRITICAL",
		Status: "CONCLUDED",
	}
	var jobStore map[int]*models.Job = make(map[int]*models.Job)
	jobStore[1] = &job

	req, err := http.NewRequest("GET", "/jobs/1/conclude", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	vars := map[string]string{
		"job_id": "1",
	}

	req = mux.SetURLVars(req, vars)

	services.ConcludeService(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}

	expectedBody := "Job already concluded\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected response body %q, got %q", expectedBody, rr.Body.String())
	}
}
func TestJobService(t *testing.T) {
	req, err := http.NewRequest("GET", "/jobs/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	vars := map[string]string{
		"job_id": "1",
	}

	req = mux.SetURLVars(req, vars)

	services.JobService(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var responseJob models.Job
	err = json.NewDecoder(rr.Body).Decode(&responseJob)
	if err != nil {
		t.Fatal(err)
	}

	if responseJob.ID != 1 {
		t.Errorf("expected job ID %d, got %d", 1, responseJob.ID)
	}

}
