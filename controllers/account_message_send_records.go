package controllers

import (
	"encoding/json"

	"github.com/1046102779/official_account/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/pkg/errors"
)

// AccountMessageSendRecordsController oprations for AccountMessageSendRecords
type AccountMessageSendRecordsController struct {
	beego.Controller
}

// 发送微信模板消息
// @router /:id/message [POST]
func (t *AccountMessageSendRecordsController) SendAccountMessage() {
	type MessageInfo struct {
		TemplateId string          `json:"template_id"`
		ToUser     string          `json:"touser"`
		Content    json.RawMessage `json:"content"`
	}
	var (
		messageInfo *MessageInfo = new(MessageInfo)
	)
	id, _ := t.GetInt(":id")
	if err := json.Unmarshal(t.Ctx.Input.RequestBody, messageInfo); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__JSON__PARSE_FAILED,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if msgId, retcode, err := models.SendAccountMessage(id, messageInfo.ToUser, messageInfo.TemplateId, messageInfo.Content); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	} else {
		t.Data["json"] = map[string]interface{}{
			"err_code": 0,
			"err_msg":  "",
			"msgid":    msgId,
		}
	}
	t.ServeJSON()
	return
}
