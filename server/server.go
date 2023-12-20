package server

import (
	"context"
	pb "env_server/grpc_env_service"
	"env_server/service"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type EnvDataServer struct {
	pb.UnimplementedEnvironmentDataServer
	dynamicDataService *service.DynamicDataService
	staticDataService  *service.StaticDataService
}

func NewServer(mongoDB *mongo.Client) *EnvDataServer {
	envDataServer := EnvDataServer{
		dynamicDataService: service.NewDynamicDataService(mongoDB),
		staticDataService:  service.NewStaticDataService(mongoDB),
	}
	return &envDataServer
}

func (s *EnvDataServer) GetStaticData(dataRequest *pb.GetStaticDataRequest, stream pb.EnvironmentData_GetStaticDataServer) error {
	start := time.Now()
	results, err := s.staticDataService.GetStaticData(context.TODO(), dataRequest)
	if err != nil {
		return err
	}
	for _, result := range results {
		response := pb.GetStaticDataResponse{
			TileID:  result.TileId,
			Content: result.Content,
		}
		if err := stream.Send(&response); err != nil {
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

func (s *EnvDataServer) UpdateCrater(ctx context.Context, crater *pb.Crater) (*pb.CraterArea, error) {
	return nil, nil
}

func (s *EnvDataServer) GetRoutePoints(ctx context.Context, points *pb.StartStopPoints) (*pb.RoutePoints, error) {
	return nil, nil
}
