package queue

import (
	"sync"
	"time"

	"github.com/imdigitalashish/QueueSystemsInGolang/internal/model"
)

type Queue struct {
	jobs    chan model.Job
	results map[string]*model.Job
	mutex   sync.RWMutex
}

func NewQueue(workers int) *Queue {
	q := &Queue{
		jobs:    make(chan model.Job, 100),
		results: make(map[string]*model.Job),
	}

	for i := 0; i < workers; i++ {
		go q.worker()
	}

	return q
}

func (q *Queue) worker() {
	for job := range q.jobs {
		time.Sleep(5 * time.Second)
		job.Result = "Proccessed: " + job.Content
		job.Status = "completed"
		job.DoneTime = time.Now()

		q.mutex.Lock()
		q.results[job.ID] = &job
		q.mutex.Unlock()
	}
}

func (q *Queue) AddJob(content string) string {
	id := generateUniqueID()
	job := model.Job{
		ID:      id,
		Content: content,
		Status:  "queued",
	}

	q.mutex.Lock()
	q.results[id] = &job
	q.mutex.Unlock()

	q.jobs <- job
	return id
}

func (q *Queue) CheckStatus(id string) (*model.Job, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	job, exists := q.results[id]
	return job, exists
}

func generateUniqueID() string {
	// Implement a unique ID generation method
	// For simplicity, we'll use a timestamp-based ID here
	return time.Now().Format("20060102150405.000000000")
}
