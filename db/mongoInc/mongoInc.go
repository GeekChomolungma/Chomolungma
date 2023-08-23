package mongoInc

import (
	"context"

	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
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
	MongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
}

type MetaType = int

const (
	BinanTest MetaType = iota
	BinanKline
	BinanAccountInfo
	BinanSyncFlag
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

func (mc *MetaCollection[T]) Retrieve(keyName string, keyValue interface{}, value T) {
	switch mc.DataType {
	case BinanTest, BinanSyncFlag, BinanKline:
		filter := bson.D{{keyName, keyValue}}
		//filter := bson.M{keyName: keyValue}
		mc.Collection.FindOne(context.TODO(), filter).Decode(value)
	}
}

func (mc *MetaCollection[T]) Store(key string, value T) {
	result, err := mc.Collection.InsertOne(context.TODO(), value)
	if err != nil {
		applogger.Error("mongo Inc k-v stored failed, e: %v.", err)
	} else {
		applogger.Info("mongo Inc k-v stored %s, _id is: %v.", key, result)
	}
}

func (mc *MetaCollection[T]) Remove(key string) {
	switch mc.DataType {
	case BinanTest:
		filter := bson.D{{"name", key}}
		count, err := mc.Collection.DeleteMany(context.TODO(), filter)
		if err != nil {
			applogger.Error("mongo Inc k-v remove failed, e: %v.", err)
		} else {
			applogger.Info("mongo Inc k-v remove %s, deleted count is: %v.", key, count)
		}
	}
}
