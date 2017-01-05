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
		retcode = utils.DB_READ_ERROR
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
		retcode = utils.DB_UPDATE_ERROR
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
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	_, err = (*o).QueryTable((&SystemMessageTemplates{}).TableName()).Filter("status", utils.STATUS_VALID).Limit(pageSize, pageIndex*pageSize).All(&messageTemplates)
	if err != nil {
		err = errors.Wrap(err, "GetSystemMessageTemplatesNoLock")
		retcode = utils.DB_READ_ERROR
		return
	}
	return
}

// GetAllMessageTemplates retrieves all MessageTemplates matches certain condition. Returns empty list if
// no records exist
func GetAllSystemMessageTemplates(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(SystemMessageTemplates))
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

	var l []SystemMessageTemplates
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
