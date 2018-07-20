package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"wmq/constant"
	es "wmq/entity/server"
	"wmq/logger"
	"wmq/store"
	"time"
)

type connInfo struct {
	conn net.Conn
	tpe  int
}

var sub_conn_pool sync.Map

func get(quene string) []connInfo {
	v, ok := sub_conn_pool.Load(quene)
	if !ok {
		return nil
	}
	conns, ok := v.([]connInfo)
	if !ok {
		return nil
	}
	return conns
}

func del(quene string, idx int) {
	if conns := get(quene); len(conns) > 0 {
		var temps []connInfo
		temps = append(temps, conns[:idx]...)
		if idx < len(conns) {
			temps = append(temps, conns[idx+1:]...)
		}
		sub_conn_pool.Store(quene, temps)
	}
}

func add(quene string, tpe int, conn net.Conn) {
	// 回复的连接不需要添加（使用推送时的通道）
	if tpe == constant.MSG_TYPE_REPLY {
		return
	}

	var infos []connInfo
	if v, ok := sub_conn_pool.Load(quene); ok {
		infos, ok = v.([]connInfo)
		if !ok {
			return
		}
	}

	// 某些情况下不连接不需要缓存
	for idx, info := range infos {
		// 已经存在相同的连接则替换成最新的
		// 注：为了能在一台机器上测试多个监听，此处将监听的连接放开
		if info.tpe == tpe && tpe != constant.MSG_TYPE_SUB {
			infos[idx].conn = conn
			return
		}
	}

	c := connInfo{
		conn: conn,
		tpe:  tpe,
	}
	infos = append(infos, c)
	sub_conn_pool.Store(quene, infos)
}

type server struct {
	conn      net.Conn
	queneName string // 队列名称
	log       *logger.Logger
	mq        *store.MsgQuene
}

func StartTCP(port string) {
	log := logger.NewStdLogger(true, true, true, true, true)
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%s", port))
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Trace("%s", err)
		os.Exit(-1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Trace("%s", err)
			continue
		}
		s := &server{
			conn: conn,
			log:  log,
			mq:   &store.MsgQuene{CanRead: make(chan bool, store.MAX_SIZE)},
		}
		go s.handle()
	}
}

func (s *server) handle() {
	r := bufio.NewReader(s.conn)
	for {
		content, err := r.ReadString(constant.END_SIGN)
		if err != nil {
			s.log.Error("读取数据失败：%s", err)
			s.conn.Close()
			break
		}
		s.log.Trace("服务端收到的数据：%s", content)
		if len(content) == 0 {
			continue
		}
		var msgInfo es.MsgInfo
		err = json.Unmarshal([]byte(content), &msgInfo)
		if err != nil {
			s.log.Error("消息格式错误，请确认：%s", err)
			continue
		}
		// TODO 先持久化
		if msgInfo.MsgType != constant.MSG_TYPE_SUB {
			go s.mq.Push(&store.MsgDetail{MsgInfo: msgInfo})
		}
		// 保存客户端的连接
		go add(msgInfo.MsgQuene, msgInfo.MsgType, s.conn)
		go s.send()
		//if msgInfo.MsgType != constant.MSG_TYPE_SUB {
		//	// 推送
		//	go s.push(&msgInfo)
		//}
	}
}

func (s *server) send() {
	for {
		s.log.Trace("共有多少条：%d", s.mq.Count)
		if s.mq.Count == 0 {
			<-s.mq.CanRead
		}
		d := s.mq.Pop()
		s.log.Trace("从缓存中读取的内容：%v", d)
		if d == nil {
			continue
		}
		// 根据失败的次数，依次延迟时间自动发送
		if d.SentTimes <= store.MAX_SEND_TIMES {
			switch d.SentTimes {
			case 0,1:
				go s.push(d)
			case 2,3:
				go func(){
					time.Sleep(time.Second)
					s.push(d)
				}()
			case 4,5:
				go func(){
					time.Sleep(time.Minute)
					s.push(d)
				}()
			case 6,7:
				go func(){
					time.Sleep(30*time.Minute)
					s.push(d)
				}()
			case 8,9:
				go func(){
					time.Sleep(time.Hour)
					s.push(d)
				}()
			case 10:
				go func(){
					time.Sleep(24*time.Hour)
					s.push(d)
				}()
			}
		}
	}
}

// 将消息推送给各个监听的客户端
func (s *server) push(msgDetail *store.MsgDetail) {
	msgInfo := msgDetail.MsgInfo
	infos := get(msgInfo.MsgQuene)
	content, _ := json.Marshal(msgInfo)
	content = append(content, '\n')
	isSendSuccess := false
	for idx, info := range infos {
		// 如果是回复消息，则不发给监听者，只发给推送者
		if msgInfo.MsgType == constant.MSG_TYPE_REPLY {
			if info.tpe == constant.MSG_TYPE_SUB {
				continue
			}
		}
		// 只推送给监听者
		if msgInfo.MsgType == constant.MSG_TYPE_PUB {
			if info.tpe == constant.MSG_TYPE_PUB {
				continue
			}
		}
		s.log.Trace("开始推送消息：%+v", msgInfo)
		_, err := info.conn.Write(content)
		if err != nil {
			s.log.Error("推送消息给客户端失败。[conn:%v], [err:%s]", info.conn, err)
			del(msgInfo.MsgQuene, idx)
			continue
		}
		isSendSuccess = true
	}
	if !isSendSuccess{
		s.log.Trace("重新添加回缓存中：%+v", msgDetail)
		go s.mq.Push(&store.MsgDetail{MsgInfo: msgInfo,
			SentTimes:msgDetail.SentTimes+1,
			LastTime:time.Now(),
			Status:func()int{if msgDetail.SentTimes > store.MAX_SEND_TIMES {return 2}else{return 0}}(),
		})
	}
}

func (s *server) response() {
}
