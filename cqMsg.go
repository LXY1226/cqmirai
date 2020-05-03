package main

import (
	"bytes"
	"encoding/base64"
	"gitee.com/LXY1226/logging"
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

func (c *CMiraiWSRConn) uploadImage(b64Img []byte, imgType string) []byte {
	img, err := base64.StdEncoding.DecodeString(string(b64Img))
	if err != nil {
		logging.WARN("反Base64出错: ", err.Error())
		return nil
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
	w.WriteField("type", imgType)
	fw, err := w.CreateFormFile("img", "1d595495-0580-49ec-b96c-cc3346096718")
	fw.Write(img)
	req.SetBody(buf.Bytes())
	err = fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("上传图片出错: ", err.Error())
	}
	parser := parserPool.Get()
	defer parserPool.Put(parser)
	j, err := parser.ParseBytes(resp.Body())
	if err != nil {
		logging.WARN("上传图片解析出错: ", err.Error())
		return nil
	}
	return j.GetStringBytes("imageId")
}

func (c *CMiraiWSRConn) sendMsg(j *fastjson.Value, rawMsg []byte) []byte {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI("http://" + c.miraiConn.miraiAddr + "/sendGroupMessage")
	req.Header.SetMethod("POST")
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	var buf bytes.Buffer
	buf.WriteString("{\"sessionKey\":\"")
	buf.WriteString(c.miraiConn.sessionKey)
	buf.WriteString("\",\"target\":")
	buf.WriteString(strconv.Itoa(j.GetInt("params", "group_id")))
	buf.WriteString(",\"messageChain\":[")
	msgs := j.Get("params", "message")
	switch msgs.Type() {
	case fastjson.TypeArray:
		for _, msg := range msgs.GetArray() {
			switch string(msg.GetStringBytes("type")) {
			case "at":
				buf.WriteString("{\"type\":\"At\",\"target\":")
				buf.WriteString(strconv.Itoa(msg.GetInt("data", "qq")))
				buf.WriteString("},")
			case "text":
				buf.WriteString("{\"type\":\"Plain\",\"text\":\"")
				buf.WriteString(strconv.QuoteToASCII(string(msgs.GetStringBytes())))
				buf.WriteString("\"},")
			case "image":
				if id := c.uploadImage(msg.GetStringBytes("data", "file"), "group"); id != nil {
					buf.WriteString("{\"type\":\"Image\",\"imageId\":\"")
					buf.Write(id)
					buf.WriteString("\"},")
				}
			}
		}
	case fastjson.TypeString:
		buf.WriteString("{\"type\":\"Plain\",\"text\":")
		buf.WriteString(strconv.QuoteToASCII(string(msgs.GetStringBytes())))
		buf.WriteString("},")
	}
	echo := j.Get("echo").MarshalTo(nil)
	buf.Truncate(buf.Len() - 1)
	buf.WriteString("]}")
	req.SetBody(buf.Bytes())
	err := fasthttp.Do(req, resp)
	if err != nil {
		logging.WARN("向Mirai请求出错: ", err.Error())
		return nil
	}
	parser := parserPool.Get()
	defer parserPool.Put(parser)
	j, err = parser.ParseBytes(resp.Body())
	if err != nil {
		logging.WARN("解析Mirai回复出错: ", err.Error())
		return nil
	}
	buf.Reset()
	buf.WriteString("{\"data\":{\"message_id\":")
	buf.WriteString(strconv.Itoa(j.GetInt("messageId")))
	buf.WriteString(",\"echo\":")
	buf.Write(echo)
	buf.WriteString(",\"retcode\":0,\"status\":\"ok\"")

	return buf.Bytes()
}

// TODO 尚未实现
func (c *CMiraiWSRConn) get_group_member_info(j *fastjson.Value) []byte {
	echo := j.Get("echo").MarshalTo(nil)
	return append([]byte("{\"data\":{},"), append(echo, "\"retcode\":0,\"status\":\"ok\"}"...)...)
}
