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
	NewWorker(queue)
	NewWorker(queue)
	NewWorker(queue)

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
