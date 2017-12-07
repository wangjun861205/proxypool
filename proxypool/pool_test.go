package proxypool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	pool, _ := NewProxyPool(context.Background())
	fmt.Println(len(*pool.Proxys))
	go pool.Serve()
	for i := 0; i < 4; i++ {
		go func() {
			for {
				// fmt.Printf("prepare pop(remaining %d)\n", len(pool.Proxys))
				proxy := pool.Pop()
				// fmt.Println("pop:", proxy.IP, proxy.Port)
				// fmt.Printf("prepare push(remaining %d)\n", len(pool.Proxys))
				pool.Push(proxy)
				// fmt.Println("push:", proxy.IP, proxy.Port)
			}
		}()
	}
	time.Sleep(10 * time.Minute)
}
