package models

import (
	"time"

	"github.com/1046102779/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

type SystemMessageTemplates struct {
	Id                  int       `orm:"column(system_message_template_id);auto"`
	Code                string    `orm:"column(code);size(20);null"`
	Title               string    `orm:"column(title);size(100);null"`
	Content             string    `orm:"column(content);null"`
	IndustryCodeQueryId int       `orm:"column(industry_code_query_id);null"`
	Status              int16     `orm:"column(status);null"`
	UpdatedAt           time.Time `orm:"column(updated_at);type(datetime);null"`
	CreatedAt           time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *SystemMessageTemplates) TableName() string {
	return "system_message_templates"
}

func (t *SystemMessageTemplates) GetSystemMessageTemplateNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter GetSystemMessageTemplateNoLock.", t.Id)
	defer Logger.Info("[%v] left GetSystemMessageTemplateNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		return
	}
	if t.Id <= 0 {
		err = errors.New("param `id` illegal")
		return
	}
	if err = (*o).Read(t); err != nil {
		err = errors.Wrap(err, "GetSystemMessageTemplateNoLock")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	return
}

func (t *SystemMessageTemplates) UpdateSystemMessageTemplatesNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter SystemMessageTemplates.", t.Id)
	defer Logger.Info("[%v] left SystemMessageTemplates.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		return
	}
	if _, err = (*o).Update(t); err != nil {
		err = errors.Wrap(err, "UpdateSystemMessageTemplatesNoLock")
		retcode = consts.ERROR_CODE__DB__UPDATE
		return
	}
	return
}

func init() {
	orm.RegisterModel(new(SystemMessageTemplates))
}

func GetSystemMessageTemplatesNoLock(pageIndex int, pageSize int, o *orm.Ormer) (messageTemplates []SystemMessageTemplates, retcode int, err error) {
	Logger.Info("enter GetSystemMessageTemplatesNoLock.")
	defer Logger.Info("left GetSystemMessageTemplatesNoLock.")
	messageTemplates = []SystemMessageTemplates{}
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	_, err = (*o).QueryTable((&SystemMessageTemplates{}).TableName()).Filter("status", consts.STATUS_VALID).Limit(pageSize, pageIndex*pageSize).All(&messageTemplates)
	if err != nil {
		err = errors.Wrap(err, "GetSystemMessageTemplatesNoLock")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	return
}

func getTemplateIdByName(name string) (templateId int, retcode int, err error) {
	Logger.Info("[%v] enter getTemplateIdByName.", name)
	defer Logger.Info("[%v] left getTemplateIdByName.", name)
	var (
		templates []*SystemMessageTemplates
		num       int64 = 0
	)
	o := orm.NewOrm()
	num, err = o.QueryTable((&SystemMessageTemplates{}).TableName()).Filter("title", name).Filter("status", consts.STATUS_VALID).All(&templates)
	if err != nil {
		err = errors.Wrap(err, "getTemplateIdByName")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		templateId = templates[0].Id
	}
	return
}
