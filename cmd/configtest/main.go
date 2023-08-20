package main

import (
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/config"
)

func main() {
	// int, string or other single type:
	// smart parse, such as 123, "123" to 123
	//
	// slice type:
	// when single value like ["123"], smart parse to [123].
	// when multi value like ["123","234"], directly parse to ["123","234"]
	config.Setup("")
	fmt.Printf("SubUid value is %d,  type %T \n", config.HuoBiApiSetting.SubUid, config.HuoBiApiSetting.SubUid)
}
