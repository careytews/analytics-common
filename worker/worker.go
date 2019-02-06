package worker

import (
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof" // 'side-effects' import for registering http handlers
	"os"
	"strings"
	"time"

	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trustnetworks/analytics-common/amqp"
	"github.com/trustnetworks/analytics-common/utils"
)

type EventHandler interface {
	Handle(message []uint8)
}

type Worker struct {
	ctrl        *os.File
	out         *OutputSet
	notifyClose chan struct{}
}

func (w *Worker) Initialise(ctx context.Context, outputs []string) error {
	var err error
	w.notifyClose = make(chan struct{})
	w.out, err = w.ParseOutputs(ctx, outputs)
	return err
}

var (
	Pgm = "undefined"
)

func (w *Worker) ParseOutputs(ctx context.Context, a []string) (*OutputSet, error) {

	outs := NewOutputSet()

	for _, elt := range a {

		toks := strings.SplitN(elt, ":", 2)

		name := toks[0]
		endpoint := toks[1]

		err := outs.Add(ctx, name, endpoint)
		if err != nil {
			return nil, err
		}

	}

	return outs, nil

}

func (w *Worker) Send(name string, msg []uint8) {
	w.out.outputs[name].Send(msg)
}

type QueueWorker struct {
	Worker
	eventsReceivedCounter *prometheus.CounterVec
	msgReceivedLatency    *prometheus.SummaryVec
	recvLabels            prometheus.Labels

	queue    string
	broker   string
	exchange string

	consumer *amqp.AMQPConsumer
}

type Handler interface {
	Handle(message []uint8, w *Worker) error
}

func (w *QueueWorker) Initialise(ctx context.Context, input string, outputs []string, pgm string) error {

	err := w.Worker.Initialise(ctx, outputs)
	Pgm = pgm
	if err != nil {
		return err
	}

	w.broker = utils.Getenv("AMQP_BROKER", "amqp://guest:guest@localhost:5672/")
	w.exchange = input
	w.queue = fmt.Sprintf("analytics-%s", Pgm)

	// Config Prom Stats
	w.eventsReceivedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_received",
			Help: "number of events received",
		},
		[]string{"analytic", "exchange", "type", "queue"},
	)

	w.msgReceivedLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "message_latency",
			Help: "Latency of messaged received",
		},
		[]string{"analytic", "exchange", "type", "queue"},
	)

	w.recvLabels = prometheus.Labels{"analytic": Pgm, "exchange": w.exchange, "queue": w.queue, "type": "amqp"}
	prometheus.MustRegister(w.eventsReceivedCounter)
	prometheus.MustRegister(w.msgReceivedLatency)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)

	return nil
}

func (w *QueueWorker) qReader(ctx context.Context, ch chan []uint8) {

	consumer := amqp.NewShardedConsumer(
		ctx,
		w.queue,
		w.exchange,
		w.broker,
		1000, // This can be high - as long as the broker/analytic has memory for it
		true, // queue is persistent
	)

	handler := func(msg []byte, ts time.Time) {
		ch <- msg

		// Record stats
		go func() {
			if !ts.IsZero() {
				elapsed := time.Now().Sub(ts)
				w.eventsReceivedCounter.With(w.recvLabels).Inc()
				w.msgReceivedLatency.With(w.recvLabels).Observe(float64(elapsed / time.Millisecond)) // this will be in ms
			}
		}()
	}

	err := consumer.Consume(handler)
	if err != nil {
		utils.Log("error: Error in reading from queue: %s", err.Error())
		close(w.notifyClose)
	}
}

func (w *QueueWorker) Run(ctx context.Context, h Handler) error {

	ch := make(chan []uint8, 100)

	go w.qReader(ctx, ch)

	for {
		select {
		case val := <-ch:
			h.Handle(val, &(w.Worker))

		case <-w.notifyClose: // The subscriber has died?
			return errors.New("qReader quit unexpectedly")

		case <-ctx.Done():
			return nil
		}
	}
}
