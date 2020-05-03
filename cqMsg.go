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

func (c *CMiraiWSRConn) uploadImage(dataURI string, imgTarget string) string {
	img, err := base64.StdEncoding.DecodeString(dataURI[9:])
	if err != nil {
		logging.WARN("反Base64出错: ", err.Error())
		return ""
	}
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI("http://" + c.miraiConn.miraiAddr + "/uploadImage")
	req.Header.SetMethod("POST")
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	w.WriteField("sessionKey", c.miraiConn.sessionKey)
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

func quoteASCII(msg string) []byte {
	return []byte(strconv.QuoteToASCII(msg))
}

func (c *CMiraiWSRConn) formMsgChain(msg jsoniter.Any, imgTarget string) []MessageChain {
	switch msg.ValueType() {
	case jsoniter.StringValue:
		return []MessageChain{{
			Type: "Plain",
			Text: quoteASCII(msg.ToString()),
		}}
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
			return []MessageChain{{
				Type: "Plain",
				Text: quoteASCII(m.Data.Get("text").ToString()),
			}}
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
				cs = append(cs, MessageChain{
					Type: "Plain",
					Text: quoteASCII(msg.Data.Get("text").ToString()),
				})
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

func (c *CMiraiWSRConn) sendMsg(params string) *cqResponse {
	msg := new(cqMessage)
	err := json.UnmarshalFromString(params, msg)
	if err != nil {
		logging.WARN("解析CQ消息失败: ", err.Error())
		return nil
	}
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("POST")
	req.Header.Add("Content-Type", "application/json; charset=gbk")
	miraiMsg := new(outgoingMessage)
	miraiMsg.SessionKey = c.miraiConn.sessionKey
	imgTarget := "friend"
	switch msg.MessageType {
	case "group":
		miraiMsg.Target = msg.GroupID
		imgTarget = "group"
		req.SetRequestURI("http://" + c.miraiConn.miraiAddr + "/sendGroupMessage")
	case "private":
		miraiMsg.Target = msg.UserID
		imgTarget = "friend"
		req.SetRequestURI("http://" + c.miraiConn.miraiAddr + "/sendFriendMessage")
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
	rj := new(MessageResponse)
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

// TODO 尚未实现
func (c *CMiraiWSRConn) getGroupMemberInfo(params string) *cqResponse {
	msg := new(cqGroupMemberInfoReq)
	err := json.UnmarshalFromString(params, msg)
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
