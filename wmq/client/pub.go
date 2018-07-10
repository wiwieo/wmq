package client

import (
	"net"
	ec "wmq/entity/client"
	"wmq/logger"

	"bufio"
	"encoding/json"
	"fmt"
	"time"
	"wmq/constant"
)

type pub struct {
	Host     string
	Port     string
	Queen    string
	conn     net.Conn
	log      *logger.Logger
	callBack func([]byte)
	closec   chan bool
	timeoutc chan int
}

func (c *Client) Publish(host, port, queen string, content interface{}, callback func([]byte)) {
	c.mux.Lock()
	for c.pub.conn == nil {
		c.tpe = constant.MSG_TYPE_PUB
		c.pub = pub{
			Host:     host,
			Port:     port,
			Queen:    queen,
			callBack: callback,
			log:      logger.NewStdLogger(true, true, true, true, true),
			closec:   make(chan bool),
			timeoutc: make(chan int),
		}
		c.listenForTCP()
	}
	c.mux.Unlock()
	msgInfo := ec.MsgInfo{
		MsgType:    constant.MSG_TYPE_PUB,
		MsgQuene:   queen,
		MsgContent: content,
	}
	data, _ := json.Marshal(msgInfo)
	c.pub.conn.Write(append(data, []byte("\n")...))
}

func (p *pub) waitReply() {
	read := bufio.NewReader(p.conn)
	for {
		content, err := read.ReadString(constant.END_SIGN)
		if err != nil {
			p.log.Error("推送消息后返回的内容为：%s, %s", err, content)
			break
		}
		if len(content) == 0 {
			continue
		}
		go func() { p.timeoutc <- constant.TIME_OUT }()
		var rtn ec.MsgInfo
		data := []byte(content)
		err = json.Unmarshal(data, &rtn)
		if err != nil {
			p.log.Error("推送消息后返回内容格式错误：%s", err)
			break
		}

		// 回复类型
		if rtn.MsgType == constant.MSG_TYPE_REPLY && p.callBack != nil {
			p.callBack(data)
		}
	}
	//p.conn.Close()
	//del(fmt.Sprintf("%s:%s:%d", p.Host, p.Port, constant.MSG_TYPE_PUB))
}

func (p *pub) timeout() {
	task := time.AfterFunc(constant.TIME_OUT*time.Second, func() {
		p.conn.Close()
		p.conn = nil
		del(fmt.Sprintf("%s:%s:%d", p.Host, p.Port, constant.MSG_TYPE_PUB))
		p.log.Trace("已经超时了。")
	})
	for {
		select {
		case <-p.timeoutc:
			task.Reset(constant.TIME_OUT * time.Second)
		}
	}
}
