package main

import "fmt"

func main() {
	tr := &testRaw{
		item: "HHHHH",
	}
	t := tr.Transfer()
	fmt.Println("the Res:", t.butem)
}

type testRaw struct {
	item string
}

func (tr *testRaw) Transfer() *test {
	// compiler will move "test" to heap.
	// so the return data could be used exportly
	// even the function is finished and its stack is deleted.
	return &test{
		butem: tr.item,
	}
}

type test struct {
	butem string
}
