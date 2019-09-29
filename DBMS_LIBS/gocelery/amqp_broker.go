package gocelery

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/streadway/amqp"
)

// AMQPExchange stores AMQP Exchange configuration
type AMQPExchange struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
}

// NewAMQPExchange creates new AMQPExchange
func NewAMQPExchange(name string) *AMQPExchange {
	return &AMQPExchange{
		Name:       name,
		Type:       "direct",
		Durable:    true,
		AutoDelete: false,
	}
}

/*func NewAMQPExchange(name, extype string, durable, autoDelete bool) *AMQPExchange {
	return &AMQPExchange{
		Name:       name,
		Type:       extype,
		Durable:    durable,
		AutoDelete: autoDelete,
	}
}*/

// AMQPQueue stores AMQP Queue configuration
type AMQPQueue struct {
	Name       string
	Durable    bool
	AutoDelete bool
}

// NewAMQPQueue creates new AMQPQueue
func NewAMQPQueue(name string) *AMQPQueue {
	return &AMQPQueue{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
	}
}

//AMQPCeleryBroker is RedisBroker for AMQP
type AMQPCeleryBroker struct {
	*amqp.Channel
	connection       *amqp.Connection
	exchange         *AMQPExchange
	queue            *AMQPQueue
	consumingChannel <-chan amqp.Delivery
	rate             int
	Confirmmsg       chan amqp.Confirmation
	hosts            []string
}

// NewAMQPConnection creates new AMQP channel
func NewAMQPConnection(host string) (*amqp.Connection, *amqp.Channel, error) {
	connection, err := amqp.Dial(host)
	if err != nil {
		return nil, nil, err
	}
	//defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, nil, err
	}
	return connection, channel, nil
}

// NewAMQPCeleryBroker creates new AMQPCeleryBroker
func NewAMQPCeleryBroker(host string) *AMQPCeleryBroker {
	conn, channel, _ := NewAMQPConnection(host)
	// ensure exchange is initialized
	broker := &AMQPCeleryBroker{
		Channel:    channel,
		connection: conn,
		exchange:   NewAMQPExchange("default"),
		queue:      NewAMQPQueue("celery"),
		rate:       4,
		hosts:      []string{host},
	}
	if err := broker.CreateExchange(); err != nil {
		panic(err)
	}
	if err := broker.CreateQueue(); err != nil {
		panic(err)
	}
	if err := broker.Qos(broker.rate, 0, false); err != nil {
		panic(err)
	}
	if err := broker.StartConsumingChannel(); err != nil {
		panic(err)
	}
	return broker
}

