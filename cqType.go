package main

import jsoniter "github.com/json-iterator/go"

type cqFriendMessage struct {
}

type cqMessage struct {
	Font        int          `json:"font,omitempty"`
	Message     jsoniter.Any `json:"message"`
	MessageID   int          `json:"message_id,omitempty"`
	MessageType string       `json:"message_type,omitempty"`
	PostType    string       `json:"post_type,omitempty"`
	RawMessage  string       `json:"raw_message,omitempty"`
	SelfID      int          `json:"self_id,omitempty"`
	Sender      cqSender     `json:"sender,omitempty"`
	SubType     string       `json:"sub_type,omitempty"`
	Time        int          `json:"time,omitempty"`
	UserID      int          `json:"user_id,omitempty"`
	Anonymous   *cqAnonymous `json:"anonymous"`
	GroupID     int          `json:"group_id,omitempty"`
	Area        string       `json:"area,omitempty"`
	Card        string       `json:"card,omitempty"`
	Level       string       `json:"level,omitempty"`
	Role        string       `json:"role,omitempty"`
	Title       string       `json:"title,omitempty"`
}

type cqMemberLists []cqMemberList

type cqMemberList struct {
	Age             int    `json:"age,omitempty"`
	Area            string `json:"area,omitempty"`
	Card            string `json:"card"`
	CardChangeable  bool   `json:"card_changeable,omitempty"`
	GroupID         int    `json:"group_id,omitempty"`
	JoinTime        int    `json:"join_time,omitempty"`
	LastSentTime    int    `json:"last_sent_time,omitempty"`
	Level           string `json:"level,omitempty"`
	Nickname        string `json:"nickname"`
	Role            string `json:"role,omitempty"`
	Sex             string `json:"sex,omitempty"`
	Title           string `json:"title,omitempty"`
	TitleExpireTime int    `json:"title_expire_time,omitempty"`
	Unfriendly      bool   `json:"unfriendly,omitempty"`
	UserID          int    `json:"user_id,omitempty"`
}

type cqMsgData struct {
	Type string       `json:"type"`
	Data jsoniter.Any `json:"data"`
}
type cqMsgDatas []cqMsgData

type cqSender struct {
	Age      int    `json:"age,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Sex      string `json:"sex,omitempty"`
	UserID   int    `json:"user_id,omitempty"`
	Role     string `json:"role,omitempty"`
}

type cqAnonymous struct {
	Flag string `json:"flag"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type cqRequest struct {
	Action string              `json:"action"`
	Params jsoniter.Any        `json:"params"`
	Echo   jsoniter.RawMessage `json:"echo"`
}

type cqGroupMemberInfoReq struct {
	SelfID  int64 `json:"self_id"`
	GroupID int   `json:"group_id"`
	UserID  int   `json:"user_id"`
	NoCache bool  `json:"no_cache"`
}

type cqGroupMemberListReq struct {
	SelfID  int64 `json:"self_id"`
	GroupID int   `json:"group_id"`
}

type cqGroupBanReq struct {
	SelfID   int64 `json:"self_id"`
	GroupID  int   `json:"group_id"`
	UserID   int   `json:"user_id"`
	Duration int   `json:"duration"`
}

type cqResponse struct {
	Data    jsoniter.RawMessage `json:"data"`
	Echo    jsoniter.RawMessage `json:"echo"`
	Retcode int                 `json:"retcode"`
	Status  string              `json:"status"`
}

type cqSendMsgResp struct {
	MessageID int `json:"message_id"`
}

type cqGroupMemberInfoResp struct {
	Age             int    `json:"age,omitempty"`
	Area            string `json:"area,omitempty"`
	Card            string `json:"card,omitempty"`
	CardChangeable  bool   `json:"card_changeable,omitempty"`
	GroupID         int    `json:"group_id,omitempty"`
	JoinTime        int    `json:"join_time,omitempty"`
	LastSentTime    int64  `json:"last_sent_time,omitempty"`
	Level           string `json:"level,omitempty"`
	Nickname        string `json:"nickname,omitempty"`
	Role            string `json:"role,omitempty"`
	Sex             string `json:"sex,omitempty"`
	Title           string `json:"title,omitempty"`
	TitleExpireTime int    `json:"title_expire_time,omitempty"`
	Unfriendly      bool   `json:"unfriendly,omitempty"`
	UserID          int    `json:"user_id,omitempty"`
}
