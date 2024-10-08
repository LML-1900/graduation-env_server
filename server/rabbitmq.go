package server

import (
	"context"
	"encoding/json"
	"env_server/data"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

type RabbitMq struct {
	Conn         *amqp.Connection
	ProducerCh   *amqp.Channel
	ConsumerCh   *amqp.Channel
	qname        string
	exchangeName string
}

func CreateMessageQueue(exchangeName string, url string) (*RabbitMq, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("fail to connect to RabbitMQ, err:%v", err)
		return nil, err
	}

	// for producer
	producerCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("fail to open a producer channel, err:%v", err)
		return nil, err
	}

	// for consumer
	consumerCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("fail to open a consumer channel, err:%v", err)
		return nil, err
	}

	err = producerCh.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange, err:%v", err)
		return nil, err
	}

	// to consume obstacles
	q, err := consumerCh.QueueDeclare(
		"",    // empty name means RabbitMQ will generate a unique queue name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
		return nil, err
	}

	err = consumerCh.QueueBind(
		q.Name,                         // queue name
		data.OSRM_OBSTACLE_ROUTING_KEY, // routing key for obstacles
		exchangeName,                   // exchange name (topic exchange)
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind obstacle routing key: %v", err)
		return nil, err
	}

	mq := RabbitMq{
		Conn:         conn,
		ProducerCh:   producerCh,
		ConsumerCh:   consumerCh,
		exchangeName: exchangeName,
		qname:        q.Name,
	}
	return &mq, nil
}

func (mq *RabbitMq) BroadCastCraters(crater data.Crater) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	craterBytes, err := json.Marshal(crater)
	if err != nil {
		log.Printf("Failed to marshal crater, err:%v\n", err)
		return
	}

	err = mq.ProducerCh.PublishWithContext(ctx,
		mq.exchangeName,         // exchange
		data.CRATER_ROUTING_KEY, // routing key
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        craterBytes,
		})
	if err != nil {
		log.Printf("Failed to publish a message, err:%v\n", err)
		return
	}

	log.Printf(" [x] Sent Crater at [lon:%v lat:%v]", crater.Position.Longitude, crater.Position.Latitude)
}

func (mq *RabbitMq) BroadCastObstacles(obstacle data.OSRM_Obstacle) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	obstacleBytes, err := json.Marshal(obstacle)
	if err != nil {
		log.Printf("Failed to marshal obstacle, err:%v\n", err)
		return
	}

	err = mq.ProducerCh.PublishWithContext(ctx,
		mq.exchangeName,                // exchange
		data.OSRM_OBSTACLE_ROUTING_KEY, // routing key
		false,                          // mandatory
		false,                          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        obstacleBytes,
		})
	if err != nil {
		log.Printf("Failed to publish a message, err:%v\n", err)
		return
	}

	log.Printf(" [x] Sent obstacle: [start:%d stop:%d]", obstacle.StartID, obstacle.StopID)
}

func OSRMBatchUpdate(obstacles []data.OSRM_Obstacle) {
	file_name := viper.GetString("osrm_routing.csv_file_name")
	file, err := os.OpenFile(file_name, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		file, err = os.Create(file_name) // Create if it doesn't exist
		if err != nil {
			log.Println("Error creating file:", err)
			return
		}
	}
	defer file.Close()
	for _, obstacle := range obstacles {
		// write in two directions
		fmt.Fprintf(file, "%d,%d,%d\n", obstacle.StartID, obstacle.StopID, 0)
		fmt.Fprintf(file, "%d,%d,%d\n", obstacle.StopID, obstacle.StartID, 0)
	}

	customizeCmd := exec.Command("/home/lml/graduation/osrm_new/osrm-backend/build/osrm-customize", viper.GetString("osrm_routing.map_name"), "--segment-speed-file", viper.GetString("osrm_routing.csv_file_name"), "--incremental=true")
	output, err := customizeCmd.CombinedOutput()
	if err != nil {
		log.Println("Error executing command:", err)
		log.Println(string(output))
		return
	}
	// reload map
	reloadCmd := exec.Command("/home/lml/graduation/osrm_new/osrm-backend/build/osrm-datastore", viper.GetString("osrm_routing.map_name"))
	output, err = reloadCmd.CombinedOutput()
	if err != nil {
		log.Println("Error executing command:", err)
		log.Println(string(output))
		return
	}
}

func (mq *RabbitMq) ConsumeObstacles() {
	msgs, err := mq.ConsumerCh.Consume(
		mq.qname, // queue name
		"",       // consumer tag (empty for auto-generated)
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
		return
	}
	log.Println("start to consume obstacles")
	defer func() {
		log.Println("end to consume obstacles")
	}()
	var obstacles []data.OSRM_Obstacle
	timer := time.NewTimer(data.OSRM_UPDATE_INTERVAL)
	defer timer.Stop()

	mu := &sync.Mutex{}

	go func() {
		for {
			<-timer.C
			mu.Lock()
			if len(obstacles) > 0 {
				log.Printf("reach time limit %d, process %d nodes\n", data.OSRM_UPDATE_INTERVAL, len(obstacles))
				OSRMBatchUpdate(obstacles)
				obstacles = nil // Reset after processing
			}
			mu.Unlock()
			timer.Reset(data.OSRM_UPDATE_INTERVAL * time.Second)
		}
	}()
	// log.Println("consume msg in a loop")
	for msg := range msgs {
		var obstacle data.OSRM_Obstacle
		if err := json.Unmarshal(msg.Body, &obstacle); err != nil {
			log.Printf("Error unmarshaling obstacle data: %v", err)
			continue
		}
		log.Printf("received obstacle: (%d-%d)\n", obstacle.StartID, obstacle.StopID)
		mu.Lock()
		obstacles = append(obstacles, obstacle)
		// If we reach the batch size, process immediately
		if len(obstacles) >= data.OSRM_UPDATE_BATCH {
			log.Printf("reach batch limit %d, process %d nodes\n", data.OSRM_UPDATE_BATCH, len(obstacles))
			OSRMBatchUpdate(obstacles)
			obstacles = nil // Reset after processing
			timer.Stop()
			timer.Reset(data.OSRM_UPDATE_INTERVAL) // Reset timer after batch processing
		}
		mu.Unlock()
	}
}
