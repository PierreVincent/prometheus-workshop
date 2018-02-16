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
	w.shutdown = true
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

type WorkerManager struct {
	workers    []*Worker
	queue      *JobQueue
	minWorkers int
	maxWorkers int
}

func NewWorkerManager(queue *JobQueue, minWorkers int, maxWorkers int) (wm *WorkerManager) {
	wm = &WorkerManager{
		workers:    make([]*Worker, 0, maxWorkers),
		queue:      queue,
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
	}

	// Initialise workerpool
	for i := 0; i < minWorkers; i++ {
		wm.addWorker()
	}
	return wm
}

func (wm *WorkerManager) addWorker() {
	wm.workers = append(wm.workers, NewWorker(wm.queue))
}

func (wm *WorkerManager) shutdownWorker() {
	if len(wm.workers) > 0 {
		wm.workers[0].ShutDown()
		wm.workers = wm.workers[1:]
	}
}

func (wm *WorkerManager) ScaleWorkers(jobsWorkerRatio int) {
	for {
		queueSize := wm.queue.Size()
		workerCount := len(wm.workers)

		if (workerCount+1)*jobsWorkerRatio < queueSize && workerCount < wm.maxWorkers {
			log.Println("[WorkerManager] Too much work, starting extra worker.")
			wm.addWorker()
		}

		if (workerCount-1)*jobsWorkerRatio > queueSize && workerCount > wm.minWorkers {
			log.Println("[WorkerManager] Not enough worker, shutting down 1 worker")
			wm.shutdownWorker()
		}

		time.Sleep(10 * time.Second)
	}
}
