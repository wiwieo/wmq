package client

import (
	"fmt"
	"net"
	"os"
	"sync"
	"wmq/constant"
)

var conn_pool sync.Map

func store(name string, conn net.Conn) {
	conn_pool.Store(name, conn)
}

func del(name string) {
	conn_pool.Delete(name)
}

func get(name string) (net.Conn, bool) {
	v, ok := conn_pool.Load(name)
	if ok {
		conn, ok := v.(net.Conn)
		if ok {
			return conn, true
		}
	}
	return nil, false
}

type Client struct {
	tpe int
	sub
	pub
	mux sync.Mutex
}

func dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

// 监听
func (c *Client) listenForTCP() {
	var addr string
	switch c.tpe {
	case constant.MSG_TYPE_PUB:
		addr = fmt.Sprintf("%s:%s", c.pub.Host, c.pub.Port)
	case constant.MSG_TYPE_SUB:
		addr = fmt.Sprintf("%s:%s", c.sub.Host, c.sub.Port)
	}
	conn, ok := get(fmt.Sprintf("%s:%d", addr, c.tpe))
	var err error
	if !ok {
		conn, err = dial("tcp", addr)
		if err != nil {
			fmt.Println("建立连接失败：", err)
			os.Exit(-1)
		}
		if c.tpe == constant.MSG_TYPE_PUB {
			go c.pub.timeout()
			if c.pub.isWaitReply {
				go c.pub.waitReply()
			}
		}
		store(fmt.Sprintf("%s:%d", addr, c.tpe), conn)
	}
	switch c.tpe {
	case constant.MSG_TYPE_PUB:
		c.pub.conn = conn
	case constant.MSG_TYPE_SUB:
		c.sub.conn = conn
	}
}
