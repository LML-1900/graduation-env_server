package service

import (
	"env_server/data"
	"env_server/store"
	"env_server/util"
	"go.mongodb.org/mongo-driver/mongo"
)

const LEVEL = 14

type DynamicDataService struct {
	mongoDB *store.MongoClient
}

func NewDynamicDataService(mongo *mongo.Client) *DynamicDataService {
	dynamicDataService := DynamicDataService{}
	dynamicDataService.mongoDB = store.NewMongoClient(mongo)
	return &dynamicDataService
}

func (dynamicDataService *DynamicDataService) InsertCrater(crater data.Crater) error {
	tileId := util.LonLatPositionToTileId(crater.Position, LEVEL)
	return dynamicDataService.mongoDB.InsertCrater(crater, tileId)
}

func (dynamicDataService *DynamicDataService) InsertObstacle(obstacle data.Obstacle) error {
	tileId := util.LonLatPositionToTileId(obstacle.Position, LEVEL)
	return dynamicDataService.mongoDB.InsertObstacle(obstacle, tileId)
}
