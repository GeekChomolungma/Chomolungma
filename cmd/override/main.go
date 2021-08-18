package main

import "fmt"

type animal struct {
}

func (a *animal) yell() {
	fmt.Println("zizizi.")
}

func (a *animal) jump() {
	fmt.Println("5m high")
}

func (a *animal) DoSomeAction() {
	a.yell()
	a.jump()
}

type people struct {
	*animal
} 

func (p *people) yell() {
	fmt.Println("hahaha.")
}

func (a *people) jump() int {
	fmt.Println("20m high")
	return 20
}

func main() {
	p := &people{}
	p.yell()
	p.jump()
	p.DoSomeAction()
}
