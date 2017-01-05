package controllers

import (
	"time"

	utils "github.com/1046102779/common"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
)

// MessageTemplatesController oprations for MessageTemplates
type MessageTemplatesController struct {
	beego.Controller
}

type MessageTemplateInfo struct {
	Id            int    `json:"id"`
	Code          string `json:"code"`
	Title         string `json:"title"`
	FirstIndustry string `json:"first_industry"`
	SecIndustry   string `json:"sec_industry"`
	Content       string `json:"content"`
}

// 获取系统中已添加的微信消息模板列表
// @router /system/message_templates [GET]
func (t *MessageTemplatesController) GetMessageTemplates() {
	type MessageTemplateInfo struct {
		Id              int    `json:"id"`
		Code            string `json:"code"`
		Content         string `json:"content"`
		Title           string `json:"title"`
		PrimaryIndustry string `json:"primary_industry"`
		DeputyIndustry  string `json:"deputy_industry"`
	}
	var (
		messageTemplateInfos []MessageTemplateInfo = []MessageTemplateInfo{}
		industryCodeQuery    *models.IndustryCodeQuerys
		realCount            int64
	)
	pageIndex, _ := t.GetInt("page_index")
	pageSize, _ := t.GetInt("page_size")
	o := orm.NewOrm()
	count, err := o.QueryTable((&models.SystemMessageTemplates{}).TableName()).Filter("status", utils.STATUS_VALID).Count()
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": utils.DB_READ_ERROR,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	messageTemplates, retcode, err := models.GetSystemMessageTemplatesNoLock(pageSize, pageIndex*pageSize, &o)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if messageTemplates != nil && len(messageTemplates) > 0 {
		realCount = int64(len(messageTemplates))
	}
	for index := 0; messageTemplates != nil && index < len(messageTemplates); index++ {
		industryCodeQuery = &models.IndustryCodeQuerys{
			Id: messageTemplates[index].IndustryCodeQueryId,
		}
		if retcode, err = industryCodeQuery.GetIndustryCodeQueryNoLock(&o); err != nil {
			Logger.Error(err.Error())
			return
		}
		messageTemplateInfos = append(messageTemplateInfos, MessageTemplateInfo{
			Id:              messageTemplates[index].Id,
			Code:            messageTemplates[index].Code,
			Title:           messageTemplates[index].Title,
			PrimaryIndustry: industryCodeQuery.MainIndustryCode,
			DeputyIndustry:  industryCodeQuery.SecIndustryCode,
			Content:         messageTemplates[index].Content,
		})
	}
	t.Data["json"] = map[string]interface{}{
		"err_code":               0,
		"err_msg":                "",
		"message_template_infos": messageTemplateInfos,
		"real_count":             realCount,
		"count":                  count,
	}
	t.ServeJSON()
	return
}

// 从系统中删除指定的消息模板
// @router /system/message_templates/:id/invalid [PUT]
func (t *MessageTemplatesController) LogicDeleteSystemMessageTemplate() {
	id, _ := t.GetInt(":id")
	now := time.Now()
	template := &models.SystemMessageTemplates{
		Id: id,
	}
	o := orm.NewOrm()
	if retcode, err := template.GetSystemMessageTemplateNoLock(&o); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if template.Code != "" {
		template.Status = utils.STATUS_DELETED
		template.UpdatedAt = now
		if retcode, err := template.UpdateSystemMessageTemplatesNoLock(&o); err != nil {
			Logger.Error(err.Error())
			t.Data["json"] = map[string]interface{}{
				"err_code": retcode,
				"err_msg":  errors.Cause(err).Error(),
			}
			t.ServeJSON()
			return
		}
		// 删除系统中公众号已使用过的该消息模板ID
		if retcode, err := models.LogicDeleteAccountMessageTemplatesNoLocks(template.Id, &o); err != nil {
			Logger.Error(err.Error())
			t.Data["json"] = map[string]interface{}{
				"err_code": retcode,
				"err_msg":  errors.Cause(err).Error(),
			}
			t.ServeJSON()
			return
		}
	}
	t.Data["json"] = map[string]interface{}{
		"err_code": 0,
		"err_msg":  "",
	}
	t.ServeJSON()
	return
}
