package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/imdigitalashish/QueueSystemsInGolang/internal/api"
	"github.com/imdigitalashish/QueueSystemsInGolang/internal/queue"
)

func main() {
	r := gin.Default()
	q := queue.NewQueue(5)

	api.SetupRoutes(r, q)

	log.Fatal(r.Run(":8080"))
}
