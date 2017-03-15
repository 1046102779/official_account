// 微信公众号支付参数服务列表
// 1. 支付证书上传
// 2. 支付参数修改
// 3. 通过公众号ID，获取公众号支付参数记录
package models

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/1046102779/official_account/common/consts"
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
		retcode = consts.ERROR_CODE__DB__UPDATE
		return
	}
	if _, err = (*o).Update(t); err != nil {
		err = errors.Wrap(err, "UpdateOfficialAccountsPayParamsNoLock")
		retcode = consts.ERROR_CODE__DB__UPDATE
		return
	}
	return
}

func (t *OfficialAccountsPayParams) InsertOfficialAccountsPayParamsNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("enter InsertOfficialAccountsPayParamsNoLock.")
	defer Logger.Info("left InsertOfficialAccountsPayParamsNoLock.")
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__DB__UPDATE
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		retcode = consts.ERROR_CODE__DB__INSERT
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
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	if header.Filename != "apiclient_key.pem" && header.Filename != "apiclient_cert.pem" {
		err = errors.New("form param `filename` error")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	defer file.Close()
	os.Mkdir(fmt.Sprintf("%s/%s", conf.CertificationDir, officialAccount.Appid), os.ModePerm)
	newFile, newErr := os.OpenFile(fmt.Sprintf("%s/%s/%s", conf.CertificationDir, officialAccount.Appid, header.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	defer newFile.Close()
	if newErr != nil {
		err = newErr
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}

	_, err = io.Copy(newFile, file)
	if err != nil {
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
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
	num, err = o.QueryTable((&OfficialAccountsPayParams{}).TableName()).Filter("official_account_id", id).Filter("status", consts.STATUS_VALID).All(&officialAccountsPayParams)
	if err != nil {
		err = errors.Wrap(err, "ModifyWechatParams")
		retcode = consts.ERROR_CODE__DB__READ
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
			Status:            consts.STATUS_VALID,
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

// 3. 通过公众号ID，获取公众号支付参数记录
func GetOfficialAccountPayParamByOfficialAccountId(id int) (officialAccountPayParam *OfficialAccountsPayParams, retcode int, err error) {
	Logger.Info("[%v] enter GetOfficialAccountPayParamByOfficialAccountId.", id)
	defer Logger.Info("[%v] left GetOfficialAccountPayParamByOfficialAccountId.", id)
	var (
		officialAccountsPayParams []*OfficialAccountsPayParams = []*OfficialAccountsPayParams{}
		num                       int64
	)
	o := orm.NewOrm()
	num, err = o.QueryTable((&OfficialAccountsPayParams{}).TableName()).Filter("official_account_id", id).Filter("status", consts.STATUS_VALID).All(&officialAccountsPayParams)
	if err != nil {
		err = errors.Wrap(err, "GetOfficialAccountPayParamByOfficialAccountId")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		officialAccountPayParam = officialAccountsPayParams[0]
	}
	return
}
