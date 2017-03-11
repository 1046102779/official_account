package models

import (
	"fmt"
	"time"

	"github.com/1046102779/common/consts"
	pb "github.com/1046102779/igrpc"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

type AccountMessageTemplates struct {
	Id                      int       `orm:"column(account_message_template_id);auto" json:"id"`
	OfficialAccountId       int       `orm:"column(official_account_id);null" json:"official_account_id"`
	TemplateId              string    `orm:"column(template_id);null" json:"template_id"`
	SystemMessageTemplateId int       `orm:"column(system_message_template_id);null" json:"system_message_template_id"`
	Status                  int16     `orm:"column(status);null"`
	UpdatedAt               time.Time `orm:"column(updated_at);type(datetime);null"`
	CreatedAt               time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *AccountMessageTemplates) TableName() string {
	return "account_message_templates"
}

func (t *AccountMessageTemplates) ReadAccountMessageTemplateNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter ReadAccountMessageTemplateNoLock.", t.Id)
	defer Logger.Info("[%v] left ReadAccountMessageTemplateNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		return
	}
	if err = (*o).Read(t); err != nil {
		err = errors.Wrap(err, "ReadAccountMessageTemplateNoLock")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	return
}

func (t *AccountMessageTemplates) UpdateAccountMessageTemplateNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter UpdateAccountMessageTemplateNoLock.", t.Id)
	defer Logger.Info("[%v] left UpdateAccountMessageTemplateNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		return
	}
	if _, err = (*o).Update(t); err != nil {
		err = errors.Wrap(err, "UpdateAccountMessageTemplateNoLock")
		retcode = consts.ERROR_CODE__DB__UPDATE
		return
	}
	return
}

func (t *AccountMessageTemplates) LogicDeleteAccountMessageTemplateNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter LogicDeleteAccountMessageTemplateNoLock.", t.Id)
	defer Logger.Info("[%v] left LogicDeleteAccountMessageTemplateNoLock.", t.Id)
	var (
		weMessageTemplate *WeMessageTemplate = new(WeMessageTemplate)
		token             string
	)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		return
	}
	if retcode, err = t.UpdateAccountMessageTemplateNoLock(o); err != nil {
		err = errors.Wrap(err, "LogicDeleteAccountMessageTemplateNoLock")
		return
	}
	token, retcode, err = GetAuthorierAccessTokenById(t.OfficialAccountId, o)
	if err != nil {
		err = errors.Wrap(err, "LogicDeleteAccountMessageTemplateNoLock")
		return
	}
	if token != "" {
		retcode, err = weMessageTemplate.DeleteMessageTemplate(token, t.TemplateId)
		if err != nil {
			err = errors.Wrap(err, "LogicDeleteAccountMessageTemplateNoLock")
			return
		}
	}
	return
}

func (t *AccountMessageTemplates) InsertAccountMessageTemplate(o *orm.Ormer) (retcode int, err error) {
	Logger.Error("[%v] enter InsertAccountMessageTemplate.", t.OfficialAccountId)
	defer Logger.Error("[%v] left InsertAccountMessageTemplate.", t.OfficialAccountId)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		err = errors.Wrap(err, "InsertAccountMessageTemplate")
		retcode = consts.ERROR_CODE__DB__INSERT
		return
	}
	return
}

func init() {
	orm.RegisterModel(new(AccountMessageTemplates))
}

