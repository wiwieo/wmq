package main

import (
	"wmq/client"
	"fmt"
	"net/http"
	"math/rand"
	"time"
)


func main() {
	http.HandleFunc("/hello", hello)
	http.ListenAndServe(":8080", nil)
}

func hello (w http.ResponseWriter, r *http.Request){
	c := &client.Client{}
	p := Param{Id:rand.Intn(100)}
	fmt.Println(time.Now(), "推送内容为：", p)
	c.Publish("localhost", "44444", "wmq_test", p, func(data []byte){
		fmt.Println(time.Now(), "pub: 收到回复，", string(data))
	})
	w.Write([]byte("hello"))
}

type Param struct {
	Id int
}