package main

import (
	"bytes"
	"gitee.com/LXY1226/logging"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"strings"
	"time"
)

func parseMsgChain(miraiMsg *Message, cqMsg *cqMessage) {
	msgs := miraiMsg.MessageChain
	cqMsg.MessageID = msgs[0].ID
	cqMsg.Time = msgs[0].Time
	var msgBuf bytes.Buffer
	for _, msg := range msgs[1:] {
		switch msg.Type {
		case "Plain":
			msgBuf.WriteString(string(msg.Text)[1 : len(msg.Text)-1])
			//buf.Write(msg.GetStringBytes("text"))
		case "Face":
			msgBuf.WriteString(msg.Name)
			//buf.WriteString("[CQ:face,id=");buf.WriteString(strconv.Itoa(msg.GetInt("time")))
			//buf.Write(msg.GetStringBytes("text"))
		//case "Quote":
		//	buf.WriteString("[CQ:at,qq=");buf.WriteString(strconv.Itoa(chain[0].GetInt("targetId")));buf.WriteByte(']')
		case "At":
			msgBuf.WriteString("[CQ:at,qq=")
			msgBuf.WriteString(strconv.Itoa(msg.Target))
			msgBuf.WriteByte(']')
		case "Image":
			msgBuf.WriteString("[CQ:image,file=")
			msgBuf.WriteString(msg.ImageID)
			msgBuf.WriteString(",url=")
			msgBuf.WriteString(msg.URL)
			msgBuf.WriteByte(']')
		}
	}
	cqMsg.Message = jsoniter.WrapString(msgBuf.String())
	cqMsg.RawMessage = msgBuf.String()
}

func formatPerm(miraiPrem string) string {
	switch miraiPrem {
	case "ADMINISTRATOR":
		return "admin"
	default:
		return strings.ToLower(miraiPrem)
	}
}

func (c *CMiraiWSRConn) MiraiGroupMessage(miraiMsg *Message) []byte {
	o, err := json.Marshal(cqGroupMemberInfoRsp{
		GroupID:      miraiMsg.Sender.Group.ID,
		LastSentTime: time.Now().Unix(),
		Nickname:     miraiMsg.Sender.MemberName,
		Role:         formatPerm(miraiMsg.Sender.Permission),
		UserID:       miraiMsg.Sender.ID,
	})
	if err != nil {
		logging.WARN("生成用户群缓存错误：", err.Error())
		return nil
	}
	if _, ok := userData[miraiMsg.Sender.ID]; !ok {
		userData[miraiMsg.Sender.ID] = make(map[int][]byte)
	}
	userData[miraiMsg.Sender.ID][miraiMsg.Sender.Group.ID] = o
	cqMsg := new(cqMessage)
	if miraiMsg.Sender.ID == 80000000 {
		cqMsg.Anonymous = new(cqAnonymous)
	}
	cqMsg.GroupID = miraiMsg.Sender.Group.ID
	parseMsgChain(miraiMsg, cqMsg)
	cqMsg.MessageType = "group"
	cqMsg.PostType = "message" //TODO 需要更多适配?
	cqMsg.SelfID = c.miraiConn.i64qNumber
	cqMsg.Sender.Nickname = miraiMsg.Sender.MemberName
	cqMsg.Sender.Role = formatPerm(miraiMsg.Sender.Permission)
	cqMsg.Sender.UserID = miraiMsg.Sender.ID
	cqMsg.UserID = miraiMsg.Sender.ID
	o, err = json.Marshal(cqMsg)
	if err != nil {
		logging.WARN("生成CQ回复错误：", err.Error())
		return nil
	}
	return o
}

func (c *CMiraiWSRConn) MiraiFriendMessage(miraiMsg *Message) []byte {
	cqMsg := new(cqMessage)
	parseMsgChain(miraiMsg, cqMsg)
	cqMsg.MessageType = "private"
	cqMsg.PostType = "message" //TODO 需要更多适配?
	cqMsg.SelfID = c.miraiConn.i64qNumber
	cqMsg.Sender.Nickname = miraiMsg.Sender.MemberName
	cqMsg.Sender.UserID = miraiMsg.Sender.ID
	cqMsg.UserID = miraiMsg.Sender.ID
	o, err := json.Marshal(cqMsg)
	if err != nil {
		logging.WARN("生成CQ回复错误：", err.Error())
		return nil
	}
	return o
}
