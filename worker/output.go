package worker

import (
	"context"
)

type Output struct {
	worker *WorkerQueue
	name   string
}

func (o *Output) Add(ctx context.Context, endpoint string) error {
	var err error
	o.worker, err = NewWorkerQueue(ctx, o.name, endpoint)
	return err
}

func (o *Output) Send(msg []uint8) error {
	err := o.worker.Send(msg)
	return err
}
