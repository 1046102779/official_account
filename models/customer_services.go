package models

import (
	"encoding/json"
	"fmt"

	"github.com/1046102779/common/httpRequest"
	"github.com/1046102779/common/types"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"
)

type CustomerServices struct{}

type Text struct {
	Content string `json:"content"`
}
type MessageInfo struct {
	ToUser  string `json:"touser"`
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

// 客服接口-发消息
func (t *CustomerServices) SendMessage(appid string, touser string, msgType string, content string) {
	Logger.Info("[%v] enter SendMessage.", touser)
	defer Logger.Info("[%v] left SendMessage.", touser)
	var err error
	message := &MessageInfo{
		ToUser:  touser,
		MsgType: msgType,
		Text: Text{
			Content: content,
		},
	}
	fmt.Println("send message info: ", *message)
	var oa *types.OfficialAccount
	if oa, err = conf.WRServerRPC.GetOfficialAccountInfo(appid); err != nil {
		err = errors.Wrap(err, "SendMessage")
		return
	}
	if oa.AuthorizerAccessToken != "" {
		httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s", oa.AuthorizerAccessToken)
		bodyData, _ := json.Marshal(*message)
		retBody, _ := httpRequest.HttpPostBody(httpStr, bodyData)
		fmt.Println("result: ", string(retBody))
		return
	}
	return
}
