package main

import (
	"gitee.com/LXY1226/logging"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"net/http"
	"os"
	"strconv"
)

type CMiraiConn struct {
	authKey    string // "1234567890"
	qNumber    string // "2702342827"
	i64qNumber int    // 2702342827
	miraiAddr  string // "127.0.0.1:8086"
	sessionKey string
	*websocket.Conn
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
	miraiConn := NewMirai("127.0.0.1:8088", "1234567890", 2702342827)
	miraiConnWSR := miraiConn.NewCQWSR("127.0.0.1:8080")
	miraiConnWSR.ListenAndRedirect()
}

func NewMirai(miraiAddr, authKey string, qNumber int) *CMiraiConn {
	c := CMiraiConn{
		authKey:    authKey,
		qNumber:    strconv.Itoa(qNumber),
		i64qNumber: qNumber,
		miraiAddr:  miraiAddr,
	}
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
		return nil
	}
	parser := parserPool.Get()
	defer parserPool.Put(parser)
	var j *fastjson.Value
	j, err = parser.ParseBytes(resp.Body())
	if err != nil {
		logging.WARN("解析Mirai会话失败: ", err.Error())
		return nil
	}
	req.SetRequestURI("http://" + c.miraiAddr + "/verify")
	c.sessionKey = string(j.GetStringBytes("session"))
	req.SetBodyString("{\"sessionKey\": \"" + c.sessionKey + "\",\"qq\":\"" + c.qNumber + "\"}")
	err = fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("验证Mirai会话失败: ", err.Error())
		return nil
	}
	j, err = parser.ParseBytes(resp.Body())
	if err != nil {
		logging.WARN("解析Mirai会话失败: ", err.Error())
		return nil
	}
	if j.GetInt("code") != 0 {
		logging.WARN("解析Mirai会话失败: ", string(j.GetStringBytes("msg")))
		return nil
	}
	c.Conn, _, err = websocket.DefaultDialer.Dial("ws://"+c.miraiAddr+"/all?sessionKey="+c.sessionKey, nil)
	if err != nil {
		logging.WARN("连接至Mirai失败: ", err.Error())
		return nil
	}
	return &c
}

func (cm *CMiraiConn) NewCQWSR(cqWSRAddr string) *CMiraiWSRConn {
	var err error
	c := CMiraiWSRConn{
		cqAddr:    cqWSRAddr,
		miraiConn: cm,
	}
	logging.INFO("尝试连接至CQbot: ws://", c.cqAddr)
	c.Conn, _, err = websocket.DefaultDialer.Dial("ws://"+c.cqAddr+"/ws/", http.Header{
		"X-Self-ID":     []string{c.miraiConn.qNumber},
		"X-Client-Role": []string{"Universal"},
		"User-Agent":    []string{"MiraiCQHttp/0.0.1"},
	})
	if err != nil {
		logging.WARN("连接至CQbot失败: ", err.Error())
		return nil
	}
	return &c
}

// 阻塞
func (c *CMiraiWSRConn) ListenAndRedirect() {
	logging.INFO("连接已建立")
	//done := make(chan struct{})
	go func() {
		for {
			//defer close(done)
			for {
				t, message, err := c.miraiConn.ReadMessage()
				if err != nil {
					logging.ERROR("从Mirai读取消息失败: ", err.Error())
					os.Exit(0) // TODO 多实例优化
				}
				if t == websocket.TextMessage {
					logging.INFO("> ", string(message))
					err := c.Conn.WriteMessage(websocket.TextMessage, c.TransMsgToCQ(message))
					if err != nil {
						logging.ERROR("向CQbot发送消息失败: ", err.Error())
						os.Exit(0) // TODO 多实例优化
					}
				} else {
					logging.WARN("未知非文本消息")
				}
			}
		}
	}()

	for {
		//select {
		//case <-done:
		//	return
		//case t := <-ticker.C:
		t, message, err := c.Conn.ReadMessage()
		if err != nil {
			logging.ERROR("从CQBot读取消息失败: ", err.Error())
			os.Exit(0) // TODO 多实例优化
		}
		if t == websocket.TextMessage {
			logging.INFO("< ", string(message))
			err = c.Conn.WriteMessage(websocket.TextMessage, c.TransMsgToMirai(message))
			if err != nil {
				logging.ERROR("向CQBot回复失败: ", err.Error())
				os.Exit(0) // TODO 多实例优化
			}
		} else {
			logging.WARN("未知非文本消息")
		}
	}
}
