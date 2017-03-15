package models

import (
	"reflect"
	"strings"
	"time"

	"github.com/1046102779/official_account/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

type IndustryCodeQuerys struct {
	Id               int       `orm:"column(industry_code_query_id);auto"`
	MainType         int       `orm:"column(main_type);null"`
	MainIndustryCode string    `orm:"column(main_industry_code);size(30);null"`
	SecIndustryCode  string    `orm:"column(sec_industry_code);size(30);null"`
	CodeNum          int       `orm:"column(code_num);null"`
	Status           int16     `orm:"column(status);null"`
	UpdatedAt        time.Time `orm:"column(updated_at);type(datetime);null"`
	CreatedAt        time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *IndustryCodeQuerys) TableName() string {
	return "industry_code_querys"
}

func (t *IndustryCodeQuerys) GetIndustryCodeQueryNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter GetIndustryCodeQueryNoLock.", t.Id)
	defer Logger.Info("[%v] left GetIndustryCodeQueryNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	if err = (*o).Read(t); err != nil {
		err = errors.Wrap(err, "GetIndustryCodeQueryNoLock")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	return
}

// 获取副行业列表
func (t *IndustryCodeQuerys) GetSecIndutryListNoLocks() (secIndustrys []IndustryInfoResp, retcode int, err error) {
	Logger.Info("enter GetSecIndutryListNoLocks.")
	defer Logger.Info("left GetSecIndutryListNoLocks.")
	var (
		industryCodeQuerys []IndustryCodeQuerys = []IndustryCodeQuerys{}
		num                int64
	)
	secIndustrys = []IndustryInfoResp{}
	if t.MainType <= 0 {
		err = errors.New("param `main_type` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	o := orm.NewOrm()
	num, err = o.QueryTable(t.TableName()).Filter("main_type", t.MainType).All(&industryCodeQuerys)
	if err != nil {
		retcode = consts.ERROR_CODE__DB__READ
		err = errors.Wrap(err, "GetSecIndutryListNoLocks")
		return
	}
	for index := 0; index < int(num); index++ {
		secIndustrys = append(secIndustrys, IndustryInfoResp{
			Id:   industryCodeQuerys[index].Id,
			Name: industryCodeQuerys[index].SecIndustryCode,
		})
	}
	return
}

func init() {
	orm.RegisterModel(new(IndustryCodeQuerys))
}

type IndustryInfoResp struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func GetIndustryIdByName(mainIndustry string, secIndustry string) (id int, retcode int, err error) {
	Logger.Info("[%v.%v] enter GetIndustryIdByName.", mainIndustry, secIndustry)
	defer Logger.Info("[%v.%v] left GetIndustryIdByName.", mainIndustry, secIndustry)
	var (
		num           int64
		industryCodes []IndustryCodeQuerys = []IndustryCodeQuerys{}
	)
	o := orm.NewOrm()
	num, err = o.QueryTable((&IndustryCodeQuerys{}).TableName()).Filter("main_industry_code", mainIndustry).Filter("sec_industry_code", secIndustry).Filter("status", consts.STATUS_VALID).All(&industryCodes)
	if err != nil {
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		id = industryCodes[0].Id
	}
	return
}

// 获取主行业列表
func GetAllMainIndutryNoLocks() (mainIndustrys []IndustryInfoResp, retcode int, err error) {
	Logger.Info("enter GetAllMainIndutry.")
	defer Logger.Info("left GetAllMainIndutry.")
	var (
		industryCodeQuerys []IndustryCodeQuerys = []IndustryCodeQuerys{}
		num                int64
	)
	mainIndustrys = []IndustryInfoResp{}
	o := orm.NewOrm()
	num, err = o.QueryTable((&IndustryCodeQuerys{}).TableName()).Filter("status", consts.STATUS_VALID).GroupBy("main_type").All(&industryCodeQuerys)
	if err != nil {
		err = errors.Wrap(err, "GetAllMainIndutryNoLocks")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	for index := 0; index < int(num); index++ {
		mainIndustrys = append(mainIndustrys, IndustryInfoResp{
			Id:   industryCodeQuerys[index].MainType,
			Name: industryCodeQuerys[index].MainIndustryCode,
		})
	}
	return
}

// GetAllIndustryCodeQuerys retrieves all IndustryCodeQuerys matches certain condition. Returns empty list if
// no records exist
func GetAllIndustryCodeQuerys(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(IndustryCodeQuerys))
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

	var l []IndustryCodeQuerys
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
