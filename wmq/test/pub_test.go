package test

import (
	"testing"
	"fmt"
	"wmq/client"
	"time"
)

func Benchmark_publish(b *testing.B){
	c := &client.Client{}
	for i:=0; i< b.N;i ++{
		p := Param{Id:i}
		c.Publish("localhost", "44444", "wmq_test", p, func(data []byte){
			fmt.Println("pub: 收到回复，", string(data))
		})
	}
}

func TestPublish(t *testing.T){
	times := 100000
	c := &client.Client{}
	for i:=0; i< times;i ++{
		go func(idx int){
			c.Publish("localhost", "44444", "wmq_test", Param{Id:idx}, func(data []byte){
				fmt.Println("pub: 收到回复，", string(data))
			})
		}(i)
	}
	time.Sleep(30*time.Second)
}

type Param struct {
	Id int
}