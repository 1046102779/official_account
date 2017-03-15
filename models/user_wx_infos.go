package models

import (
	"time"

	"github.com/1046102779/official_account/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

type UserWxInfos struct {
	Id         int       `orm:"column(user_wx_info_id);auto"`
	UserId     int       `orm:"column(user_id);null"`
	CustomerId int       `orm:"column(customer_id);null`
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

func (t *UserWxInfos) ReadUserWxInfoNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter ReadUserWxInfoNoLock.", t.Id)
	defer Logger.Info("[%v] left ReadUserWxInfoNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__DB__INSERT
		return
	}
	if err = (*o).Read(t); err != nil {
		err = errors.Wrap(err, "ReadUserWxInfoNoLock")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	return
}

func (t *UserWxInfos) UpdateUserWxInfoNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter UpdateUserWxInfoNoLock.", t.Id)
	defer Logger.Info("[%v] left UpdateUserWxInfoNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__DB__INSERT
		return
	}
	if _, err = (*o).Update(t); err != nil {
		err = errors.Wrap(err, "UpdateUserWxInfoNoLock")
		retcode = consts.ERROR_CODE__DB__UPDATE
		return
	}
	return
}
func (t *UserWxInfos) InsertUserWxInfoNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter InsertUserWxInfoNoLock.")
	defer Logger.Info("[%v] left InsertUserWxInfoNoLock.")
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__DB__INSERT
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		err = errors.Wrap(err, "InsertUserWxInfoNoLock")
		retcode = consts.ERROR_CODE__DB__INSERT
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
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		info = &userWxInfos[0]
	}
	return
}

// 通过用户ID，获取微信ID
func GetUserWxIdByUserId(userId int) (userWxInfoId int, retcode int, err error) {
	Logger.Info("[%v] enter GetUserWxIdByUserId.", userId)
	defer Logger.Info("[%v] left GetUserWxIdByUserId.", userId)
	var (
		wxInfo *UserWxInfos
	)
	if wxInfo, retcode, err = GetUserWxInfoByUserId(userId); err != nil {
		err = errors.Wrap(err, "GetUserWxIdByUserId")
		return
	}
	if wxInfo != nil {
		return wxInfo.Id, 0, nil
	}
	return
}

// 通过客户ID，获取主键ID
func GetUserWxInfoByCustomerId(customerId int) (info *UserWxInfos, retcode int, err error) {
	Logger.Info("[%v] enter GetUserWxInfoByCustomerId.", customerId)
	defer Logger.Info("[%v] left GetUserWxInfoByCustomerId.", customerId)
	var (
		userWxInfos []UserWxInfos = []UserWxInfos{}
		num         int64
	)
	o := orm.NewOrm()
	num, err = o.QueryTable((&UserWxInfos{}).TableName()).Filter("customer_id", customerId).All(&userWxInfos)
	if err != nil {
		errors.Wrap(err, "GetUserWxInfoByCustomerId")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		info = &userWxInfos[0]
	}
	return
}

// 通过客户ID，获取微信ID
func GetUserWxIdByCustomerId(customerId int) (userWxInfoId int, retcode int, err error) {
	Logger.Info("[%v] enter GetUserWxIdByCustomerId.", customerId)
	defer Logger.Info("[%v] left GetUserWxIdByCustomerId.", customerId)
	var (
		wxInfo *UserWxInfos
	)
	if wxInfo, retcode, err = GetUserWxInfoByCustomerId(customerId); err != nil {
		err = errors.Wrap(err, "GetUserWxIdByCustomerId")
		return
	}
	if wxInfo != nil {
		return wxInfo.Id, 0, nil
	}
	return
}