func GetAccountMessageTemplates(id int, pageIndex int, pageSize int, o *orm.Ormer) (accountMessageTemplates []AccountMessageTemplates, count int64, realCount int64, retcode int, err error) {
	Logger.Info("[%v] enter GetAccountMessageTemplates.", id)
	defer Logger.Info("[%v] left GetAccountMessageTemplates.", id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		return
	}
	qs := (*o).QueryTable((&AccountMessageTemplates{}).TableName()).Filter("status", consts.STATUS_VALID)
	count, _ = qs.Count()
	realCount, err = qs.Limit(pageSize, pageSize*pageIndex).All(&accountMessageTemplates)
	if err != nil {
		err = errors.Wrap(err, "GetAccountMessageTemplates")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	return
}

func AddAccountMessageTemplate(id int, messageTemplateId int) (accountMessageTemplate *AccountMessageTemplates, retcode int, err error) {
	Logger.Info("[%v] enter AddAccountMessageTemplate.", id)
	defer Logger.Info("[%v] left AddAccountMessageTemplate.", id)
	var (
		accountMessageTemplates []AccountMessageTemplates = []AccountMessageTemplates{}
		weMessageTemplate       *WeMessageTemplate        = new(WeMessageTemplate)
		token                   string
		num                     int64
	)
	accountMessageTemplate = new(AccountMessageTemplates)
	now := time.Now()
	if id <= 0 || messageTemplateId <= 0 {
		err = errors.New("param `id || message_template_id` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	// ::TODO  需要判断该公众号的消息模板数量是否达到25个，这个是微信的申请上限
	o := orm.NewOrm()
	num, err = o.QueryTable(accountMessageTemplate.TableName()).Filter("official_account_id", id).Filter("system_message_template_id", messageTemplateId).All(&accountMessageTemplates)
	if err != nil {
		err = errors.Wrap(err, "AddAccountMessageTemplate")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		if accountMessageTemplates[0].Status == consts.STATUS_VALID {
			// 有效
			accountMessageTemplate = &accountMessageTemplates[0]
			return
		}
	}

	// 获取authorizer_access_token
	token, retcode, err = GetAuthorierAccessTokenById(id, &o)
	if err != nil {
		err = errors.Wrap(err, "AddAccountMessageTemplate")
		return
	}
	// 根据message_template_id 获取消息模板编号code
	systemMessageTemplate := &SystemMessageTemplates{
		Id: messageTemplateId,
	}
	if retcode, err = systemMessageTemplate.GetSystemMessageTemplateNoLock(&o); err != nil {
		err = errors.Wrap(err, "AddAccountMessageTemplate")
		return
	}
	// 根据消息模板编码, 从微信公众号获取消息模板ID
	accountMessageTemplate.TemplateId, retcode, err = weMessageTemplate.GetTemplateId(token, systemMessageTemplate.Code)
	if err != nil {
		err = errors.Wrap(err, "AddAccountMessageTemplate")
		return
	}
	if accountMessageTemplates != nil && len(accountMessageTemplates) > 0 {
		// 无效
		accountMessageTemplates[0].Status = consts.STATUS_VALID
		accountMessageTemplates[0].UpdatedAt = now
		if retcode, err = accountMessageTemplates[0].UpdateAccountMessageTemplateNoLock(&o); err != nil {
			err = errors.Wrap(err, "AddAccountMessageTemplate")
			return
		}
	} else {
		// 添加到公众号消息模板中
		accountMessageTemplate.OfficialAccountId = id
		accountMessageTemplate.SystemMessageTemplateId = messageTemplateId
		accountMessageTemplate.Status = consts.STATUS_VALID
		accountMessageTemplate.UpdatedAt = now
		accountMessageTemplate.CreatedAt = now
		if retcode, err = accountMessageTemplate.InsertAccountMessageTemplate(&o); err != nil {
			err = errors.Wrap(err, "AddAccountMessageTemplate")
			return
		}
	}
	return
}

// 删除系统中公众号已使用过的该消息模板ID
func LogicDeleteAccountMessageTemplatesNoLocks(systemMessageTemplateId int, o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter LogicDeleteAccountMessageTemplatesNoLocks.", systemMessageTemplateId)
	defer Logger.Info("[%v] left LogicDeleteAccountMessageTemplatesNoLocks.", systemMessageTemplateId)
	var (
		accountMessageTemplates []AccountMessageTemplates = []AccountMessageTemplates{}
		weMessageTemplate       *WeMessageTemplate        = new(WeMessageTemplate)
		offAcc                  *OfficialAccounts
		token                   string
		num                     int64
	)
	now := time.Now()
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		return
	}
	if systemMessageTemplateId <= 0 {
		err = errors.New("param `id` illegal")
		return
	}
	num, err = (*o).QueryTable((&AccountMessageTemplates{}).TableName()).Filter("system_message_template_id", systemMessageTemplateId).Filter("status", consts.STATUS_VALID).All(&accountMessageTemplates)
	if err != nil {
		retcode = consts.ERROR_CODE__DB__READ
		err = errors.Wrap(err, "LogicDeleteAccountMessageTemplatesNoLocks")
		return
	}
	for index := 0; index < int(num); index++ {
		accountMessageTemplates[index].Status = consts.STATUS_DELETED
		accountMessageTemplates[index].UpdatedAt = now
		if retcode, err = (&accountMessageTemplates[index]).UpdateAccountMessageTemplateNoLock(o); err != nil {
			err = errors.Wrap(err, "LogicDeleteAccountMessageTemplatesNoLocks")
			return
		}
	}
	// 删除微信公众号各个appid下的该模板消息ID
	for index := 0; index < int(num); index++ {
		offAcc = &OfficialAccounts{
			Id: accountMessageTemplates[index].OfficialAccountId,
		}
		if retcode, err = offAcc.ReadOfficialAccountNoLock(o); err != nil {
			err = errors.Wrap(err, "LogicDeleteAccountMessageTemplatesNoLocks")
			return
		}
		if offAcc.Appid != "" {
			in := &pb.OfficialAccount{Appid: offAcc.Appid}
			conf.WxRelayServerClient.Call(fmt.Sprintf("%s.%s", "wx_relay_server", "GetOfficialAccountInfo"), in, in)
			token = in.AuthorizerAccessToken
			retcode, err = weMessageTemplate.DeleteMessageTemplate(token, accountMessageTemplates[index].TemplateId)
			if err != nil {
				err = errors.Wrap(err, "LogicDeleteAccountMessageTemplatesNoLocks")
				return
			}
		}
	}
	return
}

// @params: companyId: 公司ID
// @params: name: 模板名称
func getTemplateIdByCompanyIdAndName(companyId int, name string) (templateId string, retcode int, err error) {
	// 获取系统模板的模板ID
	var (
		systemTemplateId  int // 平台系统消息模板ID
		officialAccountId int // 公众号ID
	)
	if systemTemplateId, retcode, err = getTemplateIdByName(name); err != nil {
		err = errors.Wrap(err, "getTemplateIdByCompanyIdAndName")
		return
	}
	if officialAccountId, retcode, err = getOfficialAccountIdByCompanyId(companyId); err != nil {
		err = errors.Wrap(err, "getTemplateIdByCompanyIdAndName")
		return
	}
	templateId, retcode, err = getTemplateIdBySystemTemplateIdAndOfficialAccountId(systemTemplateId, officialAccountId)
	if err != nil {
		err = errors.Wrap(err, "getTemplateIdByCompanyIdAndName")
		return
	}
	return
}

func getTemplateIdBySystemTemplateIdAndOfficialAccountId(systemTemplateId int, officialAccountId int) (templateId string, retcode int, err error) {
	Logger.Info("[%v.%v] enter getTemplateIdBySystemTemplateIdAndOfficialAccountId.", systemTemplateId, officialAccountId)
	defer Logger.Info("[%v.%v] left getTemplateIdBySystemTemplateIdAndOfficialAccountId.", systemTemplateId, officialAccountId)
	var (
		templates []*AccountMessageTemplates
		num       int64 = 0
	)
	o := orm.NewOrm()
	num, err = o.QueryTable((&AccountMessageTemplates{}).TableName()).Filter("official_account_id", officialAccountId).Filter("system_message_template_id", systemTemplateId).Filter("status", consts.STATUS_VALID).All(&templates)
	if err != nil {
		err = errors.Wrap(err, "getTemplateIdBySystemTemplateIdAndOfficialAccountId")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		templateId = templates[0].TemplateId
	}
	return
}
