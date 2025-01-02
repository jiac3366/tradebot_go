package main

import (
	"fmt"
	"sync"
	"time"
)

type C struct {
	*sync.Mutex
}

type A struct {
	c   C
	val int32
}

type B struct {
	c   C
	val int32
}

const N = 1000000

func main() {
	a := &A{}
	b := &B{}
	go func(a *A) {
		for i := 0; i < N; i++ {
			a.c.Lock()
			a.val++
			a.c.Unlock()
		}
	}(a)

	go func(b *B) {
		for i := 0; i < N; i++ {
			b.c.Lock()
			b.val++
			b.c.Unlock()
		}
	}(b)

	time.Sleep(2 * time.Second)
	fmt.Println(a.val, b.val)
}
