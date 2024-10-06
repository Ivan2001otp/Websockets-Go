package config

import (
	"context"
	"log"
	"os"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDBClient *mongo.Client;

func ConnectDatabase(){
	clientOptions := options.Client().ApplyURI(os.Getenv("DB_URL"))

	client,err := mongo.Connect(context.TODO(),clientOptions)

	MongoDBClient = client;

	if err!=nil{
		log.Fatal(err);
	}

	err = MongoDBClient.Ping(context.TODO(),nil);

	if err!=nil{
		log.Fatal(err)
	}

	log.Println("Database connected.")
}