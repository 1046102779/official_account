package controllers

import (
	"encoding/json"
	"time"

	"github.com/1046102779/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
)

// AccountMessageTemplatesController oprations for AccountMessageTemplates
type AccountMessageTemplatesController struct {
	beego.Controller
}

// 删除公众号指定的消息模板
// @router /accounts/:id/message_templates/:message_id/invalid [PUT]
func (t *AccountMessageTemplatesController) LogicDeleteAccountMessageTemplate() {
	id, _ := t.GetInt(":id")
	messageId, _ := t.GetInt(":message_id")
	o := orm.NewOrm()
	now := time.Now()
	accountMessageTemplate := &models.AccountMessageTemplates{
		Id: messageId,
	}
	if retcode, err := accountMessageTemplate.ReadAccountMessageTemplateNoLock(&o); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if accountMessageTemplate.OfficialAccountId != id {
		err := errors.New("param `:id` hasn't include `:message_id`")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__SOURCE_DATA__ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	accountMessageTemplate.Status = consts.STATUS_DELETED
	accountMessageTemplate.UpdatedAt = now
	if retcode, err := accountMessageTemplate.LogicDeleteAccountMessageTemplateNoLock(&o); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	t.Data["json"] = map[string]interface{}{
		"err_code": 0,
		"err_msg":  "",
	}
	t.ServeJSON()
	return
}

// 给公众号添加消息模板
// @router /:id/message_template [POST]
func (t *AccountMessageTemplatesController) AddAccountMessageTemplate() {
	type MessageTemplateInfo struct {
		Id int `json:"id"`
	}
	var (
		messageTemplateInfo *MessageTemplateInfo = new(MessageTemplateInfo)
	)
	id, _ := t.GetInt(":id")
	if err := json.Unmarshal(t.Ctx.Input.RequestBody, messageTemplateInfo); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__JSON__PARSE_FAILED,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if messageTemplate, retcode, err := models.AddAccountMessageTemplate(id, messageTemplateInfo.Id); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	} else {
		t.Data["json"] = map[string]interface{}{
			"err_code":              0,
			"err_msg":               "",
			"message_template_info": *messageTemplate,
		}
	}
	t.ServeJSON()
	return
}

// 获取公众号的消息模板列表
// @router /:id/message_templates [GET]
func (t *AccountMessageTemplatesController) GetAccountMessageTemplates() {
	type MessageTemplateInfo struct {
		Id         int    `json:"id"`
		TemplateId string `json:"template_id"`
		Content    string `json:"content"`
	}
	var (
		messageTemplateInfos  []MessageTemplateInfo = []MessageTemplateInfo{}
		systemMessageTemplate *models.SystemMessageTemplates
	)
	id, _ := t.GetInt(":id")
	pageIndex, _ := t.GetInt("page_index")
	pageSize, _ := t.GetInt("page_size")
	o := orm.NewOrm()
	accountMessageTemplates, count, realCount, retcode, err := models.GetAccountMessageTemplates(id, pageIndex, pageSize, &o)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	for index := 0; index < int(realCount); index++ {
		systemMessageTemplate = &models.SystemMessageTemplates{
			Id: accountMessageTemplates[index].SystemMessageTemplateId,
		}
		if retcode, err = systemMessageTemplate.GetSystemMessageTemplateNoLock(&o); err != nil {
			Logger.Error(err.Error())
			t.Data["json"] = map[string]interface{}{
				"err_code": retcode,
				"err_msg":  errors.Cause(err).Error(),
			}
			t.ServeJSON()
			return
		}
		messageTemplateInfos = append(messageTemplateInfos, MessageTemplateInfo{
			Id:         accountMessageTemplates[index].Id,
			TemplateId: accountMessageTemplates[index].TemplateId,
			Content:    systemMessageTemplate.Content,
		})
	}
	t.Data["json"] = map[string]interface{}{
		"err_code":               0,
		"err_msg":                "",
		"count":                  count,
		"real_count":             realCount,
		"message_template_infos": messageTemplateInfos,
	}
	t.ServeJSON()
	return
}
