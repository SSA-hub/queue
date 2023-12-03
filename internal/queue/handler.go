package queue

import "net/http"

type Handler interface {
	Process(prefix string) http.HandlerFunc
}
