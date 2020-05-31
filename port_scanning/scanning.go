package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}
	number := make(chan bool, 5000)
 	for i := 1;i <= 65535;i++ {
		wg.Add(1)
		number<-true
		go func(p int) {
			defer func() {
				<-number
				wg.Done()
			}()
			c, err := net.Dial(os.Args[1], fmt.Sprintf("%s:%d", os.Args[2], p))
			if err == nil {
				c.Close()
				fmt.Println("端口：", p, "开放")
			}
		}(i)
	}
	wg.Wait()
}
