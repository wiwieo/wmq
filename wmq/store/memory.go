package store

import (
	"fmt"
	"sync"
	"time"
	"wmq/entity/server"
)

type MsgDetail struct {
	server.MsgInfo
	SentTimes int       // 已发送次数
	Status    int       // 发送状态（0：待发送，1：已发送，2：已死亡）
	LastTime  time.Time // 最后发送时间
	Next      *MsgDetail
	Pre       *MsgDetail
}

type MsgQuene struct {
	Top     *MsgDetail
	Bottom  *MsgDetail
	Count   int
	mux     sync.Mutex
	CanRead chan bool
}

func NewMsgQuene(msg *MsgDetail) *MsgQuene {
	q := &MsgQuene{
		Top:    msg,
		Bottom: msg,
		Count:  1,
	}
	return q
}

func (mq *MsgQuene) Push(msg *MsgDetail) {
	mq.mux.Lock()
	defer mq.mux.Unlock()
	if mq.Count == 0 {
		mq.Top = msg
		mq.Bottom = msg
		mq.Count = 1
		mq.CanRead <- true
		return
	}
	mq.Top.Pre = msg
	msg.Next = mq.Top
	mq.Top = msg
	mq.Count++
	mq.CanRead <- true
}

func (mq *MsgQuene) Pop() *MsgDetail {
	mq.mux.Lock()
	defer mq.mux.Unlock()
	if mq.Bottom == nil && mq.Count == 0 {
		return nil
	}
	b := mq.Bottom
	mq.Count--
	if mq.Bottom == mq.Top && mq.Count == 1 {
		mq.Bottom = nil
		mq.Top = nil
		return b
	}
	mq.Bottom = b.Pre
	if b.Pre != nil {
		b.Pre.Next = nil
	}
	return b
}

func (mq *MsgQuene) List() {
	next := mq.Top
	for {
		if next == nil {
			return
		}
		fmt.Println(next)
		next = next.Next
	}
}
