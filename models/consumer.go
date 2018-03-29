// kafka消息队列
package models

import (
	"encoding/json"
	"fmt"
	"sync"

	. "github.com/1046102779/official_account/logger"

	"github.com/pkg/errors"
)

var (
	wg sync.WaitGroup
)

type MessageInfoReq struct {
	TemplateId string          `json:"template_id"`
	ToUser     string          `json:"touser"`
	Content    json.RawMessage `json:"content"`
}

type Values struct {
	Value string `json:"value"`
	Color string `json:"color"`
}

// 通用模板消息内容
type Content struct {
	First    Values `json:"first"`
	Keyword1 Values `json:"keyword1"`
	Keyword2 Values `json:"keyword2"`
	Keyword3 Values `json:"keyword3"`
	Keyword4 Values `json:"keyword4"`
	Keyword5 Values `json:"keyword5"`
	Remark   Values `json:"remark"`
	//Keywords map[string]Values
}

// 封装消息为模板消息
// 输入：字段消息内容
// 输出：消息模板内容
func getTemplateMessageContent(first string, remarks string, keywords []string) (bts []byte, err error) {
	content := &Content{}
	content.First = Values{
		Value: first,
		Color: fmt.Sprintf("#173177"),
	}
	content.Remark = Values{
		Value: remarks,
		Color: fmt.Sprintf("#173177"),
	}
	//content.Keywords = make(map[string]Values, 0)
	for index := 0; keywords != nil && index < len(keywords); index++ {
		switch index + 1 {
		case 1:
			content.Keyword1 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		case 2:
			content.Keyword2 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		case 3:
			content.Keyword3 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		case 4:
			content.Keyword4 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		case 5:
			content.Keyword5 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		}
	}
	if bts, err = json.Marshal(content); err != nil {
		err = errors.Wrap(err, "getTemplateMessageContent")
		return
	}
	return
}

func ConsumeAccountTemplateMessage(officialAccountId int, bts []byte) (err error) {
	Logger.Info("[%v] enter ConsumeAccountTemplateMessage.", string(bts))
	defer Logger.Info("[%v] left ConsumeAccountTemplateMessage.", string(bts))
	var (
		msgId string
	)

	var (
		messageInfo *MessageInfoReq = new(MessageInfoReq)
	)
	if err = json.Unmarshal(bts, messageInfo); err != nil {
		Logger.Error(err.Error())
		return
	}
	if msgId, _, err = SendAccountMessage(officialAccountId, messageInfo.ToUser, messageInfo.TemplateId, messageInfo.Content); err != nil {
		Logger.Error(err.Error())
		return
	}
	fmt.Println("msg_id: ", msgId)
	return
}
