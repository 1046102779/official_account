package models

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	utils "github.com/1046102779/common"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

type OfficialAccountsPayParams struct {
	Id                int       `orm:"column(official_accounts_pay_param_id);auto"`
	OfficialAccountId int       `orm:"column(official_account_id);null"`
	MchId             string    `orm:"column(mch_id);size(30);null"`
	Name              string    `orm:"column(name);size(100);null"`
	Appkey            string    `orm:"column(appkey);size(50);null"`
	Status            int16     `orm:"column(status);null"`
	UpdatedAt         time.Time `orm:"column(updated_at);type(datetime);null"`
	CreatedAt         time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *OfficialAccountsPayParams) TableName() string {
	return "official_accounts_pay_params"
}

func (t *OfficialAccountsPayParams) UpdateOfficialAccountsPayParamsNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter UpdateOfficialAccountsPayParamsNoLock.", t.Id)
	defer Logger.Info("[%v] left UpdateOfficialAccountsPayParamsNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = utils.DB_UPDATE_ERROR
		return
	}
	if _, err = (*o).Update(t); err != nil {
		err = errors.Wrap(err, "UpdateOfficialAccountsPayParamsNoLock")
		retcode = utils.DB_UPDATE_ERROR
		return
	}
	return
}

func (t *OfficialAccountsPayParams) InsertOfficialAccountsPayParamsNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("enter InsertOfficialAccountsPayParamsNoLock.")
	defer Logger.Info("left InsertOfficialAccountsPayParamsNoLock.")
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = utils.DB_UPDATE_ERROR
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		retcode = utils.DB_INSERT_ERROR
		err = errors.Wrap(err, "InsertOfficialAccountsPayParamsNoLock")
		return
	}
	return
}

func init() {
	orm.RegisterModel(new(OfficialAccountsPayParams))
}

func UploadCertification(id int, req *http.Request) (retcode int, err error) {
	Logger.Info("[%v] enter UploadCertification.", id)
	defer Logger.Info("[%v] left UploadCertification.", id)
	o := orm.NewOrm()
	officialAccount := &OfficialAccounts{
		Id: id,
	}
	if retcode, err = officialAccount.ReadOfficialAccountNoLock(&o); err != nil {
		err = errors.Wrap(err, "UploadCertification")
		return
	}
	// 读写文件
	req.ParseMultipartForm(32 << 20)
	file, header, newErr := req.FormFile("certification_file")
	if header == nil {
		err = errors.New("parse file empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	if header.Filename != "apiclient_key.pem" && header.Filename != "apiclient_cert.pem" {
		err = errors.New("form param `filename` error")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	defer file.Close()
	os.Mkdir(fmt.Sprintf("%s/%s", conf.CertificationDir, officialAccount.Appid), os.ModePerm)
	newFile, newErr := os.OpenFile(fmt.Sprintf("%s/%s/%s", conf.CertificationDir, officialAccount.Appid, header.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	defer newFile.Close()
	if newErr != nil {
		err = newErr
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}

	_, err = io.Copy(newFile, file)
	if err != nil {
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	// end
	return
}

func ModifyWechatParams(id int, appkey string, mchid string, name string) (retcode int, err error) {
	Logger.Info("[%v] enter ModifyWechatParams.", id)
	defer Logger.Info("[%v] left ModifyWechatParams.", id)
	var (
		officialAccountsPayParams []OfficialAccountsPayParams = []OfficialAccountsPayParams{}
		num                       int64
	)
	o := orm.NewOrm()
	now := time.Now()
	officialAccount := &OfficialAccounts{
		Id: id,
	}
	if retcode, err = officialAccount.ReadOfficialAccountNoLock(&o); err != nil {
		err = errors.Wrap(err, "ModifyWechatParams")
		return
	}
	num, err = o.QueryTable((&OfficialAccountsPayParams{}).TableName()).Filter("official_account_id", id).Filter("status", utils.STATUS_VALID).All(&officialAccountsPayParams)
	if err != nil {
		err = errors.Wrap(err, "ModifyWechatParams")
		retcode = utils.DB_READ_ERROR
		return
	}
	if num > 0 {
		// update wechat pay params
		if "" != strings.TrimSpace(appkey) {
			officialAccountsPayParams[0].Appkey = appkey
		}
		if "" != strings.TrimSpace(mchid) {
			officialAccountsPayParams[0].MchId = mchid
		}
		if "" != strings.TrimSpace(name) {
			officialAccountsPayParams[0].Name = name
		}
		officialAccountsPayParams[0].UpdatedAt = now
		if retcode, err = officialAccountsPayParams[0].UpdateOfficialAccountsPayParamsNoLock(&o); err != nil {
			err = errors.Wrap(err, "ModifyWechatParams")
			return
		}
	} else {
		// add wechat pay params
		officialAccountsPayParam := &OfficialAccountsPayParams{
			OfficialAccountId: id,
			MchId:             mchid,
			Name:              name,
			Appkey:            appkey,
			Status:            utils.STATUS_VALID,
			UpdatedAt:         now,
			CreatedAt:         now,
		}
		if retcode, err = officialAccountsPayParam.InsertOfficialAccountsPayParamsNoLock(&o); err != nil {
			err = errors.Wrap(err, "ModifyWechatParams")
			return
		}
	}
	return
}

// GetAllOfficialAccountsPayParams retrieves all OfficialAccountsPayParams matches certain condition. Returns empty list if
// no records exist
func GetAllOfficialAccountsPayParams(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(OfficialAccountsPayParams))
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

	var l []OfficialAccountsPayParams
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
