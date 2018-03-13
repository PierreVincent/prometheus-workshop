package main

import (
	"github.com/google/uuid"
	"math/rand"
	"sync"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

type Job struct {
	ID string
}

func NewJob() *Job {
	return &Job{ID: uuid.New().String()}
}

func (*Job) run() {
	// Run the job (5 - 15 seconds)
	time.Sleep(time.Duration(5+rng.Intn(10)) * time.Second)
}

type JobQueue struct {
	q     []*Job
	mutex *sync.Mutex
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		q:     make([]*Job, 0, 1000),
		mutex: &sync.Mutex{},
	}
}

func (jq *JobQueue) Size() int {
	jq.mutex.Lock()
	defer jq.mutex.Unlock()

	return len(jq.q)
}

func (jq *JobQueue) Push(job *Job) {
	jq.mutex.Lock()
	defer jq.mutex.Unlock()

	jq.q = append(jq.q, job)
}

func (jq *JobQueue) Pull() (job *Job) {
	jq.mutex.Lock()
	defer jq.mutex.Unlock()

	if len(jq.q) > 0 {
		job = jq.q[0]
		jq.q = jq.q[1:]
	}
	return
}
