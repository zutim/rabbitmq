package main

import (
	"bytes"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

const (
	//AMQP URI
	uri = "amqp://user:password@115.28.129.153:8082/"
	//Durable AMQP exchange nam
	exchangeName = ""
	//Durable AMQP queue name
	queueName = "test-queues-acknowledgments"
)

//如果存在错误，则输出
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func main() {
	//调用消息接收者
	consumer(uri, exchangeName, queueName)
}

//接收者方法
//@amqpURI, amqp的地址
//@exchange, exchange的名称
//@queue, queue的名称
func consumer(amqpURI string, exchange string, queue string) {
	//建立连接
	log.Printf("dialing %q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connection.Close()

	//创建一个Channel
	log.Printf("got Connection, getting Channel")
	channel, err := connection.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	log.Printf("got queue, declaring %q", queue)

	//创建一个queue
	q, err := channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	log.Printf("Queue bound to Exchange, starting Consume")
	//订阅消息
	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	//创建一个channel
	forever := make(chan bool)

	//调用gorountine
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dot_count := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dot_count)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	//没有写入数据，一直等待读，阻塞当前线程，目的是让线程不退出
	<-forever
}
