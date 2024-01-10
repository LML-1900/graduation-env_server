package server

import (
	"context"
	"env_server/data"
	pb "env_server/grpc_env_service"
	"env_server/service"
	"fmt"
	osrm "github.com/gojuno/go.osrm"
	geo "github.com/paulmach/go.geo"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"os/exec"
	"time"
)

type EnvDataServer struct {
	pb.UnimplementedEnvironmentDataServer
	dynamicDataService *service.DynamicDataService
	staticDataService  *service.StaticDataService
	osrmClient         *osrm.OSRM
}

func NewServer(mongoDB *mongo.Client) *EnvDataServer {
	envDataServer := EnvDataServer{
		dynamicDataService: service.NewDynamicDataService(mongoDB),
		staticDataService:  service.NewStaticDataService(mongoDB),
		osrmClient:         osrm.NewFromURL(viper.GetString("osrm_routing.url")),
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
		return nil, err
	}
	return &pb.UpdateCraterResponse{Message: "ok"}, nil
}

func (s *EnvDataServer) GetRoutePoints(ctx context.Context, points *pb.StartStopPoints) (*pb.RoutePoints, error) {
	resp, err := s.osrmClient.Route(ctx, osrm.RouteRequest{
		Profile: "car", // seems useless
		Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{
			{points.Start.Longitude, points.Start.Latitude},
			{points.End.Longitude, points.End.Latitude},
		}),
		Steps:       osrm.StepsTrue,
		Annotations: osrm.AnnotationsFalse,
		Overview:    osrm.OverviewFalse,
		Geometries:  osrm.GeometriesPolyline6,
	})
	if err != nil {
		log.Printf("Fail to get route points, err: %s", err)
		return nil, err
	}
	routePoints := pb.RoutePoints{
		Pos: make([]*pb.Position, 0),
	}
	if len(resp.Routes) >= 1 {
		route := resp.Routes[0]
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				for _, coordinate := range step.Geometry.PointSet {
					pos := pb.Position{
						Longitude: coordinate[0],
						Latitude:  coordinate[1],
					}
					routePoints.Pos = append(routePoints.Pos, &pos)
				}
			}
		}
	}
	return &routePoints, nil
}

func (s *EnvDataServer) GetRoutes(ctx context.Context, points *pb.StartStopPoints) {
	resp, err := s.osrmClient.Route(ctx, osrm.RouteRequest{
		Profile: "car", // seems useless
		Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{
			{points.Start.Longitude, points.Start.Latitude},
			{points.End.Longitude, points.End.Latitude},
		}),
		Steps:       osrm.StepsTrue,
		Annotations: osrm.AnnotationsFalse,
		Overview:    osrm.OverviewFalse,
		Geometries:  osrm.GeometriesPolyline6,
	})
	if err != nil {
		log.Printf("Fail to get route points, err: %s", err)
		return
	}
	//log.Printf("routes are: %+v", resp.Routes)
	if len(resp.Routes) >= 1 {
		route := resp.Routes[0]
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				for _, coordinate := range step.Geometry.PointSet {
					fmt.Printf("[%v,%v],", coordinate[0], coordinate[1])
				}
			}
		}
	}
}

func (s *EnvDataServer) AddObstacle(position data.LonLatPosition) {
	// get nearest nodes IDs
	resp, err := s.osrmClient.Nearest(context.TODO(), osrm.NearestRequest{
		Profile: "car",
		Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{
			{position.Longitude, position.Latitude},
		}),
		Number: 5,
	})
	if err != nil {
		log.Printf("Fail to get nearest points, err: %s", err)
		return
	}
	if resp.ResponseStatus.Code != "Ok" {
		log.Printf("Fail to get nearest points, err: %s", resp.ResponseStatus.Message)
		return
	}
	// get different ids(except 0)
	obstacleId := make([]uint64, 0)
	for _, wayPoint := range resp.Waypoints {
		for _, node := range wayPoint.Nodes {
			if node != 0 {
				if len(obstacleId) == 0 {
					obstacleId = append(obstacleId, node)
				} else {
					if node != obstacleId[0] {
						obstacleId = append(obstacleId, node)
						break
					}
				}
			}
		}
		if len(obstacleId) >= 2 {
			break
		}
	}
	// write nodes into csv file
	if len(obstacleId) >= 2 {
		file_name := viper.GetString("osrm_routing.csv_file_name")
		file, err := os.OpenFile(file_name, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			file, err = os.Create(file_name) // Create if it doesn't exist
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
		}
		defer file.Close()
		// write in two directions
		fmt.Fprintf(file, "%d,%d,%d\n", obstacleId[0], obstacleId[1], 0)
		fmt.Fprintf(file, "%d,%d,%d\n", obstacleId[1], obstacleId[0], 0)
	}

}

func OSRMReCustomize() {
	// run re-customize command
	cmd := exec.Command("osrm-customize", viper.GetString("osrm_routing.map_name"), "--segment-speed-file", viper.GetString("osrm_routing.csv_file_name"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing command:", err)
		fmt.Println(string(output))
		return
	}
}
