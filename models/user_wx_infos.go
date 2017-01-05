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

type UserWxInfos struct {
	Id         int       `orm:"column(user_wx_info_id);auto"`
	UserId     int       `orm:"column(user_id);null"`
	Nickname   string    `orm:"column(nickname);size(50);null"`
	Sex        int16     `orm:"column(sex);null"`
	Province   string    `orm:"column(province);size(100);null"`
	City       string    `orm:"column(city);size(100);null"`
	Country    string    `orm:"column(country);size(30);null"`
	Headimgurl string    `orm:"column(headimgurl);size(100);null"`
	Privilege  string    `orm:"column(privilege);size(100);null"`
	CreatedAt  time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *UserWxInfos) TableName() string {
	return "user_wx_infos"
}

func (t *UserWxInfos) InsertUserWxInfoNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter InsertUserWxInfoNoLock.")
	defer Logger.Info("[%v] left InsertUserWxInfoNoLock.")
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = utils.DB_INSERT_ERROR
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		err = errors.Wrap(err, "InsertUserWxInfoNoLock")
		retcode = utils.DB_INSERT_ERROR
		return
	}
	return
}

func init() {
	orm.RegisterModel(new(UserWxInfos))
}

// 通过用户UserId，获取主键ID
func GetUserWxInfoByUserId(userId int) (info *UserWxInfos, retcode int, err error) {
	Logger.Info("[%v] enter GetUserWxInfoByUserId.", userId)
	defer Logger.Info("[%v] left GetUserWxInfoByUserId.", userId)
	var (
		userWxInfos []UserWxInfos = []UserWxInfos{}
		num         int64
	)
	o := orm.NewOrm()
	num, err = o.QueryTable((&UserWxInfos{}).TableName()).Filter("user_id", userId).All(&userWxInfos)
	if err != nil {
		errors.Wrap(err, "GetUserWxInfoByUserId")
		retcode = utils.DB_READ_ERROR
		return
	}
	if num > 0 {
		info = &userWxInfos[0]
	}
	return
}

// GetAllUserWxInfos retrieves all UserWxInfos matches certain condition. Returns empty list if
// no records exist
func GetAllUserWxInfos(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(UserWxInfos))
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

	var l []UserWxInfos
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
