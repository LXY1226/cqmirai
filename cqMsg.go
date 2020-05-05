package main

import (
	"bytes"
	"encoding/base64"
	"gitee.com/LXY1226/logging"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"mime/multipart"
	"strconv"
	"strings"
)

func (c *CMiraiConn) DoReq(method, path, param string, body []byte) *fastjson.Value {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI("http://" + c.miraiAddr + path + "?sessionKey=" + c.sessionKey + param)
	req.Header.SetMethod(method)
	req.SetBody(body)
	err := fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("向Mirai请求出错: ", err.Error())
		return nil
	}
	parser := parserPool.Get()
	defer parserPool.Put(parser)
	j, err := parser.ParseBytes(resp.Body())
	if err != nil {
		logging.WARN("解析Mirai回复出错: ", err.Error())
		return nil
	}
	return j
}

func (c *CMiraiConn) uploadImage(dataURI string, imgTarget string) string {
	img, err := base64.StdEncoding.DecodeString(dataURI[9:])
	if err != nil {
		logging.WARN("反Base64出错: ", err.Error())
		return ""
	}
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI("http://" + c.miraiAddr + "/uploadImage")
	req.Header.SetMethod("POST")
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	w.WriteField("sessionKey", c.sessionKey)
	w.WriteField("type", imgTarget)
	fw, err := w.CreateFormFile("img", imgTarget)
	fw.Write(img)
	w.Close()
	req.SetBody(buf.Bytes())
	req.Header.SetContentType(w.FormDataContentType())
	req.MultipartForm()
	err = fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("上传图片出错: ", err.Error())
	}
	parser := parserPool.Get()
	defer parserPool.Put(parser)
	j, err := parser.ParseBytes(resp.Body())
	if err != nil {
		logging.WARN("上传图片解析出错: ", err.Error())
		return ""
	}
	return string(j.GetStringBytes("imageId"))
}

func (c *CMiraiConn) parseCQMsg(msg string, imgTarget string) []MessageChain {
	last := 0
	var mc []MessageChain
	for {
		start := strings.Index(msg[last:], `[CQ`)
		if start == -1 {
			break
		}
		start += last
		mc = append(mc, MessageChain{
			Type: "Plain",
			Text: []byte(strconv.QuoteToASCII(msg[last:start])),
		})
		end := strings.Index(msg[last:], `]`) + last
		last = end + 1
		switch msg[start+4 : start+6] {
		case "at":
			num, err := strconv.Atoi(msg[strings.Index(msg[start+6:end], "qq=")+start+9 : end])
			if err != nil {
				logging.WARN("CQ码解析失败：", msg[start:end])
			}
			mc = append(mc, MessageChain{
				Type:   "At",
				Target: num,
			})
		case "im": //image
			b64Img := msg[strings.Index(msg[start+6:end], "base64://")+start+6 : end]

			mc = append(mc, MessageChain{
				Type:    "Image",
				ImageID: c.uploadImage(b64Img, imgTarget),
			})
		default:
			logging.WARN("CQ码未解析：", msg[start:end])
		}
	}
	mc = append(mc, MessageChain{
		Type: "Plain",
		Text: []byte(strconv.QuoteToASCII(msg[last:])),
	})
	return mc
}

func (c *CMiraiConn) formMsgChain(msg jsoniter.Any, imgTarget string) []MessageChain {
	switch msg.ValueType() {
	case jsoniter.StringValue:
		return c.parseCQMsg(msg.ToString(), imgTarget)
	case jsoniter.ObjectValue:
		m := new(cqMsgData)
		err := json.UnmarshalFromString(msg.ToString(), &m)
		if err != nil {
			logging.WARN("解析CQ消息失败: ", err.Error())
			return nil
		}
		switch m.Type {
		case "at":
			return []MessageChain{{
				Type:   "At",
				Target: m.Data.Get("qq").ToInt(),
			}}
		case "text":
			return c.parseCQMsg(m.Data.Get("text").ToString(), imgTarget)
		case "image":
			return []MessageChain{{
				Type:    "Image",
				ImageID: c.uploadImage(m.Data.Get("file").ToString(), imgTarget),
			}}
		}
	case jsoniter.ArrayValue:
		var cs []MessageChain
		chains := cqMsgDatas{}
		err := json.UnmarshalFromString(msg.ToString(), &chains)
		if err != nil {
			logging.WARN("解析CQ消息失败: ", err.Error())
			return nil
		}
		for _, msg := range chains {
			switch msg.Type {
			case "at":
				cs = append(cs, MessageChain{
					Type:   "At",
					Target: msg.Data.Get("qq").ToInt(),
				})
			case "text":
				cs = append(cs, c.parseCQMsg(msg.Data.Get("text").ToString(), imgTarget)...)
			case "image":
				c.uploadImage(msg.Data.Get("file").ToString(), imgTarget)
				cs = append(cs, MessageChain{
					Type:    "Image",
					ImageID: c.uploadImage(msg.Data.Get("file").ToString(), imgTarget),
				})
			}
		}
		return cs
	default:
		return nil
	}
	return nil
}

