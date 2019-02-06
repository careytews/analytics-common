package amqp

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"os"
	"strconv"
	"time"

	"github.com/trustnetworks/analytics-common/utils"
)

type AMQPSession struct {
	*amqp.Channel
	*amqp.Connection
}

type AMQPClient struct {
	Broker   string
	Exchange string
	sessions chan chan AMQPSession
}

type AMQPPublisher struct {
	AMQPClient
}

type AMQPConsumer struct {
	AMQPClient
	QueueName       string
	DLQName         string
	DLExchange      string
	ShardedExchange string
	Prefetch        int
	Persistent      bool
	AckThreshold    int

	ctx context.Context
}

func (s AMQPSession) Close() error {
	if s.Connection == nil {
		return nil
	}
	return s.Connection.Close()
}

// redial continually connects to the URL, exiting the program when no longer possible
func redial(ctx context.Context, url string, exchange string) chan chan AMQPSession {
	sessions := make(chan chan AMQPSession)

	go func() {
		sess := make(chan AMQPSession)
		defer close(sess)

		for {
			select {
			case sessions <- sess:
			case <-ctx.Done():
				utils.Log("amqp: shutting down session factory")
				return
			}

			conn, err := amqp.Dial(url)
			if err != nil {
				utils.Log("amqp: cannot (re)dial: %v: %q", err, url)
				return
			}

			ch, err := conn.Channel()
			if err != nil {
				utils.Log("amqp: cannot create channel: %v", err)
				return
			}

			if err := ch.ExchangeDeclare(
				exchange, // name
				"fanout", // type
				true,     // durable
				false,    // auto-deleted
				false,    // internal
				false,    // no-wait
				nil,      // arguments
			); err != nil {
				utils.Log("amqp: cannot declare fanout exchange: %v", err)
				return
			}

			select {
			case sess <- AMQPSession{ch, conn}:
			case <-ctx.Done():
				utils.Log("amqp: shutting down new session")
				return
			}
		}
	}()

	return sessions
}

// Return a new object that can be used to publish to a fanout exchange.
func NewPublisher(ctx context.Context, exchange string, broker string) *AMQPPublisher {
	p := new(AMQPPublisher)
	p.Broker = broker
	p.Exchange = exchange

	p.sessions = redial(ctx, p.Broker, p.Exchange)

	return p
}

// Declare and bind a consumer queue onto the channel
// so that it can be used to consumer messages from
// We use Queue like a consumer group - named after the worker
// and we use an Exchange like a destination - named after the type of
// event we would expect on the exchange:
//
//            ) -> |-------------| -> [worker.elasticsearch] -> (N * analytics-elasticsearch)
// PUBLISHERS > -> | event.trust |
//            ) -> |-------------| -> [worker.risk-graph] -> (N * analytics-risk-graph)
//
// name = QueueName
// excahnge = Exchange Name
// broker = Broker URL
func NewConsumer(ctx context.Context, name string, exchange string, broker string, prefetch int, persistent bool) *AMQPConsumer {

	c := new(AMQPConsumer)
	c.Broker = broker
	c.QueueName = name
	c.Exchange = exchange
	c.Prefetch = prefetch
	c.sessions = redial(ctx, c.Broker, c.Exchange)
	c.Persistent = persistent
	c.AckThreshold = 100
	c.ctx = ctx

	return c
}

// Declare and bind a consumer queue onto the channel, using an intermediary sharded
// exchange so that it can be used to consumer messages from
// We use Queue like a consumer group - named after the worker
// and we use an Exchange like a destination - named after the type of
// event we would expect on the exchange:
//																	  |-------------------------|
//                                    | analytics-elasticsearch | -> [N * analytics-elasticsearch ] -> (N * analytics-elasticsearch)
//            ) -> |-------------| -> |-------------------------|
// PUBLISHERS > -> | trust_event | ->
//            ) -> |-------------| -> |----------------------|
//																	  | analytics-risk-graph | -> [N * analytics-risk-graph] -> (N * analytics-risk-graph)
//                                    |----------------------|
// name = Sharded exchange name
// exchange = Exchange Name
// broker = Broker URL
func NewShardedConsumer(ctx context.Context, name string, exchange string, broker string, prefetch int, persistent bool) *AMQPConsumer {

	hostname, err := os.Hostname()
	if err != nil {
		// been unable to get hostname. Use random number instead
		hostname = fmt.Sprintf("%s-%s", name, uuid.New())
	}
	c := NewConsumer(ctx, hostname, exchange, broker, prefetch, persistent)
	c.ShardedExchange = name
	c.DLExchange = fmt.Sprintf("%s-dlx", name)
	c.DLQName = fmt.Sprintf("%s-dlq", name)

	return c
}

