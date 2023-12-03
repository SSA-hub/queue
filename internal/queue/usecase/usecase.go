package queue_usecase

import (
	"context"
	"errors"
	"queue/internal/queue"
	queue_model "queue/internal/queue/model"
	"sync"
	"sync/atomic"
	"time"
)

const (
	_defaultQueueLen = 1024
)

type uc struct {
	mem            sync.Map
	queueLen       int
	queuesCount    atomic.Uint64
	maxQueuesCount int64
}

func New(queueLen int, maxQueueCount int64) queue.UC {
	if queueLen <= 0 {
		queueLen = _defaultQueueLen
	}

	return &uc{
		mem:            sync.Map{},
		queueLen:       queueLen,
		queuesCount:    atomic.Uint64{},
		maxQueuesCount: maxQueueCount,
	}
}

func (u *uc) Send(queueName string, input queue_model.Message) error {
	if input.Message == nil {
		return errors.New("empty input")
	}

	var err error
	q := u.getQueue(queueName)
	if q == nil {
		q, err = u.createQueue(queueName)
		if err != nil {
			return err
		}
	}

	if len(q) >= u.queueLen {
		return errors.New("messages limit has been reached in this queue")
	}

	q <- *input.Message
	return nil
}

func (u *uc) Get(queueName string, timeout *uint64) (*string, error) {
	ctx := context.Background()
	var cancel context.CancelFunc
	if timeout != nil {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
		defer cancel()
	}

	q := u.getQueue(queueName)
	if q == nil {
		return nil, errors.New("not found")
	}

	for {
		select {
		case msg, ok := <-q:
			if !ok {
				return nil, errors.New("queue is closed")
			}

			return &msg, nil
		case <-ctx.Done():
			return nil, errors.New("timeout")
		default:
		}
	}
}

func (u *uc) getQueue(name string) chan string {
	val, ok := u.mem.Load(name)
	if !ok {
		return nil
	}

	q, ok := val.(chan string)
	if !ok {
		return nil
	}

	return q
}

func (u *uc) createQueue(name string) (chan string, error) {
	if u.maxQueuesCount > 0 && u.queuesCount.Load() >= uint64(u.maxQueuesCount) {
		return nil, errors.New("queues limit has been reached")
	}

	val := make(chan string, u.queueLen)
	u.mem.Store(name, val)
	u.queuesCount.Add(1)
	return val, nil
}