func (c *CMiraiConn) sendMsg(params []byte) *cqResponse {
	msg := new(cqMessage)
	err := json.Unmarshal(params, msg)
	if err != nil {
		logging.WARN("解析CQ消息失败: ", err.Error())
		return nil
	}
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("POST")
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	miraiMsg := new(MessageReq)
	miraiMsg.SessionKey = c.sessionKey
	var imgTarget string
	switch msg.MessageType {
	case "group":
		miraiMsg.Target = msg.GroupID
		imgTarget = "group"
		req.SetRequestURI("http://" + c.miraiAddr + "/sendGroupMessage")
	case "private":
		miraiMsg.Target = msg.UserID
		imgTarget = "friend"
		req.SetRequestURI("http://" + c.miraiAddr + "/sendFriendMessage")
	default:
		logging.WARN("尚未实现的msg.MessageType: ", msg.MessageType)
		return nil
	}
	miraiMsg.MessageChain = c.formMsgChain(msg.Message, imgTarget)

	o, err := json.Marshal(miraiMsg)
	if err != nil {
		logging.WARN("生成Mirai消息失败: ", err.Error())
		return nil
	}
	req.SetBody(o)
	err = fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("向Mirai请求出错: ", err.Error())
		return nil
	}
	rj := new(MessageResp)
	err = json.Unmarshal(resp.Body(), rj)
	if err != nil {
		logging.WARN("解析Mirai回复出错: ", err.Error())
		return nil
	}
	rs := new(cqResponse)
	rs.Data, err = json.Marshal(cqSendMsgResp{MessageID: rj.MessageID})
	if err != nil {
		logging.WARN("生成CQ回复出错: ", err.Error())
		return nil
	}
	rs.Retcode = 0
	rs.Status = "ok"
	return rs
}

func (c *CMiraiConn) getGroupMemberInfo(params []byte) *cqResponse {
	msg := new(cqGroupMemberInfoReq)
	err := json.Unmarshal(params, msg)
	if err != nil {
		logging.WARN("解析CQ消息失败: ", err.Error())
		return nil
	}
	if data, ok := userData[msg.UserID][msg.GroupID]; ok {
		return &cqResponse{
			Data:    data,
			Retcode: 0,
			Status:  "ok",
		}
	}
	return &cqResponse{
		Data:    nil,
		Retcode: 0,
		Status:  "ok",
	}
}

func (c *CMiraiConn) setGroupBan(params []byte) *cqResponse {
	msg := new(cqGroupBanReq) // !!!!!!!!!!!!!!!!!!!!!!
	err := json.Unmarshal(params, msg)
	if err != nil {
		logging.WARN("解析CQ消息失败: ", err.Error())
		return nil
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("POST")
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	req.SetRequestURI("http://" + c.miraiAddr + "/mute")
	mReq := new(MuteReq) // !!!!!!!!!!!!!!!!!!!!!!
	mReq.SessionKey = c.sessionKey
	mReq.Target = msg.GroupID
	mReq.MemberID = msg.UserID
	mReq.Time = msg.Duration

	o, err := json.Marshal(mReq)
	if err != nil {
		logging.WARN("生成Mirai消息失败: ", err.Error())
		return nil
	}
	req.SetBody(o)
	err = fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("向Mirai请求出错: ", err.Error())
		return nil
	}

	rj := new(MuteResp) // !!!!!!!!!!!!!!!!!!!!!!

	err = json.Unmarshal(resp.Body(), rj)
	if err != nil {
		logging.WARN("解析Mirai回复出错: ", err.Error())
		return nil
	}
	rs := new(cqResponse)
	rs.Retcode = 0
	rs.Status = "ok"
	return rs
}

func (c *CMiraiConn) getGroupMemberList(params []byte) *cqResponse {
	msg := new(cqGroupMemberListReq)
	err := json.Unmarshal(params, msg)
	if err != nil {
		logging.WARN("解析CQ消息失败: ", err.Error())
		return nil
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("GET")

	req.SetRequestURI("http://" + c.miraiAddr + "/memberList?sessionKey=" + c.sessionKey + "&target=" + strconv.Itoa(msg.GroupID))

	err = fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("向Mirai请求出错: ", err.Error())
		return nil
	}

	var rj MemberListResp

	err = json.Unmarshal(resp.Body(), &rj)
	if err != nil {
		logging.WARN("解析Mirai回复出错: ", err.Error())
		return nil
	}

	rs := new(cqResponse)
	var re []cqMemberList
	for _, info := range rj {
		re = append(re, cqMemberList{
			UserID:   info.ID,
			GroupID:  info.Group.ID,
			Nickname: info.MemberName,
			Role:     formatPerm(info.Permission),
		})
	}
	rs.Data, err = json.Marshal(&re)
	if err != nil {
		logging.WARN("生成CQ回复出错: ", err.Error())
		return nil
	}
	rs.Retcode = 0
	rs.Status = "ok"
	return rs
}

func (c *CMiraiConn) getGroupList(params []byte) *cqResponse {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("GET")

	req.SetRequestURI("http://" + c.miraiAddr + "/groupList?sessionKey=" + c.sessionKey)

	err := fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("向Mirai请求出错: ", err.Error())
		return nil
	}

	var rj GroupListResp

	err = json.Unmarshal(resp.Body(), &rj)
	if err != nil {
		logging.WARN("解析Mirai回复出错: ", err.Error())
		return nil
	}

	rs := new(cqResponse)
	var re []cqGroupListResp
	for _, info := range rj {
		re = append(re, cqGroupListResp{
			GroupID:   info.ID,
			GroupName: info.Name,
		})
	}
	rs.Data, err = json.Marshal(&re)
	if err != nil {
		logging.WARN("生成CQ回复出错: ", err.Error())
		return nil
	}
	rs.Retcode = 0
	rs.Status = "ok"
	return rs
}