// publish publishes messages to a reconnecting session to a fanout exchange.
// It receives from the application specific source of messages.
func (p *AMQPPublisher) Publish(messages <-chan []byte) error {
	for session := range p.sessions {
		var (
			running bool
			reading = messages
			pending = make(chan []byte, 1)
			confirm = make(chan amqp.Confirmation, 1)
		)

		pub, ok := <-session
		if !ok {
			break
		}
		// publisher confirms for this channel/connection
		if err := pub.Confirm(false); err != nil {
			utils.Log("amqp: publisher confirms not supported")
			close(confirm) // confirms not supported, simulate by always nacking
		} else {
			pub.NotifyPublish(confirm)
		}

		utils.Log("amqp: publishing events to: %s", p.Exchange)

	Pub:
		for {
			var body []byte
			select {
			case confirmed, ok := <-confirm:
				if !ok {
					break Pub
				}
				if !confirmed.Ack {
					utils.Log("amqp: nack message %d, body: %q", confirmed.DeliveryTag, string(body))
				}
				reading = messages

			case body = <-pending:
				err := pub.Publish(p.Exchange, "", false, false, amqp.Publishing{
					ContentType: "text/plain",
					Body:        body,
					Headers: amqp.Table{
						"timestamp_in_ns": strconv.FormatInt(time.Now().UnixNano(), 10),
					},
				})
				// Retry failed delivery on the next session
				if err != nil {
					pending <- body
					pub.Close()
					break Pub
				}

			case body, running = <-reading:
				// all messages consumed
				if !running {
					return nil
				}
				// work on pending delivery until ack'd
				pending <- body
				reading = nil
			}
		}
	}
	return errors.New("No more sessions left to try")
}

// Async acknolwedgement goroutine
func acker(sub AMQPSession, ackQueue chan uint64, x chan struct{}) {

	// Continually read queue
	for {
		select {

		// Exit case.
		case <-x:
			utils.Log("amqp: closing down acker")
			return

		// Incoming message case
		case tag := <-ackQueue:
			// Acknlowedge
			sub.Ack(tag, true)
		}
	}
}

func (c *AMQPConsumer) SetAckThreshold(threshold int) {
	c.AckThreshold = threshold
}

func (c *AMQPConsumer) declareAndBindQ(sub AMQPSession) error {
	qArgs := make(amqp.Table)
	qExchange := c.Exchange

	if c.ShardedExchange != "" {
		exArgs := amqp.Table{
			"alternate-exchange": c.DLExchange,
		}

		qExchange = c.ShardedExchange
		if err := sub.ExchangeDeclare(
			c.ShardedExchange, // name
			"x-random",        // type
			true,              // durable
			false,             // auto-deleted
			true,              // internal
			true,              // no-wait
			exArgs,            // arguments
		); err != nil {
			return fmt.Errorf("Cannot declare the analytic exchange %q: %v", c.ShardedExchange, err)
		}

		if err := sub.ExchangeDeclare(
			c.DLExchange, // name
			"direct",     // type
			true,         // durable
			false,        // auto-deleted
			false,        // internal
			true,         // no-wait
			nil,          // arguments
		); err != nil {
			return fmt.Errorf("Cannot declare the analytic dlexchange %q: %v", c.DLExchange, err)
		}

		if err := sub.ExchangeBind(
			c.ShardedExchange, // destination
			"",                // key
			c.Exchange,        // Source
			true,              // no-wait
			nil,               // args
		); err != nil {
			return fmt.Errorf("Cannot bind analytic exchange %q to event type %e: %v", c.ShardedExchange, c.Exchange, err)
		}

		qArgs["x-dead-letter-exchange"] = c.DLExchange
		qArgs["x-expires"] = int32(30000)
		qArgs["x-message-ttl"] = int32(10000)
	}

	if _, err := sub.QueueDeclare(
		c.QueueName,   // name
		c.Persistent,  // durable
		!c.Persistent, // delete when unused
		!c.Persistent, // exclusive
		true,          // no-wait
		qArgs,         // arguments
	); err != nil {
		return fmt.Errorf("Cannot declare replica queue %q: %v", c.QueueName, err)
	}

	if err := sub.QueueBind(
		c.QueueName, // queue name
		"",          // routing key
		qExchange,   // exchange
		true,        // no-wait
		nil,         // args
	); err != nil {
		return fmt.Errorf("Cannot consume without a binding to exchange: %q, %v", qExchange, err)
	}

	if c.DLExchange != "" {
		if _, err := sub.QueueDeclare(
			c.DLQName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			true,      // no-wait
			nil,       // arguments
		); err != nil {
			return fmt.Errorf("Cannot declare dead letter queue %q: %v", c.DLQName, err)
		}

		if err := sub.QueueBind(
			c.DLQName,    // queue name
			"",           // routing key
			c.DLExchange, // exchange
			true,         // no-wait
			nil,          // args
		); err != nil {
			return fmt.Errorf("Cannot consume without a binding to exchange: %q, %v", c.DLExchange, err)
		}
	}

	utils.Log("amqp: Queue %s has been declared and bound on exchange %s, feeding event type %s", c.QueueName, qExchange, c.Exchange)
	return nil
}