func NewAMQPCeleryBrokerCluster(hosts []string, ex, extype, q string) (*AMQPCeleryBroker, error) {
	var conn *amqp.Connection
	var channel *amqp.Channel
	var err error

	var hhost string
	var i int
	for i, hhost = range hosts {
		conn, channel, err = NewAMQPConnection(hhost)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if i == len(hosts) {
		return nil, err
	}

	// ensure exchange is initialized
	broker := &AMQPCeleryBroker{
		Channel:    channel,
		connection: conn,
		exchange:   NewAMQPExchange(ex),
		queue:      NewAMQPQueue(q),
		rate:       1,
		hosts:      hosts,
	}
	if err := broker.CreateExchange(); err != nil {
		return nil, err
	}
	if err := broker.CreateQueue(); err != nil {
		return nil, err
	}
	if err := broker.Qos(broker.rate, 0, false); err != nil {
		return nil, err
	}
	/*
		if err := broker.StartConsumingChannel(); err != nil {
			return nil, err
		}
	*/
	if err := broker.Confirm(false); err != nil {
		return nil, err
	}

	broker.Confirmmsg = broker.NotifyPublish(make(chan amqp.Confirmation, 1))

	return broker, nil
}

func NewAMQPCeleryBrokerMiffa(hosts []string, ex, extype, q string) (*AMQPCeleryBroker, error) {
	var conn *amqp.Connection
	var channel *amqp.Channel
	var err error

	var hhost string
	var i int
	for i, hhost = range hosts {
		conn, channel, err = NewAMQPConnection(hhost)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if i == len(hosts) {
		return nil, err
	}
	// ensure exchange is initialized
	broker := &AMQPCeleryBroker{
		Channel:    channel,
		connection: conn,
		exchange:   NewAMQPExchange(ex),
		queue:      NewAMQPQueue(q),
		rate:       1,
		hosts:      hosts,
	}
	if err := broker.CreateExchange(); err != nil {
		return nil, err
	}
	if err := broker.CreateQueue(); err != nil {
		return nil, err
	}
	if err := broker.Qos(broker.rate, 0, false); err != nil {
		return nil, err
	}
	/*
		if err := broker.StartConsumingChannel(); err != nil {
			return nil, err
		}
	*/
	if err := broker.Confirm(false); err != nil {
		return nil, err
	}

	broker.Confirmmsg = broker.NotifyPublish(make(chan amqp.Confirmation, 1))

	return broker, nil
}

func (b *AMQPCeleryBroker) Close() {
	b.Channel.Close()
	b.connection.Close()
}

func (b *AMQPCeleryBroker) Reconnect() error {
	b.Close()
	var err error
	var conn *amqp.Connection
	var channel *amqp.Channel
	var i int
	var c string
	for i, c = range b.hosts {
		conn, channel, err = NewAMQPConnection(c)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if i == len(b.hosts) {
		return err
	}
	b.Channel = channel
	b.connection = conn
	if err := b.CreateExchange(); err != nil {
		return err
	}
	if err := b.CreateQueue(); err != nil {
		return err
	}
	if err := b.Qos(b.rate, 0, false); err != nil {
		return err
	}
	/*
		if err := broker.StartConsumingChannel(); err != nil {
			return nil, err
		}
	*/
	if err := b.Confirm(false); err != nil {
		return err
	}

	b.Confirmmsg = b.NotifyPublish(make(chan amqp.Confirmation, 1))
	return nil
}

// StartConsumingChannel spawns receiving channel on AMQP queue
func (b *AMQPCeleryBroker) StartConsumingChannel() error {
	channel, err := b.Consume(b.queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	b.consumingChannel = channel
	return nil
}

// SendCeleryMessage sends CeleryMessage to broker
func (b *AMQPCeleryBroker) SendCeleryMessage(message *CeleryMessage) error {
	taskMessage := message.GetTaskMessage()
	log.Printf("sending task ID %s\n", taskMessage.ID)
	_, err := b.QueueDeclare(
		b.queue.Name,       // name
		b.queue.Durable,    // durable
		b.queue.AutoDelete, // autoDelete
		false,              // exclusive
		false,              // noWait
		nil,                // args
	)
	if err != nil {
		return err
	}
	err = b.ExchangeDeclare(
		b.exchange.Name,
		b.exchange.Type,
		b.exchange.Durable,
		b.exchange.AutoDelete,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	resBytes, err := json.Marshal(taskMessage)
	if err != nil {
		return err
	}

	publishMessage := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         resBytes,
	}

	/*return b.Publish(
		"",
		b.queue.Name,
		false,
		false,
		publishMessage,
	)*/
	err = b.Publish(
		"",
		b.queue.Name,
		false,
		false,
		publishMessage,
	)
	if err != nil {
		return err
	}

	msgack := <-b.Confirmmsg
	log.Debugf("celery mq ack:%v  seq_no:%d", msgack.Ack, msgack.DeliveryTag)
	if msgack.Ack {
		//mymsg, ierr := b.GetTaskMessage()
		//if ierr != nil {
		//	return ierr
		//}
		//log.Debugf("delivery %s ok", mymsg.ID)
		return nil
	}
	return fmt.Errorf("send to celery ack is noack")
}

// GetTaskMessage retrieves task message from AMQP queue
func (b *AMQPCeleryBroker) GetTaskMessage() (*TaskMessage, error) {
	delivery := <-b.consumingChannel
	delivery.Ack(false)
	var taskMessage TaskMessage
	if err := json.Unmarshal(delivery.Body, &taskMessage); err != nil {
		return nil, err
	}
	return &taskMessage, nil
}

// CreateExchange declares AMQP exchange with stored configuration
func (b *AMQPCeleryBroker) CreateExchange() error {
	return b.ExchangeDeclare(
		b.exchange.Name,
		b.exchange.Type,
		b.exchange.Durable,
		b.exchange.AutoDelete,
		false,
		false,
		nil,
	)
}

// CreateQueue declares AMQP Queue with stored configuration
func (b *AMQPCeleryBroker) CreateQueue() error {
	_, err := b.QueueDeclare(
		b.queue.Name,
		b.queue.Durable,
		b.queue.AutoDelete,
		false,
		false,
		nil,
	)
	return err
}
