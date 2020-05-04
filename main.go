package main

import (
	"gitee.com/LXY1226/logging"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"net/http"
	"strconv"
	"time"
)

const (
	TypeWSR = iota
)

type CMiraiConn struct {
	authKey    string // "1234567890"
	qNumber    string // "2702342827"
	i64qNumber int    // 2702342827
	miraiAddr  string // "127.0.0.1:8086"
	sessionKey string
	miraiConn  *websocket.Conn

	cqConn     *websocket.Conn
	cqConnType int
	cqAddr     string
}

type CMiraiWSRConn struct {
	cqAddr string // "127.0.0.1:8080"
	*websocket.Conn
	miraiConn *CMiraiConn
}

var parserPool fastjson.ParserPool

var userData map[int]map[int][]byte
var IteratorPool jsoniter.IteratorPool

func main() {
	/*	m := sJson.Unmarshal([]byte("{\"action\": \"get_group_member_info\", \"params\": {\"self_id\": 2702342827, \"group_id\": 1065962966, \"user_id\": 767763591, \"no_cache\": true}, \"echo\": {\"seq\": 77}}"))
		println(m)
		os.Exit(0)*/
	IteratorPool = jsoniter.Config{EscapeHTML: false}.Froze()
	userData = make(map[int]map[int][]byte)
	logging.Init()

	miraiConn := CMiraiConn{
		authKey:    authKey,
		qNumber:    strconv.Itoa(qNumber),
		i64qNumber: qNumber,
		miraiAddr:  miraiAddr,
		cqConnType: TypeWSR,
		cqAddr:     cqWSRAddr,
	}
	for !miraiConn.ConnectMirai() {
		time.Sleep(3 * time.Second)
	}
	for !miraiConn.ConnectCQBot() {
		time.Sleep(3 * time.Second)
	}
	miraiConn.Redirect()
}

func (c *CMiraiConn) ConnectMirai() bool {
	logging.INFO("尝试连接至Mirai: ws://", c.miraiAddr)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI("http://" + c.miraiAddr + "/auth")
	req.Header.SetMethod("POST")
	req.SetBodyString("{\"authKey\": \"" + c.authKey + "\"}")
	err := fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("请求Mirai会话失败: ", err.Error())
		return false
	}
	parser := parserPool.Get()
	defer parserPool.Put(parser)
	var j *fastjson.Value
	j, err = parser.ParseBytes(resp.Body())
	if err != nil {
		logging.WARN("解析Mirai会话失败: ", err.Error())
		return false
	}
	req.SetRequestURI("http://" + c.miraiAddr + "/verify")
	c.sessionKey = string(j.GetStringBytes("session"))
	req.SetBodyString("{\"sessionKey\": \"" + c.sessionKey + "\",\"qq\":\"" + c.qNumber + "\"}")
	err = fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("验证Mirai会话失败: ", err.Error())
		return false
	}
	j, err = parser.ParseBytes(resp.Body())
	if err != nil {
		logging.WARN("解析Mirai会话失败: ", err.Error())
		return false
	}
	if j.GetInt("code") != 0 {
		logging.WARN("解析Mirai会话失败: ", string(j.GetStringBytes("msg")))
		return false
	}
	c.miraiConn, _, err = websocket.DefaultDialer.Dial("ws://"+c.miraiAddr+"/all?sessionKey="+c.sessionKey, nil)
	if err != nil {
		logging.WARN("连接至Mirai失败: ", err.Error())
		return false
	}
	return true
}

func (c *CMiraiConn) ConnectCQBot() bool {
	var err error
	switch c.cqConnType {
	case TypeWSR:
		logging.INFO("尝试连接至CQbot: ws://", c.cqAddr)
		c.cqConn, _, err = websocket.DefaultDialer.Dial("ws://"+c.cqAddr+"/ws/", http.Header{
			"X-Self-ID":     []string{c.qNumber},
			"X-Client-Role": []string{"Universal"},
			"User-Agent":    []string{"MiraiCQHttp/0.0.1"},
		})
		if err != nil {
			logging.WARN("连接至CQbot失败: ", err.Error())
			return false
		}
	}
	return true
}

// 阻塞
func (c *CMiraiConn) Redirect() {
	logging.INFO("连接已建立")
	go func() {
		for {
			t, message, err := c.miraiConn.ReadMessage()
			if err != nil {
				logging.ERROR("从Mirai读取消息失败: ", err.Error())
				c.miraiConn.Close()
				for !c.ConnectMirai() {
					time.Sleep(3 * time.Second)
				}
				continue
			}
			if t == websocket.TextMessage {
				err := c.cqConn.WriteMessage(websocket.TextMessage, c.TransMsgToCQ(message))
				if err != nil {
					logging.ERROR("向CQbot发送消息失败: ", err.Error())
					c.cqConn.Close()
					for !c.ConnectCQBot() {
						time.Sleep(3 * time.Second)
					}
					continue
				}
			} else {
				logging.WARN("未知非文本消息")
			}
		}
	}()

	for {
		t, message, err := c.cqConn.ReadMessage()
		if err != nil {
			logging.ERROR("从CQBot读取消息失败: ", err.Error())
			c.cqConn.Close()
			for !c.ConnectCQBot() {
				time.Sleep(3 * time.Second)
			}
			continue
		}
		if t == websocket.TextMessage {
			err = c.cqConn.WriteMessage(websocket.TextMessage, c.TransMsgToMirai(message))
			if err != nil {
				logging.ERROR("向CQBot回复失败: ", err.Error())
				c.cqConn.Close()
				for !c.ConnectCQBot() {
					time.Sleep(3 * time.Second)
				}
				continue
			}
		} else {
			logging.WARN("未知非文本消息")
		}

	}
}
