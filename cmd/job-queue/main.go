package main

import (
	"github.com/varungujarathi9/job-queue/internal/handlers"
)

// @title Job Queue
// @version 1.0
// @description This is a simple job queue server.
// @termsOfService http://swagger.io/terms/

// @host localhost:8080
// @BasePath /jobs
func main() {
	handlers.Init()
}
