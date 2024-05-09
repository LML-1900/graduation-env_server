package main

import (
	"context"
	pb "env_server/grpc_env_service"
	"env_server/server"
	"flag"
	"fmt"
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
		fmt.Printf("fail to new a Mongodb client!\n")
		panic(err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// new an inset static data service
	//staticDataService := service.NewAddStaticDataService(mongoClient)
	//directoryPath := "/home/lml/env_server/11-736-158demData/11_736_158_WGS84_terrain"
	//staticDataService.ReadDirectory(directoryPath, data.DEM_DATA_TYPE)

	// new an osrm test service
	//osrmserver := server.NewServer(mongoClient)

	// test routing function
	//points := pb.StartStopPoints{
	//	Start: &pb.Position{
	//		Longitude: 113.5439372,
	//		Latitude:  22.2180642,
	//	},
	//	End: &pb.Position{
	//		Longitude: 113.5425177,
	//		Latitude:  22.2252363,
	//	},
	//}
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Adjust timeout as needed
	//defer cancel()
	//osrmserver.GetRoutes(ctx, &points)

	// test add obstacle function
	//point := data.LonLatPosition{
	//	Longitude: 113.542676,
	//	Latitude:  22.217951,
	//}
	//osrmserver.AddObstacle(point)
	//server.OSRMReCustomize()

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	envDataServer := server.NewServer(mongoClient)
	pb.RegisterEnvironmentDataServer(s, envDataServer)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
