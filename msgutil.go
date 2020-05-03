package main

import (
	"gitee.com/LXY1226/logging"
	"github.com/valyala/fastjson"
)

func (c *CMiraiWSRConn) TransMsgToMirai(msg []byte) []byte {
	parser := parserPool.Get()
	defer parserPool.Put(parser)
	j, err := parser.ParseBytes(msg)
	if err != nil {
		logging.WARN("解析CQ消息失败: ", err.Error())
		return nil
	}
	switch string(j.GetStringBytes("action")) {
	case "send_msg":
		return c.sendMsg(j, msg)
	case "get_group_member_info":
		return c.get_group_member_info(j)
	default:
		return EmptyCQResp(j)
	}
}

func EmptyCQResp(j *fastjson.Value) []byte {
	echo := j.Get("echo").MarshalTo(nil)
	return append([]byte("{"), append(echo, '}')...)
}

func (c *CMiraiWSRConn) TransMsgToCQ(msg []byte) []byte {
	parser := parserPool.Get()
	defer parserPool.Put(parser)
	j, err := parser.ParseBytes(msg)
	if err != nil {
		logging.WARN("解析Mirai消息失败: ", err.Error())
		return nil
	}
	switch string(j.GetStringBytes("type")) {
	case "GroupMessage":
		return c.MiraiGroupMessage(j)
	default:
		return nil

	}
}
