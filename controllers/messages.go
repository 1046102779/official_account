package controllers

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	pb "github.com/1046102779/official_account/igrpc"

	"github.com/1046102779/official_account/common/consts"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/chanxuehong/util"
	"github.com/gomydodo/wxencrypter"
)

type MessageController struct {
	beego.Controller
}

type WechatMessageReq struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Event        string `xml:"Event"`
	Content      string `xml:"Content"`
	MsgID        string `xml:"MsgID"`
	Status       string `xml:"Status"`
}

func GlobalPublishTestGetContent(reqMap map[string]string) (content string) {
	Logger.Info("enter GlobalPublishTest.")
	defer Logger.Info("left GlobalPublishTest.")
	switch reqMap["MsgType"] {
	case "event":
		content = fmt.Sprintf("%s%s", reqMap["Event"], "from_callback")
	case "text":
		if strings.HasPrefix(reqMap["Content"], "QUERY_AUTH_CODE") {
			content = ""
		} else {
			content = fmt.Sprintf("%s%s", reqMap["Content"], "_callback")
		}
	}
	return
}

var textBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 64<<10)) // 默认 64KB
	},
}

// 公众号消息与事件接收URL
// @router /:appid/callback [POST]
func (t *MessageController) WechatCallback() {
	type WechatMessageTextReq struct {
		ToUserName   string `xml:"ToUserName"`
		FromUserName string `xml:"FromUserName"`
		CreateTime   int64  `xml:"CreateTime"`
		MsgType      string `xml:"MsgType"`
		Content      string `xml:"Content"`
	}
	now := time.Now()
	appid := t.GetString(":appid")
	timestamp := t.GetString("timestamp")
	nonce := t.GetString("nonce")
	msgSignature := t.GetString("msg_signature")
	in := &pb.OfficialAccountPlatform{}
	conf.WxRelayServerClient.Call(fmt.Sprintf("%s.%s", "wx_relay_server", "GetOfficialAccountPlatformInfo"), in, in)
	e, err := wxencrypter.NewEncrypter(in.Token, in.EncodingAesKey, in.Appid)
	if err != nil {
		Logger.Error("NewEncrypter failed. " + err.Error())
	}
	b, err := e.Decrypt(msgSignature, timestamp, nonce, t.Ctx.Input.RequestBody)
	if err != nil {
		Logger.Error(err.Error())
	}
	reader := strings.NewReader(string(b))
	reqMap, err := util.DecodeXMLToMap(reader)
	if err != nil {
		Logger.Error(err.Error())
	}
	// 全网发布测试代码
	if appid == "wx570bc396a51b8ff8" {
		reqMap["Content"] = GlobalPublishTestGetContent(reqMap)
		fmt.Println("content: ", reqMap["Content"])
		service := new(models.CustomerServices)
		if reqMap["Content"] == "" {
			t.Ctx.Output.Body([]byte(""))
			go func() {
				content := fmt.Sprintf("%s%s", conf.QueryAuthCodeTest, "_from_api")
				service.SendMessage(appid, reqMap["FromUserName"], reqMap["MsgType"], content)
			}()
			return
		} else {
			t.Ctx.Output.Body([]byte(""))
			go service.SendMessage(appid, reqMap["FromUserName"], "text", reqMap["Content"])
			return
		}
	}
	// 测试代码结束
	o := orm.NewOrm()
	messageReceiptRecord := &models.WechatMessageReceiptRecords{
		Appid:        appid,
		ToUserName:   reqMap["ToUserName"],
		FromUserName: reqMap["FromUserName"],
		CreateTime:   now,
		MsgType:      reqMap["MsgType"],
		Event:        reqMap["Event"],
		Content:      reqMap["Content"],
		MsgId:        reqMap["MsgID"],
		Status:       consts.STATUS_VALID,
		CreatedAt:    now,
	}
	if _, err := messageReceiptRecord.InsertWechatMessageReceiptRecordNoLock(&o); err != nil {
		Logger.Error(err.Error())
	}
	t.Ctx.Output.Body([]byte("success"))
	t.ServeXML()
	return
}
