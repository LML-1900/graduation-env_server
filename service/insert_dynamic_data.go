package service

import (
	"env_server/data"
	"env_server/store"
	"env_server/util"
	"go.mongodb.org/mongo-driver/mongo"
)

const LEVEL = 14

type InsertDynamicDataService struct {
	mongoDB *store.MongoClient
}

func NewInsertDynamicDataService(mongo *mongo.Client) *InsertDynamicDataService {
	insertDynamicDataService := InsertDynamicDataService{}
	insertDynamicDataService.mongoDB = store.NewMongoClient(mongo)
	return &insertDynamicDataService
}

func (dynamicDataService *InsertDynamicDataService) InsertCrater(crater data.Crater) error {
	tileId := util.LonLatPositionToTileId(crater.Position, LEVEL)
	return dynamicDataService.mongoDB.InsertCrater(crater, tileId)
}
