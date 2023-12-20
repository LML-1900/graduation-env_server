package server

import (
	"context"
	pb "env_server/grpc_env_service"
	"env_server/service"
	"env_server/util"
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
	tileIds := util.LonLatAreaToTileIds(dataRequest.Area, int(dataRequest.Level))
	for _, tileId := range tileIds {
		content, err := s.staticDataService.GetStaticData(context.TODO(), tileId, dataRequest.DataType)
		if err != nil {
			return err
		}
		response := pb.GetStaticDataResponse{
			TileID:  tileId,
			Content: content,
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
