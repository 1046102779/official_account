package models

import (
	"reflect"
	"strings"
	"time"

	utils "github.com/1046102779/common"
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
		retcode = utils.DB_READ_ERROR
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
		retcode = utils.DB_UPDATE_ERROR
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
		retcode = utils.DB_INSERT_ERROR
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
	qs := (*o).QueryTable((&AccountMessageTemplates{}).TableName()).Filter("status", utils.STATUS_VALID)
	count, _ = qs.Count()
	realCount, err = qs.Limit(pageSize, pageSize*pageIndex).All(&accountMessageTemplates)
	if err != nil {
		err = errors.Wrap(err, "GetAccountMessageTemplates")
		retcode = utils.DB_READ_ERROR
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
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	// ::TODO  需要判断该公众号的消息模板数量是否达到25个，这个是微信的申请上限
	o := orm.NewOrm()
	num, err = o.QueryTable(accountMessageTemplate.TableName()).Filter("official_account_id", id).Filter("system_message_template_id", messageTemplateId).All(&accountMessageTemplates)
	if err != nil {
		err = errors.Wrap(err, "AddAccountMessageTemplate")
		retcode = utils.DB_READ_ERROR
		return
	}
	if num > 0 {
		if accountMessageTemplates[0].Status == utils.STATUS_VALID {
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
		accountMessageTemplates[0].Status = utils.STATUS_VALID
		accountMessageTemplates[0].UpdatedAt = now
		if retcode, err = accountMessageTemplates[0].UpdateAccountMessageTemplateNoLock(&o); err != nil {
			err = errors.Wrap(err, "AddAccountMessageTemplate")
			return
		}
	} else {
		// 添加到公众号消息模板中
		accountMessageTemplate.OfficialAccountId = id
		accountMessageTemplate.SystemMessageTemplateId = messageTemplateId
		accountMessageTemplate.Status = utils.STATUS_VALID
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
	num, err = (*o).QueryTable((&AccountMessageTemplates{}).TableName()).Filter("system_message_template_id", systemMessageTemplateId).Filter("status", utils.STATUS_VALID).All(&accountMessageTemplates)
	if err != nil {
		retcode = utils.DB_READ_ERROR
		err = errors.Wrap(err, "LogicDeleteAccountMessageTemplatesNoLocks")
		return
	}
	for index := 0; index < int(num); index++ {
		accountMessageTemplates[index].Status = utils.STATUS_DELETED
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
			if _, ok := conf.WechatAuthTTL.AuthorizerMap[offAcc.Appid]; !ok {
				Logger.Error("param `appid` not exists in maps")
				continue
			}
			token = conf.WechatAuthTTL.AuthorizerMap[offAcc.Appid].AuthorizerAccessToken
			retcode, err = weMessageTemplate.DeleteMessageTemplate(token, accountMessageTemplates[index].TemplateId)
			if err != nil {
				err = errors.Wrap(err, "LogicDeleteAccountMessageTemplatesNoLocks")
				return
			}
		}
	}
	return
}

// GetAllAccountMessageTemplates retrieves all AccountMessageTemplates matches certain condition. Returns empty list if
// no records exist
func GetAllAccountMessageTemplates(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(AccountMessageTemplates))
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Filter(k, v)
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields")
		}
	}

	var l []AccountMessageTemplates
	qs = qs.OrderBy(sortFields...)
	if _, err = qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(fname).Interface()
				}
				ml = append(ml, m)
			}
		}
		return ml, nil
	}
	return nil, err
}
