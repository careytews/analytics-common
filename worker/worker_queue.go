package worker

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/trustnetworks/analytics-common/amqp"
	"github.com/trustnetworks/analytics-common/utils"
	"context"
)

type WorkerQueue struct {
	endpoint      string
	internalQueue chan []uint8

	notifyClose chan struct{}

	exchange string
	broker   string

	eventsSentCounter *prometheus.CounterVec
	sentLabels        prometheus.Labels
}

func (w *WorkerQueue) qWriter(ctx context.Context) {
	publisher := amqp.NewPublisher(ctx, w.exchange, w.broker)
	err := publisher.Publish(w.internalQueue)
	if err != nil {
		utils.Log("error: Failed to write to queue with error: %s", err.Error())
		close(w.notifyClose)
	}
}

func (w *WorkerQueue) Send(msg []uint8) error {
	w.eventsSentCounter.With(w.sentLabels).Inc()
	select {
	case <-w.notifyClose:
		return errors.New("qWriter has stopped unexpectedly")
	default:
		w.internalQueue <- msg
	}
	return nil
}

// Name is the name of the output type
// Endpoint is the routing key
func NewWorkerQueue(ctx context.Context, name string, endpoint string) (w *WorkerQueue, err error) {

	w = new(WorkerQueue)
	w.internalQueue = make(chan []uint8, 100)
	w.endpoint = endpoint
	w.notifyClose = make(chan struct{})

	w.broker = utils.Getenv("AMQP_BROKER", "amqp://guest:guest@localhost:5672/")
	w.exchange = endpoint

	// Config Prom Stats
	w.eventsSentCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_events_sent", name),
			Help: "number of events sent",
		},
		[]string{"analytic", "exchange", "type"},
	)
	prometheus.MustRegister(w.eventsSentCounter)
	w.sentLabels = prometheus.Labels{"analytic": Pgm, "exchange": w.exchange, "type": "amqp"}

	go w.qWriter(ctx)

	return
}
