package main

import (
	"fmt"
	"sort"
	"time"
)

func main() {
	t := time.Now()
	sc := []int64{33, 87, 1, 34, 5, 999, 222, 4445, 666755, 222323232}
	sort.Slice(sc, func(i, j int) bool {
		return sc[i] < sc[j]
	})
	fmt.Println(sc)
	fmt.Println(sc[len(sc)-1])
	dur := time.Since(t)
	fmt.Println(dur)

	sc = sc[1:]
	fmt.Println(len(sc))
	fmt.Println(sc[len(sc)-1])

	TickMap := make(map[int64]int)
	for k := range TickMap {
		var test []int
		fmt.Println("test length is", k, len(test))
	}
}
