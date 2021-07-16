package db

import (
	"github.com/GeekChomolungma/Chomolungma/config"
	_ "github.com/sbunce/bson"
	"gopkg.in/mgo.v2"
	"labix.org/v2/mgo/bson"
)

// return a mongo session
func CreateMarketDBSession() (*mgo.Session, error) {
	mongo, err := mgo.Dial(config.MongoSetting.MarketUrl)
	if err != nil {
		return nil, err
	}
	return mongo, nil
}

func CreateAccountDBSession() (*mgo.Session, error) {
	mongo, err := mgo.Dial(config.MongoSetting.AccountUrl)
	if err != nil {
		return nil, err
	}
	return mongo, nil
}

type Student struct {
	Name   string
	Age    int
	Sid    string
	Status int
}
type Per struct {
	Per []Student
}

var (
	ip = "127.0.0.1"
)

func insert(data interface{}) bool {
	mongo, err := mgo.Dial(ip)
	defer mongo.Close()
	if err != nil {
		return false
	}

	client := mongo.DB("huobi").C("t_student")

	cErr := client.Insert(&data)
	if cErr != nil {
		return false
	}
	return true
}

func findOne() bool {
	mongo, err := mgo.Dial(ip)
	defer mongo.Close()
	if err != nil {
		return false
	}

	client := mongo.DB("mydb_tutorial").C("t_student")
	user := Student{}
	cErr := client.Find(bson.M{"sid": "learn_001"}).One(&user)
	if cErr != nil {
		return false
	}
	return true
}

func findAll() bool {
	mongo, err := mgo.Dial(ip)
	defer mongo.Close()
	if err != nil {
		return false
	}

	client := mongo.DB("mydb_tutorial").C("t_student")
	iter := client.Find(bson.M{"status": 1}).Sort("_id").Skip(1).Limit(15).Iter()
	var stu Student
	var users Per
	for iter.Next(&stu) {
		users.Per = append(users.Per, stu)
	}
	if err := iter.Close(); err != nil {
		return false
	}
	return true
}
