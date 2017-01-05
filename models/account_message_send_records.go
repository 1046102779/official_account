package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	utils "github.com/1046102779/common"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

type AccountMessageSendRecords struct {
	Id                int       `orm:"column(account_message_send_record_id);auto"`
	OfficialAccountId int       `orm:"column(official_account_id);null"`
	TemplateId        string    `orm:"column(template_id);size(100);null"`
	Content           string    `orm:"column(content);size(2000);null"`
	ReceiverUser      string    `orm:"column(receiver_user);size(100);null"`
	Status            int16     `orm:"column(status);null"`
	UpdatedAt         time.Time `orm:"column(updated_at);type(datetime);null"`
	CreatedAt         time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *AccountMessageSendRecords) TableName() string {
	return "account_message_send_records"
}

func (t *AccountMessageSendRecords) InsertAccountMessageSendRecordNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter InsertAccountMessageSendRecordNoLock.", t.OfficialAccountId)
	defer Logger.Info("[%v] left InsertAccountMessageSendRecordNoLock.", t.OfficialAccountId)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		retcode = utils.DB_INSERT_ERROR
		return
	}
	return
}

func init() {
	orm.RegisterModel(new(AccountMessageSendRecords))
}

func SendAccountMessage(id int, toUser string, templateId string, content json.RawMessage) (msgId string, retcode int, err error) {
	var (
		accountMessageTemplates []AccountMessageTemplates = []AccountMessageTemplates{}
		weMessageTemplate       *WeMessageTemplate        = new(WeMessageTemplate)
		token                   string
		num                     int64
	)
	now := time.Now()
	o := orm.NewOrm()
	num, err = o.QueryTable((&AccountMessageTemplates{}).TableName()).Filter("official_account_id", id).Filter("template_id", templateId).All(&accountMessageTemplates)
	if err != nil {
		retcode = utils.DB_READ_ERROR
		return
	}
	if num <= 0 {
		err = errors.New("the official account hasn't this message_template , please firstly apply template")
		retcode = utils.WECHAT_MESSAGE_NOT_EXIST
		return
	}
	// 获取access_token
	token, retcode, err = GetAuthorierAccessTokenById(id, &o)
	if err != nil {
		err = errors.Wrap(err, "SendAccountMessage ")
		return
	}
	msgId, retcode, err = weMessageTemplate.SendMessage(token, templateId, content, toUser)
	if err != nil {
		err = errors.Wrap(err, "SendAccountMessage ")
		return
	}
	// 增加消息发送记录
	messageRecord := &AccountMessageSendRecords{
		OfficialAccountId: id,
		TemplateId:        templateId,
		Content:           string(content),
		ReceiverUser:      toUser,
		Status:            utils.STATUS_VALID,
		UpdatedAt:         now,
		CreatedAt:         now,
	}
	if retcode, err = messageRecord.InsertAccountMessageSendRecordNoLock(&o); err != nil {
		err = errors.Wrap(err, "SendAccountMessage ")
		return
	}
	return
}

// GetAllAccountMessageSendRecords retrieves all AccountMessageSendRecords matches certain condition. Returns empty list if
// no records exist
func GetAllAccountMessageSendRecords(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(AccountMessageSendRecords))
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

	var l []AccountMessageSendRecords
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
