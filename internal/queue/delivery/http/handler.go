package queue_http

import (
	"encoding/json"
	"io"
	"net/http"
	"queue/internal/queue"
	queue_model "queue/internal/queue/model"
	"strconv"
	"strings"
)

type httpHandler struct {
	uc queue.UC
}

func NewHttpHandler(uc queue.UC) queue.Handler {
	return &httpHandler{
		uc: uc,
	}
}

func (h *httpHandler) Process(prefix string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "GET":
			h.get(prefix, writer, request)
		case "POST":
			h.send(prefix, writer, request)
		}
	}
}

func (h *httpHandler) send(prefix string, writer http.ResponseWriter, request *http.Request) {
	queueName := strings.TrimPrefix(request.URL.Path, prefix)
	if queueName == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var body queue_model.Message
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.uc.Send(queueName, body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	return

}

func (h *httpHandler) get(prefix string, writer http.ResponseWriter, request *http.Request) {
	queueName := strings.TrimPrefix(request.URL.Path, prefix)
	if queueName == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var timeout *uint64
	queryParamsTimeout, err := strconv.ParseUint(request.URL.Query().Get("timeout"), 10, 64)
	if err == nil {
		timeout = &queryParamsTimeout
	}

	result, err := h.uc.Get(queueName, timeout)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResult, err := json.Marshal(queue_model.Message{
		Message: result,
	})
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = writer.Write(jsonResult)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
