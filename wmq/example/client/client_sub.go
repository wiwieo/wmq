package main

import (
	"wmq/client"
	"encoding/json"
	"fmt"
)

func main() {
	client.Subscript("localhost", "44444", "wmq_test", callback)
}

func callback(data []byte) []byte{
	var p Param
	json.Unmarshal(data, &p)
	return []byte(fmt.Sprintf("%d", p.Id+1))
}

type Param struct {
	Id int
}
