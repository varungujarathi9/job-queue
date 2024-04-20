package main

import (
	"github.com/varungujarathi9/job-queue/internal/handlers"
	"github.com/varungujarathi9/job-queue/internal/utils"
)

// @title Job Queue
// @version 1.0
// @description This is a simple job queue server.
// @termsOfService http://swagger.io/terms/

// @host localhost:8080
// @BasePath /jobs
func main() {
	utils.InitLogger()
	handlers.Init()

}
