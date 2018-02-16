package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// Create Job Queue
	queue := NewJobQueue()

	// Create a few workers
	wm := NewWorkerManager(queue, 1, 4)

	// Scale workers (10 jobs per worker)
	go wm.ScaleWorkers(10)

	// Start API
	startHttpApi(queue)
}

func startHttpApi(queue *JobQueue) {
	router := mux.NewRouter()
	router.HandleFunc("/jobs", func(rw http.ResponseWriter, r *http.Request) {
		queue.Push(NewJob())
	}).Methods("POST")

	log.Println("[Api] Starting server")
	panic(http.ListenAndServe(":8888", router))
}
