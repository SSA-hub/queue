package queue

import queue_model "queue/internal/queue/model"

type UC interface {
	Send(queueName string, input queue_model.Message) error
	Get(queueName string, timeout *uint64) (*string, error)
}
