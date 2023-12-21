package server

import (
	"context"
	"env_server/data"
	pb "env_server/grpc_env_service"
	"env_server/service"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type EnvDataServer struct {
	pb.UnimplementedEnvironmentDataServer
	dynamicDataService *service.DynamicDataService
	staticDataService  *service.StaticDataService
	mq                 *RabbitMq
}

func NewServer(mongoDB *mongo.Client, rabbitMq *RabbitMq) *EnvDataServer {
	envDataServer := EnvDataServer{
		dynamicDataService: service.NewDynamicDataService(mongoDB),
		staticDataService:  service.NewStaticDataService(mongoDB),
		mq:                 rabbitMq,
	}
	return &envDataServer
}

func (s *EnvDataServer) GetStaticData(dataRequest *pb.GetStaticDataRequest, stream pb.EnvironmentData_GetStaticDataServer) error {
	start := time.Now()
	results, err := s.staticDataService.GetStaticData(context.TODO(), dataRequest)
	if err != nil {
		log.Printf("fail to get static data, err:%v\n", err)
		return err
	}
	for _, result := range results {
		response := pb.GetStaticDataResponse{
			TileID:  result.TileId,
			Content: result.Content,
		}
		if err := stream.Send(&response); err != nil {
			log.Printf("grpc stream fail to send static data, err:%v\n", err)
			return err
		}
	}
	fmt.Printf("GetStaticData for {%v, %v}-{%v, %v} cost %v\n",
		dataRequest.Area.Bottomleft.Longitude,
		dataRequest.Area.Bottomleft.Latitude,
		dataRequest.Area.Topright.Longitude,
		dataRequest.Area.Topright.Latitude,
		time.Now().Sub(start))
	return nil
}

func (s *EnvDataServer) UpdateCrater(ctx context.Context, crater *pb.Crater) (*pb.UpdateCraterResponse, error) {
	newCrater := data.Crater{
		Position: data.LonLatPosition{
			Longitude: crater.Pos.Longitude,
			Latitude:  crater.Pos.Latitude,
		},
		Depth: crater.Depth,
		Width: crater.Width,
	}
	err := s.dynamicDataService.InsertCrater(newCrater)
	if err != nil {
		log.Printf("fail to insert crater, err:%v\n", err)
		return nil, err
	}
	go s.mq.BroadCastCraters(newCrater)
	return &pb.UpdateCraterResponse{Message: "ok"}, nil
}

func (s *EnvDataServer) GetRoutePoints(ctx context.Context, points *pb.StartStopPoints) (*pb.RoutePoints, error) {
	return nil, nil
}
