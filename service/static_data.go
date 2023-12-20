package service

import (
	"context"
	"env_server/data"
	pb "env_server/grpc_env_service"
	"env_server/store"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"strconv"
	"strings"
	"time"
)

type StaticDataService struct {
	mongoDB *store.MongoClient
}

func NewStaticDataService(mongo *mongo.Client) *StaticDataService {
	staticDataService := StaticDataService{}
	staticDataService.mongoDB = store.NewMongoClient(mongo)
	return &staticDataService
}

func (staticDataService *StaticDataService) GetStaticData(ctx context.Context, tileId string, dataType pb.DataType) ([]byte, error) {
	switch dataType {
	case pb.DataType_DEM:
		result, err := staticDataService.mongoDB.GetStaticDemData(ctx, tileId)
		return result.Content, err
	}
	return nil, nil
}

func (staticDataService *StaticDataService) ReadDirectory(directoryPath string, fileType string) {
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		fmt.Printf("ImportDemData: fail to read directory, err: %s\n", err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			staticDataService.ReadDirectory(directoryPath+"/"+file.Name(), fileType)
		} else {
			switch fileType {
			case data.DEM_DATA_TYPE:
				err := staticDataService.ImportDemData(directoryPath + "/" + file.Name())
				if err != nil {
					fmt.Printf("ImportDemData: fail to import data, err: %s\n", err)
					return
				}
			}
		}
	}
}

func (staticDataService *StaticDataService) ImportDemData(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	tiles := strings.Split(filePath, "/")
	if len(tiles) < 3 {
		fmt.Printf("wrong filePath: %s", filePath)
		return nil
	}
	tiles = tiles[len(tiles)-3:]
	// delete file suffix
	parts := strings.Split(tiles[2], ".")
	if len(parts) >= 2 {
		tiles[2] = parts[0]
	}
	level, err := strconv.Atoi(tiles[0])
	if err != nil {
		return err
	}
	tileID := getTileId(tiles)
	beijing, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(beijing)
	demData := data.DemData{
		TileId:   tileID,
		Level:    level,
		Content:  content,
		InsertAt: now,
	}
	err = staticDataService.mongoDB.InsertDemData(demData)
	if err != nil {
		return err
	}
	return nil
}

func getTileId(tiles []string) string {
	builder := strings.Builder{}
	for i, tile := range tiles {
		builder.WriteString(tile)
		if i < len(tiles)-1 {
			builder.WriteString("-")
		}
	}
	return builder.String()
}
