package model

import "time"

type Job struct {
	ID       string    `json:"id"`
	Language string    `json:"language"`
	Content  string    `json:"content"`
	Result   string    `json:"result"`
	Status   string    `json:"status"`
	DoneTime time.Time `json:"done_time"`
}
