package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {
	fmt.Println("Hello, World!")
	fmt.Println(strconv.FormatInt(time.Now().UnixMilli(), 10))
}
