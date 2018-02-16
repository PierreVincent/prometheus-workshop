package main

import (
	"github.com/google/uuid"
	"log"
	"time"
)

type Worker struct {
	id    string
	queue *JobQueue

	shutdown bool
}

func NewWorker(queue *JobQueue) (w *Worker) {
	w = &Worker{
		id:    uuid.New().String(),
		queue: queue,
	}
	go w.Run()
	return
}

func (w *Worker) Run() {
	log.Printf("[Worker %s] Starting", w.id)
	for !w.shutdown {
		w.pullJobAndRun()
	}
	log.Printf("[Worker %s] Stopped", w.id)
}

func (w *Worker) ShutDown() {
	log.Printf("[Worker %s] Shutting down", w.id)
}

func (w *Worker) pullJobAndRun() {
	job := w.queue.Pull()
	if job != nil {
		log.Printf("[Worker %s] Starting job: %s", w.id, job.ID)
		job.run()
		log.Printf("[Worker %s] Finished job: %s", w.id, job.ID)
	} else {
		log.Printf("[Worker %s] Queue is empty. Backing off 5 seconds", w.id)
		time.Sleep(5 * time.Second)
	}
}
