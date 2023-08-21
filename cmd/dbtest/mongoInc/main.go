package main

import (
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/db/mongoInc"
)

type TestMeta struct {
	Name string `json:"name" bson:"name"`
}

func main() {
	// config server
	config.Setup("./../../../my.ini")
	mongoInc.Init(config.MongoSetting.Uri)
	fmt.Println("uri is", config.MongoSetting.Uri)

	metaCol := mongoInc.NewMetaCollection[*TestMeta]("authorInfo", "authorProfile", mongoInc.BinanTest)
	value := &TestMeta{}
	metaCol.Retrieve("alex", value)
	fmt.Println("Value Retrieved:", value.Name)

	valueInserted := &TestMeta{
		Name: "zxx",
	}
	metaCol.Store("zxx", valueInserted)
	metaCol.Store("zxx", valueInserted)

	metaCol.Remove("zxx")
}
