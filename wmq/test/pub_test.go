package test

import (
	"testing"
	"fmt"
	"wmq/client"
	es "wmq/entity/server"
	"wmq/store"
	"wmq/constant"
)

func Benchmark_push(b *testing.B){
	mq := &store.MsgQuene{CanRead:make(chan bool)}
	for i:=0; i< b.N;i ++ {
		d := &store.MsgDetail{
			MsgInfo: es.MsgInfo{
				MsgType:    constant.MSG_TYPE_PUB,
				MsgQuene:   fmt.Sprintf("test", i),
				MsgContent: fmt.Sprintf("test", i),
			},
		}
		mq.Push(d)
	}


	for i:=0; i< b.N;i ++ {
		d := mq.Pop()
		if d == nil{
			fmt.Println("没有了")
		}else {
			fmt.Println(d)
		}
	}
}

func Benchmark_publish(b *testing.B){
	c := &client.Client{}
	for i:=0; i< b.N;i ++{
		p := Param{Id:i}
		c.Publish("localhost", "44444", "wmq_test", p, false, func(data []byte){
			fmt.Println("pub: 收到回复，", string(data))
		})
	}
	//time.Sleep(10*time.Second)
}

func TestPublish(t *testing.T){
	times := 100000
	c := &client.Client{}
	for i:=0; i< times;i ++{
		go func(idx int){
			c.Publish("localhost", "44444", "wmq_test", Param{Id:idx}, false, func(data []byte){
				fmt.Println("pub: 收到回复，", string(data))
			})
		}(i)
	}
	//time.Sleep(60*time.Second)
}

type Param struct {
	Id int
}