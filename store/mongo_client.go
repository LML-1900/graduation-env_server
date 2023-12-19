package store

import (
	"context"
	"env_server/data"
	"fmt"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoClient struct {
	client        *mongo.Client
	staticDemColl *mongo.Collection
	dynamicColl   *mongo.Collection
}

func NewMongoClient(mongo *mongo.Client) *MongoClient {
	mongoClinet := MongoClient{}
	mongoClinet.client = mongo
	databaseName := viper.GetString("mongodb.database")
	mongoClinet.staticDemColl = mongo.Database(databaseName).Collection(viper.GetString("mongodb.staticDemColl"))
	mongoClinet.dynamicColl = mongo.Database(databaseName).Collection(viper.GetString("mongodb.dynamicColl"))
	return &mongoClinet
}

func (mongoClient *MongoClient) InsertDemData(demData data.DemData) error {
	result, err := mongoClient.staticDemColl.InsertOne(context.TODO(), demData)
	if err != nil {
		return err
	}
	fmt.Println(result.InsertedID)
	return nil
}

//func (mongoClient *MongoClient) PrintInfoByName(name string) error {
//	var result bson.M
//	err := mongoClient.staticColl.FindOne(context.TODO(), bson.D{{"name", name}}).Decode(&result)
//	if err == mongo.ErrNoDocuments {
//		fmt.Printf("No document was found with the name %s\n", name)
//		return nil
//	}
//	if err != nil {
//		return err
//	}
//
//	jsonData, err := json.MarshalIndent(result, "", "    ")
//	if err != nil {
//		return err
//	}
//	fmt.Printf("%s\n", jsonData)
//	return nil
//}
