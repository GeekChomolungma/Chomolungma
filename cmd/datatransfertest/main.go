package main

import (
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type TestFull struct {
	ItemA string      `json:"ia"`
	ItemB string      `json:"ib"`
	ItemC string      `json:"ic"`
	Ti    TestInclude `json:"ti"`
}
type TestInclude struct {
	Item string `json:"item"`
}

type Test struct {
	ItemA string `json:"ia"`
}

type Test2 struct {
	ItemB string `json:"ib"`
	ItemC string `json:"ic"`
}

func main() {
	tf := &TestFull{
		ItemA: "test1",
		ItemB: "test2",
		ItemC: "test3",
		Ti: TestInclude{
			Item: "this is item",
		},
	}
	raw, err := jsoniter.Marshal(tf)
	if err != nil {
		fmt.Println("cannot Marshal")
	}

	testForOffsetGoIter(raw)
}

func testForOffsetGoIter(data []byte) {
	// Get the item
	t := time.Now()
	itemstr := jsoniter.Get(data, "ti").Get("item").ToString()
	fmt.Println("the string of json:", string(data), time.Since(t).Nanoseconds())
	fmt.Println("Got the data:", itemstr)

	// use iterator
	c := jsoniter.ConfigDefault
	iter := jsoniter.ParseBytes(c, data)
	fmt.Println("before Read, the current bytes is:", iter.CurrentBuffer())
	t1 := &Test{}
	iter.ReadVal(t1)
	fmt.Println("data t1 ItemA is:", t1.ItemA)
	fmt.Println("after t1 Read, the current bytes is:", iter.CurrentBuffer())

	t2 := &Test2{}
	iter.ReadVal(t2)
	fmt.Println("data t2 ItemB is:", t2.ItemB)
	fmt.Println("after t2 Read, the current bytes is:", iter.CurrentBuffer())
}
