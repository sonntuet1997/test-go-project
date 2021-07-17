package main

import (
	"fmt"
	"time"
)

func f(from string, c chan<- string) {
	c <- from
}

func main() {

	messages := make(chan string)
	go func() {
		time.Sleep(time.Second)
		f("zxcasd", messages)
	}()
	msg := <-messages
	fmt.Println(msg)

}
