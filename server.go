package main

import (
	"context"
	"env_server/data"
	"env_server/service"
	"fmt"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// new an insert dynamic data service
	dynamicDataService := service.NewInsertDynamicDataService(mongoClient)
	crater := data.Crater{
		Position: data.LonLatPosition{Longitude: 113.34, Latitude: 23.78},
		Width:    5.78,
		Depth:    5.99,
	}
	err = dynamicDataService.InsertCrater(crater)
	if err != nil {
		fmt.Printf("fail to insert crater, err: %s\n", err)
	}
}
