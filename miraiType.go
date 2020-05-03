package main

import jsoniter "github.com/json-iterator/go"

type Message struct {
	Type         string        `json:"type"`
	MessageChain MessageChains `json:"messageChain"`
	Sender       Sender        `json:"sender"`
}

type MessageChain struct {
	Type     string              `json:"type"`
	Content  string              `json:"content,omitempty"`
	JSON     string              `json:"json,omitempty"`
	Name     string              `json:"name,omitempty"`
	XML      string              `json:"xml,omitempty"`
	ImageID  string              `json:"imageId,omitempty"`
	Time     int                 `json:"time,omitempty"`
	URL      string              `json:"url,omitempty"`
	Path     interface{}         `json:"path,omitempty"`
	Text     jsoniter.RawMessage `json:"text,omitempty"`
	FaceID   int                 `json:"faceId,omitempty"`
	Target   int                 `json:"target,omitempty"`
	Display  string              `json:"display,omitempty"`
	ID       int                 `json:"id,omitempty"`
	GroupID  int                 `json:"groupId,omitempty"`
	SenderID int                 `json:"senderId,omitempty"`
	TargetID int64               `json:"targetId,omitempty"`
	Origin   []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"origin,omitempty"`
}

type MessageChains []MessageChain

type outgoingMessage struct {
	SessionKey   string        `json:"sessionKey"`
	Target       int           `json:"target"`
	MessageChain MessageChains `json:"messageChain"`
}

type MessageResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	MessageID int    `json:"messageId"`
}

type Sender struct {
	ID         int    `json:"id"`
	MemberName string `json:"memberName"`
	Permission string `json:"permission"`
	Group      struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Nickname   string `json:"nickname"`
		Permission string `json:"permission"`
		Remark     string `json:"remark"`
	} `json:"group"`
}
