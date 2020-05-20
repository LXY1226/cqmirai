package main

import "gitee.com/LXY1226/logging"

func (c *CMiraiConn) MiraiMemberJoinEvent(miraiEvent *Event) []byte {
	joinEvent := new(MemberJoinEvent)
	err := json.Unmarshal(miraiEvent.data, joinEvent)
	if err != nil {
		logging.WARN("解析Mirai事件消息失败: ", err.Error())
		return nil
	}
	cqMsg := new(cqGroupMemberEvent)
	cqMsg.PostType = "notice"
	cqMsg.NoticeType = "group_increase"
	cqMsg.GroupID = joinEvent.Member.Group.ID
	cqMsg.UserID = joinEvent.Member.ID
	o, err := json.Marshal(cqMsg)
	if err != nil {
		logging.WARN("生成CQ回复错误：", err.Error())
		return nil
	}
	//	str := string(o)
	//	logging.INFO(str)
	//	logging.INFO(string(cqMsg.GroupID))
	return o
}

func (c *CMiraiConn) MiraiMemberLeaveEvent(miraiEvent *Event) []byte {
	leaveEvent := new(MemberLeaveEvent)
	err := json.Unmarshal(miraiEvent.data, leaveEvent)
	if err != nil {
		logging.WARN("解析Mirai事件消息失败: ", err.Error())
		return nil
	}
	cqMsg := new(cqGroupMemberEvent)
	cqMsg.PostType = "notice"
	cqMsg.NoticeType = "group_decrease"
	cqMsg.SubType = "leave"
	cqMsg.GroupID = leaveEvent.Member.Group.ID
	cqMsg.UserID = leaveEvent.Member.ID
	o, err := json.Marshal(cqMsg)
	if err != nil {
		logging.WARN("生成CQ回复错误：", err.Error())
		return nil
	}
	//      str := string(o)
	//      logging.INFO(str)
	//      logging.INFO(string(cqMsg.GroupID))
	return o
}

func (c *CMiraiConn) MiraiMemberLeaveEventKick(miraiEvent *Event) []byte {
	leaveEventKick := new(MemberLeaveEventKick)
	err := json.Unmarshal(miraiEvent.data, leaveEventKick)
	if err != nil {
		logging.WARN("解析Mirai事件消息失败: ", err.Error())
		return nil
	}
	cqMsg := new(cqGroupMemberEvent)
	cqMsg.PostType = "notice"
	cqMsg.NoticeType = "group_decrease"
	cqMsg.SubType = "kick"
	cqMsg.GroupID = leaveEventKick.Member.Group.ID
	cqMsg.UserID = leaveEventKick.Member.ID
	cqMsg.OperatorID = leaveEventKick.Operator.ID
	o, err := json.Marshal(cqMsg)
	if err != nil {
		logging.WARN("生成CQ回复错误：", err.Error())
		return nil
	}
	//      str := string(o)
	//      logging.INFO(str)
	//      logging.INFO(string(cqMsg.GroupID))
	return o
}