func (c *AMQPConsumer) Consume(handle func([]byte, time.Time)) error {

	exitQ := make(chan struct{})
	exitDLQ := make(chan struct{})

	go consumeQ(c, c.QueueName, c.Exchange, true, handle, exitQ)
	if c.DLExchange != "" {
		go consumeQ(c, c.DLQName, c.DLExchange, false, handle, exitDLQ)
	}

	select {
	case <-exitQ:
		return fmt.Errorf("Consuming form the Queue unexepctedly exited")
	case <-exitDLQ:
		return fmt.Errorf("Consuming form the DLQ unexepctedly exited")
	}
}

// subscribe consumes deliveries from an exclusive queue from a fanout exchange and sends to the application specific messages chan.
func consumeQ(c *AMQPConsumer, queue string, exchange string, exclusive bool, handle func([]byte, time.Time), exit chan struct{}) {
	declared := false
	for session := range c.sessions {

		utils.Log("amqp: Attempting to join a session to consume")
		sub, ok := <-session
		if !ok {
			break // we are out of sessions
		}

		// Create ack queue and closedown queue
		ackQueue := make(chan uint64, 5000)
		x := make(chan struct{})

		// control the interval of acks
		tickChan := time.NewTicker(time.Second * 1).C

		// Launch ack goroutine
		go acker(sub, ackQueue, x)
		notify := sub.Channel.NotifyClose(make(chan *amqp.Error))

		if !declared {
			if err := c.declareAndBindQ(sub); err != nil {
				utils.Log("Failed to bind queues: %v", queue, err)
				break
			}
			declared = true
		}

		if err := sub.Qos(
			c.Prefetch, // prefetch count
			0,          // prefetch size
			false,      // global
		); err != nil {
			utils.Log("Failed to set prefetch count on queue: %q, %v", queue, err)
			break
		}

		deliveries, err := sub.Consume(
			queue,     // queue
			"",        // consumer
			false,     // auto-ack
			exclusive, // exclusive
			false,     // no-local
			true,      // no-wait
			nil,       // args
		)
		if err != nil {
			utils.Log("Cannot consume from: %q, %v", queue, err)
			break
		}

		utils.Log("amqp: subscribed to events from: %s", exchange)
		count := 0

	Sub:
		for { //receive loop
			select { //check connection
			case <-c.ctx.Done():
				if c.ShardedExchange != "" {
					//we need to unbind the queue
					if err := sub.QueueUnbind(
						queue,             // queue name
						"",                // routing key
						c.ShardedExchange, // exchange
						nil,
					); err != nil {
						utils.Log("Failure unbinding the queue from exchange: %q, %e - %v", queue, c.ShardedExchange, err)
						break Sub
					}
				}
				break Sub
			case err = <-notify:
				break Sub //reconnect

			case msg, ok := <-deliveries:
				if !ok {
					utils.Log("amqp: consumer on %s has ended, attempting to reconnect", queue)
					deliveries = nil
					break Sub
				}

				nsTime := time.Time{}
				if s, ok := msg.Headers["timestamp_in_ns"].(string); ok {
					secs, _ := strconv.Atoi(s[0 : len(s)-9])
					nSecs, _ := strconv.Atoi(s[len(s)-9:])
					nsTime = time.Unix(int64(secs), int64(nSecs))
				}
				handle(msg.Body, nsTime) // Handle the message (normally place on channel and metricate)
				count += 1

				select {
				case <-tickChan:
					ackQueue <- msg.DeliveryTag
					break
				default:
					if (count % c.AckThreshold) == 0 {
						ackQueue <- msg.DeliveryTag
					}
					break
				}
			}
		}
		close(x)
		sub.Close()
		time.Sleep(1 * time.Second)
	}

	utils.Log("amqp: No more sessions left to try for q: %s", queue)
	close(exit)
}
