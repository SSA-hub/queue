package queue_http

import (
	"net/http"
	"queue/internal/queue"
)

func MapRoutes(h queue.Handler) {
	http.HandleFunc("/queue", h.Process("/queue/"))
}
