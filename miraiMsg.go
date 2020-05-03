package main

import (
	"bytes"
	"encoding/json"
	"github.com/valyala/fastjson"
	"strconv"
)

type GroupMessage struct {
	Type         string         `json:"type"`
	MessageChain []MessageChain `json:"messageChain"`
	Sender       Sender         `json:"sender"`
}
type MessageChain struct {
	Type string `json:"type"`
	ID   int    `json:"id,omitempty"`
	Time int    `json:"time,omitempty"`
	Text string `json:"text,omitempty"`
}
type Group struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
}
type Sender struct {
	ID         int    `json:"id"`
	MemberName string `json:"memberName"`
	Permission string `json:"permission"`
	Group      Group  `json:"group"`
}

/*
{
	"anonymous": null,
	"font": 9560640,
	"group_id": 1065962966,
	"message": "1",
	"message_id": 21,
	"message_type": "group",
	"post_type": "message",
	"raw_message": "1",
	"self_id": 2702342827,
	"sender": {
		"age": 19,
		"area": "大连",
		"card": "",
		"level": "潜水",
		"nickname": "105消毒液",
		"role": "owner",
		"sex": "male",
		"title": "",
		"user_id": 767763591
	},
	"sub_type": "normal",
	"time": 1588396270,
	"user_id": 767763591
}
*/

func formMsgChain(buf *bytes.Buffer, chain []*fastjson.Value) {
	var msgBuf bytes.Buffer
	buf.WriteString(",\"message_id\":")
	buf.WriteString(strconv.Itoa(chain[0].GetInt("id")))
	buf.WriteString(",\"time\":")
	buf.WriteString(strconv.Itoa(chain[0].GetInt("time")))
	buf.WriteString(",\"raw_message\":\"")
	for _, msg := range chain[1:] {
		switch string(msg.GetStringBytes("type")) {
		case "Plain":
			//buf.Write(msg.GetStringBytes("text"))
			json.HTMLEscape(&msgBuf, msg.GetStringBytes("text"))
		case "Face":
			//buf.WriteString("[CQ:face,id=");buf.WriteString(strconv.Itoa(msg.GetInt("time")))
			//buf.Write(msg.GetStringBytes("text"))
			json.HTMLEscape(&msgBuf, msg.GetStringBytes("name"))
		//case "Quote":
		//	buf.WriteString("[CQ:at,qq=");buf.WriteString(strconv.Itoa(chain[0].GetInt("targetId")));buf.WriteByte(']')
		case "At":
			msgBuf.WriteString("[CQ:at,qq=")
			msgBuf.WriteString(strconv.Itoa(msg.GetInt("target")))
			msgBuf.WriteByte(']')
		}
	}
	buf.Write(msgBuf.Bytes())
	buf.WriteString("\",\"message\":\"")
	buf.Write(msgBuf.Bytes())
}

func formatPerm(miraiPrem []byte) []byte {
	switch string(miraiPrem) {
	case "ADMINISTRATOR":
		return []byte("admin")
	default:
		return bytes.ToLower(miraiPrem)
	}
}

func (c *CMiraiWSRConn) MiraiGroupMessage(j *fastjson.Value) []byte {
	var buf bytes.Buffer
	buf.WriteString("{\"anonymous\":null") //;buf.WriteString("null") //TODO QNumber == 80000000
	buf.WriteString(",\"font\":0")         //;buf.WriteString("0")
	buf.WriteString(",\"group_id\":")
	buf.WriteString(strconv.Itoa(j.GetInt("sender", "group", "id")))
	formMsgChain(&buf, j.GetArray("messageChain"))
	//buf.WriteString(",\"message\":");
	//buf.WriteString(",\"message_id\":null");
	buf.WriteString("\",\"message_type\":\"group\"")
	buf.WriteString(",\"post_type\":\"message\"")
	//buf.WriteString(",\"raw_message\":null")
	buf.WriteString(",\"self_id\":")
	buf.WriteString(c.miraiConn.qNumber)
	buf.WriteString(",\"sender\":{")

	//buf.WriteString("\"age\":null")
	//buf.WriteString(",\"area\":null")
	//buf.WriteString(",\"card\":null")
	//buf.WriteString(",\"level\":null")
	buf.WriteString("\"nickname\":\"")
	buf.Write(j.GetStringBytes("sender", "memberName"))
	buf.WriteByte('"')
	buf.WriteString(",\"role\":\"")
	buf.Write(formatPerm(j.GetStringBytes("sender", "permission")))
	//buf.WriteString(",\"sex\":null")
	buf.WriteString("\",\"title\":\"\"")
	buf.WriteString(",\"user_id\":")
	buf.WriteString(strconv.Itoa(j.GetInt("sender", "id")))

	buf.WriteString("},\"sub_type\":\"normal\"")
	//buf.WriteString(",\"time\":null")
	buf.WriteString(",\"user_id\":")
	buf.WriteString(strconv.Itoa(j.GetInt("sender", "id")))
	buf.WriteByte('}')
	return buf.Bytes()
}
