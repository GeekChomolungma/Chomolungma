package main

import "fmt"

func main() {
	a := []int{1, 2, 3, 4, 5}
	b := a
	a[0] = 10
	fmt.Println("a is ", a) // [10 2 3 4 5]
	fmt.Println("b is ", b) // [10 2 3 4 5]

	a = a[1:]
	fmt.Println("a is ", a) // [2 3 4 5]
	fmt.Println("b is ", b) // [10 2 3 4 5]

	mapA := make(map[int]int)
	mapA[0] = 15
	mapA[1] = 1
	mapA[2] = 2
	mapB := mapA
	delete(mapA, 0)
	println("a map is", mapA[0]) // a map is 0
	println("b map is", mapB[0]) // b map is 0
}
