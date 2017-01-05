package models

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	utils "github.com/1046102779/common"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

type OfficialAccountIndustryCodes struct {
	Id                int       `orm:"column(official_account_industry_code_id);auto"`
	OfficialAccountId int       `orm:"column(official_account_id);null"`
	IndustryId1       int       `orm:"column(industry_id1);null"`
	IndustryId2       int       `orm:"column(industry_id2);null"`
	Status            int16     `orm:"column(status);null"`
	UpdatedAt         time.Time `orm:"column(updated_at);type(datetime);null"`
	CreatedAt         time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *OfficialAccountIndustryCodes) TableName() string {
	return "official_account_industry_codes"
}

// 添加行业信息
func (t *OfficialAccountIndustryCodes) InsertOfficialAccountIndustryCodesNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v.%v] enter InsertOfficialAccountIndustryCodesNoLock.", t.IndustryId1, t.IndustryId2)
	defer Logger.Info("[%v.%v] left InsertOfficialAccountIndustryCodesNoLock.", t.IndustryId1, t.IndustryId2)
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

func (t *OfficialAccountIndustryCodes) ReadOfficialAccountIndustryNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter ReadOfficialAccountIndustryNoLock.", t.Id)
	defer Logger.Info("[%v] left ReadOfficialAccountIndustryNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	if err = (*o).Read(t); err != nil {
		retcode = utils.DB_READ_ERROR
		return
	}
	return
}

// 获取设置的行业信息
func (t *OfficialAccountIndustryCodes) GetOfficialAccountIndustryNoLock() (firstIndustry string, secIndustry string, retcode int, err error) {
	Logger.Info("[%v] enter GetOfficialAccountIndustry.", t.OfficialAccountId)
	defer Logger.Info("[%v] left GetOfficialAccountIndustry.", t.OfficialAccountId)
	var (
		officialAccountIndustryCodes []OfficialAccountIndustryCodes = []OfficialAccountIndustryCodes{}
		num                          int64
	)
	if t.OfficialAccountId <= 0 {
		err = errors.New("param `official_account_id` empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	o := orm.NewOrm()
	num, err = o.QueryTable(t.TableName()).Filter("official_account_id", t.OfficialAccountId).Filter("status", utils.STATUS_VALID).All(&officialAccountIndustryCodes)
	if err != nil {
		err = errors.Wrap(err, "GetOfficialAccountIndustryNoLock")
		retcode = utils.DB_READ_ERROR
		return
	}
	if num > 0 {
		*t = officialAccountIndustryCodes[0]
		if officialAccountIndustryCodes[0].IndustryId1 > 0 {
			industryCodeQuery := &IndustryCodeQuerys{
				Id: officialAccountIndustryCodes[0].IndustryId1,
			}
			retcode, err = industryCodeQuery.GetIndustryCodeQueryNoLock(&o)
			if err != nil {
				err = errors.Wrap(err, "GetOfficialAccountIndustryNoLock")
				return
			}
		}
		if officialAccountIndustryCodes[0].IndustryId2 > 0 {
			industryCodeQuery := &IndustryCodeQuerys{
				Id: officialAccountIndustryCodes[0].IndustryId2,
			}
			retcode, err = industryCodeQuery.GetIndustryCodeQueryNoLock(&o)
			if err != nil {
				err = errors.Wrap(err, "GetOfficialAccountIndustryNoLock")
				return
			}
		}
	} else {
		// 从微信公众号接口获取
	}
	return
}

// 更新公众号所属行业
func (t *OfficialAccountIndustryCodes) UpdateOfficialAccountIndustryCodeNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter UpdateOfficialAccountIndustryCodeNoLock.", t.OfficialAccountId)
	defer Logger.Info("[%v] left UpdateOfficialAccountIndustryCodeNoLock.", t.OfficialAccountId)
	var (
		original []OfficialAccountIndustryCodes = []OfficialAccountIndustryCodes{}
		num      int64
	)
	now := time.Now()
	if t.IndustryId1 <= 0 || t.IndustryId2 <= 0 {
		err = errors.New("params `industry_id1 | industry_id2` empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr  empty")
		retcode = utils.DB_UPDATE_ERROR
		return
	}

	// 公众号模板消息所属行业编号,每个月只允许修改一次
	num, err = (*o).QueryTable(t.TableName()).Filter("official_account_id", t.OfficialAccountId).All(&original)
	if err != nil {
		err = errors.Wrap(err, "UpdateOfficialAccountIndustryCodeNoLock")
		retcode = utils.DB_READ_ERROR
		return
	}
	if num > 0 {
		upDate := now.AddDate(0, -1, 0)
		if upDate.After(original[0].UpdatedAt) {
			fmt.Println("Set wechat official account industry.")
			// 允许修改
			t.Id = original[0].Id
			t.UpdatedAt = now
			if _, err = (*o).Update(t); err != nil {
				err = errors.Wrap(err, "UpdateOfficialAccountIndustryCodeNoLock")
				retcode = utils.DB_UPDATE_ERROR
				return
			}
			// 设置公众号所属行业
			weMessageTemplate := new(WeMessageTemplate)
			officialAccount := OfficialAccounts{
				Id: t.OfficialAccountId,
			}
			if retcode, err = officialAccount.ReadOfficialAccountNoLock(o); err != nil {
				err = errors.Wrap(err, "UpdateOfficialAccountIndustryCodeNoLock")
				retcode = utils.DB_READ_ERROR
				return
			}
			retcode, err = weMessageTemplate.SetOfficialAccountIndustry(t.IndustryId1, t.IndustryId2, conf.WechatAuthTTL.AuthorizerMap[officialAccount.Appid].AuthorizerAccessToken)
			if err != nil {
				err = errors.Wrap(err, "UpdateOfficialAccountIndustryCodeNoLock")
				return
			}
		}
	}
	return
}

func init() {
	orm.RegisterModel(new(OfficialAccountIndustryCodes))
}

// GetAllOfficialAccountIndustryCodes retrieves all OfficialAccountIndustryCodes matches certain condition. Returns empty list if
// no records exist
func GetAllOfficialAccountIndustryCodes(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(OfficialAccountIndustryCodes))
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

	var l []OfficialAccountIndustryCodes
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
