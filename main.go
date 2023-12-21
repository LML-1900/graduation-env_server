package main

import (
	"context"
	pb "env_server/grpc_env_service"
	"env_server/server"
	"flag"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	// set viper
	viper.SetConfigFile("./config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// set mongodb
	uri := viper.GetString("mongodb.uri")
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("fail to new a Mongodb client!")
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// set rabbitmq
	mq, err := server.CreateMessageQueue("dynamic_data_topic")
	if err != nil {
		log.Fatalf("failed to create message queue: %v", err)
	}
	defer func(Conn *amqp.Connection) {
		err := Conn.Close()
		if err != nil {

		}
	}(mq.Conn)
	defer func(Ch *amqp.Channel) {
		err := Ch.Close()
		if err != nil {

		}
	}(mq.Ch)

	// new an inset static data service
	//staticDataService := service.NewAddStaticDataService(mongoClient)
	//directoryPath := "/home/lml/env_server/11-736-158demData/11_736_158_WGS84_terrain"
	//staticDataService.ReadDirectory(directoryPath, data.DEM_DATA_TYPE)

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	envDataServer := server.NewServer(mongoClient, mq)
	pb.RegisterEnvironmentDataServer(s, envDataServer)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
