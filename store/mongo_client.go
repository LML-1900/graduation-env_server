package store

import (
	"context"
	"env_server/data"
	"fmt"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	client        *mongo.Client
	staticDemColl *mongo.Collection
	dynamicColl   *mongo.Collection
	craterColl    *mongo.Collection
}

func NewMongoClient(mongo *mongo.Client) *MongoClient {
	mongoClinet := MongoClient{}
	mongoClinet.client = mongo
	databaseName := viper.GetString("mongodb.database")
	mongoClinet.staticDemColl = mongo.Database(databaseName).Collection(viper.GetString("mongodb.staticDemColl"))
	mongoClinet.dynamicColl = mongo.Database(databaseName).Collection(viper.GetString("mongodb.dynamicColl"))
	mongoClinet.craterColl = mongo.Database(databaseName).Collection(viper.GetString("mongodb.craterColl"))
	return &mongoClinet
}

func (mongoClient *MongoClient) GetStaticDemData(ctx context.Context, tileIds []string) (results []data.DemData, err error) {
	filter := bson.D{{"tile_id", bson.D{{"$in", tileIds}}}}
	cursor, err := mongoClient.staticDemColl.Find(ctx, filter)
	err = cursor.All(ctx, &results)
	return results, err
}

func (mongoClient *MongoClient) InsertDemData(demData data.DemData) error {
	result, err := mongoClient.staticDemColl.InsertOne(context.TODO(), demData)
	if err != nil {
		return err
	}
	fmt.Println(result.InsertedID)
	return nil
}

func (mongoClient *MongoClient) InsertCrater(crater data.Crater, tileId string) error {
	filter := bson.D{{"tileId", tileId}}
	update := bson.D{{"$push", bson.D{{"craters", crater}}}}
	// set upsert to true, insert a new one if not exit
	_, err := mongoClient.craterColl.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}
