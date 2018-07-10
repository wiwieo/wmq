package server

type MsgInfo struct {
	MsgId      int64
	MsgType    int //消息类型：1、监听 2、
	MsgQuene   string
	MsgContent interface{}
}
