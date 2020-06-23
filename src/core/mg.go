package core

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type Mongo struct {
  *mongo.Client
  *mongo.Database
  ctx context.Context
}


func ConnectMongo(config *Config)(cli *Mongo) {
  mgoOpt := options.Client().ApplyURI(config.MongoURL)
  mgoOpt.SetMaxPoolSize(50)
  mgoOpt.SetMinPoolSize(5)
  mgoOpt.SetConnectTimeout(12 * time.Second)

  client, err := mongo.NewClient(mgoOpt)
  if err != nil {
		log.Print("config: ", config.MongoURL)
    log.Fatal(err)
	}
	
  if err = client.Connect(context.TODO()); err != nil {
		log.Print("config: ", config.MongoURL)
    log.Fatal(err)
  }
	
  log.Print("Conneced to Mongdb", config.MongoURL)

  log.Print("Mongo DB name: '", config.MongoDBName, "'")
  db := client.Database(config.MongoDBName)
  if db == nil {
    log.Fatal("cannot use DB")
  }

  return &Mongo{ client, db, context.TODO() }
}