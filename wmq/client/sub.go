package client

import (
	"bufio"
	"net"
	"wmq/constant"
	ec "wmq/entity/client"
	"wmq/logger"

	"encoding/json"
)

type sub struct {
	Host     string
	Port     string
	Queen    string
	conn     net.Conn
	log      *logger.Logger
	callBack func([]byte) []byte
}

func Subscript(host, port, queen string, callback func([]byte) []byte) {
	c := Client{
		tpe: constant.MSG_TYPE_SUB,
	}
	c.sub = sub{
		Host:     host,
		Port:     port,
		Queen:    queen,
		callBack: callback,
		log:      logger.NewStdLogger(true, true, true, true, true),
	}

	c.listenForTCP()
	c.sendSub()
	c.read()
}

func (c *sub) sendSub() {
	info := ec.MsgInfo{MsgType: constant.MSG_TYPE_SUB, MsgQuene: c.Queen}
	data, _ := json.Marshal(info)
	c.conn.Write(append(data, []byte("\n")...))
}

func (c *sub) read() {
	read := bufio.NewReader(c.conn)
	for {
		content, err := read.ReadString(constant.END_SIGN)
		c.log.Trace("监听到的内容为：%s, %d", content, len(content))
		if err != nil {
			c.log.Error("读取监听消息内容错误：%s", err)
			c.conn.Close()
			break
		}
		if len(content) == 0 {
			continue
		}
		go c.exec(content)
	}
}

func (c *sub) exec(content string) {
	var msg ec.MsgInfo
	err := json.Unmarshal([]byte(content), &msg)
	if err != nil {
		c.log.Error("读取监听消息内容格式错误：%s，%v", err, content)
		return
	}
	if msg.MsgType != constant.MSG_TYPE_PUB{
		return
	}
	data, _ := json.Marshal(msg.MsgContent)
	rtn := c.callBack(data)
	msg.MsgContent = string(rtn)
	msg.MsgType = constant.MSG_TYPE_REPLY
	data, _ = json.Marshal(msg)
	c.conn.Write(append(data, []byte("\n")...))
}
