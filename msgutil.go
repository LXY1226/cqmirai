package main

import (
	"gitee.com/LXY1226/logging"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (c *CMiraiConn) TransMsgToMirai(msg []byte) []byte {
	//iter := IteratorPool.BorrowIterator(msg)
	//iter.
	req := new(cqRequest)
	err := json.Unmarshal(msg, &req)
	if err != nil {
		logging.WARN("解析CQ消息失败: ", err.Error())
		return nil
	}
	logging.INFO("< ", req.Action)
	var cqResp *cqResponse
	switch req.Action {
	case "send_msg":
		cqResp = c.sendMsg(req.Params)
	case "get_group_member_info":
		cqResp = c.getGroupMemberInfo(req.Params)
	case "set_group_ban":
		cqResp = c.setGroupBan(req.Params)
	case "get_group_member_list":
		cqResp = c.getGroupMemberList(req.Params)
	case "get_group_list":
		cqResp = c.getGroupList(req.Params)
	default:
		logging.INFO("< 未知请求：", string(req.Params))
	}

	if cqResp == nil {
		return append([]byte(`{"data":null,`), append(req.Echo, `,"retcode":0,"status":"ok"}`...)...)
	}
	cqResp.Echo = req.Echo
	o, err := json.Marshal(cqResp)
	if err != nil {
		logging.WARN("生成CQ回执失败: ", err.Error())
		return nil
	}
	return o
}

func (c *CMiraiConn) TransMsgToCQ(msg []byte) []byte {
	miraiMsg := new(Message)
	err := json.Unmarshal(msg, miraiMsg)
	if err != nil {
		logging.WARN("解析Mirai消息失败: ", err.Error())
		return nil
	}
	logging.INFO("> ", miraiMsg.Type)
	switch miraiMsg.Type {
	case "GroupMessage":
		return c.MiraiGroupMessage(miraiMsg)
	case "FriendMessage":
		return c.MiraiFriendMessage(miraiMsg)
	default:
		return nil

	}
}
