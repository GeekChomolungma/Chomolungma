package mongoInc

import (
	"context"
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// func init() {
// 	var err error
// 	MongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(config.MongoSetting.Uri))
// 	if err != nil {
// 		panic(err)
// 	}
// }

func Init(uri string) {
	var err error
	MongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(config.MongoSetting.Uri))
	if err != nil {
		panic(err)
	}
}

type MetaType = int

const (
	BinanTest MetaType = iota
	BinanKline
	BinanAccountInfo
)

type MetaCollection[T any] struct {
	DatabaseName   string
	CollectionName string
	DataType       MetaType

	Collection *mongo.Collection
}

func NewMetaCollection[T any](dbName, colName string, dataType MetaType) *MetaCollection[T] {
	mc := &MetaCollection[T]{
		DatabaseName:   dbName,
		CollectionName: colName,
		DataType:       dataType,
	}
	collection := MongoClient.Database(dbName).Collection(colName)
	mc.Collection = collection
	return mc
}

func (mc *MetaCollection[T]) Retrieve(key string, value T) {
	fmt.Sprintln("ready to retrieve")
	switch mc.DataType {
	case BinanTest:
		filter := bson.D{{"name", key}}
		mc.Collection.FindOne(context.TODO(), filter).Decode(value)
		fmt.Sprintln("Retrieved key-value: %v, %v", key, value)
	}
}

func (mc *MetaCollection[T]) Store(key string, value interface{}) {

}

func (mc *MetaCollection[T]) Remove(key string) {

}
