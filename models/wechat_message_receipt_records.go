package models

import (
	"reflect"
	"strings"
	"time"

	utils "github.com/1046102779/common"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

type WechatMessageReceiptRecords struct {
	Id           int       `orm:"column(wechat_message_receipt_record_id);auto"`
	Appid        string    `orm:"column(appid);size(50);null"`
	ToUserName   string    `orm:"column(to_user_name);size(50);null"`
	FromUserName string    `orm:"column(from_user_name);size(50);null"`
	CreateTime   time.Time `orm:"column(create_time);type(datetime);null"`
	MsgType      string    `orm:"column(msg_type);size(50);null"`
	Event        string    `orm:"column(event);size(50);null"`
	Content      string    `orm:"column(content);size(1000);null"`
	MsgId        string    `orm:"column(msg_id);size(50);null"`
	Status       int16     `orm:"column(status);size(50);null"`
	CreatedAt    time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *WechatMessageReceiptRecords) TableName() string {
	return "wechat_message_receipt_records"
}

func (t *WechatMessageReceiptRecords) InsertWechatMessageReceiptRecordNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter WechatMessageReceiptRecordsNoLock", t.Appid)
	defer Logger.Info("[%v] left WechatMessageReceiptRecordsNoLock", t.Appid)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		err = errors.Wrap(err, "InsertWechatMessageReceiptRecordNoLock")
		retcode = utils.DB_INSERT_ERROR
		return
	}
	return
}

func init() {
	orm.RegisterModel(new(WechatMessageReceiptRecords))
}

// GetAllWechatMessageReceiptRecords retrieves all WechatMessageReceiptRecords matches certain condition. Returns empty list if
// no records exist
func GetAllWechatMessageReceiptRecords(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(WechatMessageReceiptRecords))
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

	var l []WechatMessageReceiptRecords
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
