package queue

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"io/ioutil"

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
		output, err := runPythonCodeInDocker(job.Content)
		if err != nil {
			job.Result = "Error: " + err.Error()
			job.Status = "failed"
		} else {
			job.Result = "Output: " + output
			job.Status = "completed"
		}
		job.DoneTime = time.Now()
		q.mutex.Lock()
		q.results[job.ID] = &job
		q.mutex.Unlock()
	}
}

func runPythonCodeInDocker(code string) (string, error) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "python-code")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary Python script
	scriptPath := filepath.Join(tempDir, "code.py")
	print(scriptPath)
	if err := ioutil.WriteFile(scriptPath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("error writing to temp file: %w", err)
	}

	// Docker command to run the script
	cmd := exec.Command("docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/app", tempDir),
		"-w", "/app",
		"python:3.9-slim",
		"python", "code.py")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running Docker: %w\n%s", err, string(output))
	}

	return string(output), nil
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
