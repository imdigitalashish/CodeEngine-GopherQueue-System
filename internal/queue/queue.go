package queue

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		output, err := runCode(job.Language, job.Content)
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

func runCode(language, code string) (string, error) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "code-execution")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Define file extensions and Docker images for each language
	languageConfig := map[string]struct {
		extension string
		image     string
		command   []string
	}{
		"js":         {".js", "node:14-alpine", []string{"node"}},
		"golang":     {".go", "golang:1.16-alpine", []string{"go", "run"}},
		"python":     {".py", "python:3.9-slim", []string{"python"}},
		"typescript": {".ts", "node:14-alpine", []string{"npx", "ts-node"}},
		"c++":        {".cpp", "gcc:latest", []string{"g++", "-o", "program", "code.cpp", "&&", "./program"}},
		"c":          {".c", "gcc:latest", []string{"gcc", "-o", "program", "code.c", "&&", "./program"}},
	}

	config, exists := languageConfig[strings.ToLower(language)]
	if !exists {
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	// Create a temporary script file
	scriptPath := filepath.Join(tempDir, "code"+config.extension)
	if err := ioutil.WriteFile(scriptPath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("error writing to temp file: %w", err)
	}

	// Prepare Docker command
	dockerArgs := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/app", tempDir),
		"-w", "/app",
		config.image,
	}
	dockerArgs = append(dockerArgs, config.command...)
	dockerArgs = append(dockerArgs, "code"+config.extension)

	// Run Docker command
	cmd := exec.Command("docker", dockerArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running Docker: %w\n%s", err, string(output))
	}

	return string(output), nil
}

func (q *Queue) AddJob(language string, content string) string {
	id := generateUniqueID()
	job := model.Job{
		ID:       id,
		Content:  content,
		Status:   "queued",
		Language: language,
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
