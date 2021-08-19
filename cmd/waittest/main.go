package main

import (
	"fmt"
	"sync"
)

func main() {
	isOver := false
	var wg sync.WaitGroup
	wg = sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			fmt.Println("loop number", i)
			wg.Done()
		}(i)
	}
	go func() {
		wg = sync.WaitGroup{}
		wg.Wait()
		isOver = true
		fmt.Println("set isOver")
	}()

	for {
		if isOver {
			fmt.Println("loop over")
			return
		}
	}
}
