// 公众号服务列表
// 1. 公众号与公司ID的绑定
// 2. 获取公众号基本信息
// 3. 获取公众号列表
// 4. 通过公司ID，获取公众号信息
package models

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/1046102779/official_account/common/consts"
	"github.com/1046102779/official_account/conf"
	pb "github.com/1046102779/official_account/igrpc"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"

	"github.com/astaxie/beego/orm"
)

const ()

type OfficialAccounts struct {
	Id                    int       `orm:"column(official_account_id);auto" json:"id"`
	CompanyId             int       `orm:"column(company_id);null" json:"company_id"`
	Nickname              string    `orm:"column(nickname);size(50);null" json:"nickname"`
	AvartarUrl            string    `orm:"column(avartar_url);size(300);null" json:"avatar_url"`
	ServiceTypeId         int16     `orm:"column(service_type_id);null" json:"service_type_id"`
	VerifyTypeId          int16     `orm:"column(verify_type_id);null" json:"verify_type_id"`
	OriginalId            string    `orm:"column(original_id);size(40);null" json:"original_id"`
	PrincipalName         string    `orm:"column(principal_name);size(300);null" json:"principal_name"`
	Alias                 string    `orm:"column(alias);size(100);null" json:"alias"`
	BusinessInfoOpenStore int16     `orm:"column(business_info_open_store);null" json:"business_info_open_store"`
	BusinessInfoOpenScan  int16     `orm:"column(business_info_open_scan);null" json:"business_info_open_scan"`
	BusinessInfoOpenPay   int16     `orm:"column(business_info_open_pay);null" json:"business_info_open_pay"`
	BusinessInfoOpenCard  int16     `orm:"column(business_info_open_card);null" json:"business_info_open_card"`
	BusinessInfoOpenShake int16     `orm:"column(business_info_open_shake);null" json:"business_info_open_shake"`
	QrcodeUrl             string    `orm:"column(qrcode_url);size(300);null" json:"qrcode_url"`
	Appid                 string    `orm:"column(appid);size(100);null" json:"appid"`
	FuncIds               string    `orm:"column(func_ids);size(100);null" json:"func_ids"`
	Status                int16     `orm:"column(status);null"`
	UpdatedAt             time.Time `orm:"column(updated_at);type(datetime);null"`
	CreatedAt             time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *OfficialAccounts) TableName() string {
	return "official_accounts"
}

func (t *OfficialAccounts) UpdatedOfficialAccountNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter UpdatedOfficialAccountNoLock.", t.Id)
	defer Logger.Info("[%v] left UpdatedOfficialAccountNoLock.", t.Id)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	if _, err = (*o).Update(t); err != nil {
		err = errors.Wrap(err, "UpdatedOfficialAccountNoLock")
		retcode = consts.ERROR_CODE__DB__UPDATE
		return
	}
	return
}

func (t *OfficialAccounts) ReadOfficialAccountNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter ReadOfficialAccount.")
	defer Logger.Info("[%v] enter ReadOfficialAccount.")
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	if err = (*o).Read(t); err != nil {
		err = errors.Wrap(err, "ReadOfficialAccountNoLock")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	return
}

func (t *OfficialAccounts) InsertOfficialAccountNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter InsertOfficialAccountNoLock.", t.Appid)
	defer Logger.Info("[%v] left InsertOfficialAccountNoLock.", t.Appid)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	if "" == strings.TrimSpace(t.Nickname) || "" == strings.TrimSpace(t.Appid) {
		err = errors.New("param `official_account's nickname | appid` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		err = errors.Wrap(err, "InsertOfficialAccountNoLock")
		retcode = consts.ERROR_CODE__DB__INSERT
		return
	}
	return
}

func init() {
	orm.RegisterModel(new(OfficialAccounts))
}

func GetOfficialAccountBaseInfo(platformAppid string, appid string) (offAcc *OfficialAccounts, retcode int, err error) {
	var (
		funcIdStr        string
		num              int64
		officialAccounts []OfficialAccounts = []OfficialAccounts{}
	)
	now := time.Now()
	authorizer := new(Authorizer)
	in := &pb.OfficialAccountPlatform{}
	conf.WxRelayServerClient.Call(fmt.Sprintf("%s.%s", "wx_relay_server", "GetOfficialAccountPlatformInfo"), in, in)
	authorizer, retcode, err = GetOfficialAccountBaseInfoExternal(in.ComponentAccessToken, platformAppid, appid)
	if err != nil {
		err = errors.Wrap(err, "authorized get baseinfo failed.")
		return
	}
	for index := 0; authorizer.AuthorizationInfo.FuncInfos != nil && index < len(authorizer.AuthorizationInfo.FuncInfos); index++ {
		if funcIdStr == "" {
			funcIdStr = fmt.Sprintf("%d", authorizer.AuthorizationInfo.FuncInfos[index].FuncScope.Id)
		} else {
			funcIdStr = fmt.Sprintf("%s,%d", funcIdStr, authorizer.AuthorizationInfo.FuncInfos[index].FuncScope.Id)
		}
	}
	o := orm.NewOrm()
	// 判断该公众号是否在系统中已存在
	num, err = o.QueryTable((&OfficialAccounts{}).TableName()).Filter("appid", appid).Filter("status", consts.STATUS_VALID).All(&officialAccounts)
	if err != nil {
		err = errors.Wrap(err, "GetOfficialAccountBaseInfo")
		return
	}
	if num > 0 {
		offAcc = &officialAccounts[0]
		return
	}
	// 新增公众号到库表
	offAcc = &OfficialAccounts{
		Nickname:      authorizer.AuthorizerInfo.Nickname,
		AvartarUrl:    authorizer.AuthorizerInfo.HeadImg,
		ServiceTypeId: int16(authorizer.AuthorizerInfo.ServiceTypeInfo.Id),
		VerifyTypeId:  int16(authorizer.AuthorizerInfo.VerifyTypeInfo.Id),
		OriginalId:    authorizer.AuthorizerInfo.UserName,
		PrincipalName: authorizer.AuthorizerInfo.PrincipalName,
		Alias:         authorizer.AuthorizerInfo.Alias,
		BusinessInfoOpenStore: int16(authorizer.AuthorizerInfo.BusinessInfo.OpenStore),
		BusinessInfoOpenScan:  int16(authorizer.AuthorizerInfo.BusinessInfo.OpenScan),
		BusinessInfoOpenPay:   int16(authorizer.AuthorizerInfo.BusinessInfo.OpenPay),
		BusinessInfoOpenCard:  int16(authorizer.AuthorizerInfo.BusinessInfo.OpenCard),
		BusinessInfoOpenShake: int16(authorizer.AuthorizerInfo.BusinessInfo.OpenShake),
		QrcodeUrl:             authorizer.AuthorizerInfo.QrcodeUrl,
		Appid:                 appid,
		FuncIds:               funcIdStr,
		Status:                consts.STATUS_VALID,
		UpdatedAt:             now,
		CreatedAt:             now,
	}
	if retcode, err = offAcc.InsertOfficialAccountNoLock(&o); err != nil {
		err = errors.Wrap(err, "GetOfficialAccountBaseInfo")
		return
	}
	return
}

// GetAllOfficialAccounts retrieves all OfficialAccounts matches certain condition. Returns empty list if
// no records exist
func GetAllOfficialAccounts(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(OfficialAccounts))
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

	var l []OfficialAccounts
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

// 1. 公众号与公司ID的绑定
func BindingCompanyAndOfficialAccount(officialAccountId int, companyId int) (retcode int, err error) {
	Logger.Info("[%v.%v] enter BindingCompanyAndOfficialAccount.", officialAccountId, companyId)
	defer Logger.Info("[%v.%v] left BindingCompanyAndOfficialAccount.", officialAccountId, companyId)
	o := orm.NewOrm()
	now := time.Now()
	officialAccount := &OfficialAccounts{
		Id: officialAccountId,
	}
	if retcode, err = officialAccount.ReadOfficialAccountNoLock(&o); err != nil {
		err = errors.Wrap(err, "BindingCompanyAndOfficialAccount")
		return
	}
	// 如果已经绑定，则直接返回
	if officialAccount.CompanyId > 0 {
		return
	}
	officialAccount.UpdatedAt = now
	officialAccount.CompanyId = companyId
	if retcode, err = officialAccount.UpdatedOfficialAccountNoLock(&o); err != nil {
		err = errors.Wrap(err, "BindingCompanyAndOfficialAccount")
		return
	}
	return
}

type OfficialAccountSimpleInfo struct {
	Id               int    `json:"id"`
	IsStartWechatPay int16  `json:"is_start_wechat_pay"`
	Name             string `json:"name"`
}

// 3. 获取公众号列表
func GetOfficialAccounts(companyId int) (officialAccountSimpleInfos []*OfficialAccountSimpleInfo, retcode int, err error) {
	Logger.Info("[%v] enter GetOfficialAccounts.", companyId)
	defer Logger.Info("[%v] left GetOfficialAccounts.", companyId)
	var (
		officialAccountsPayParam *OfficialAccountsPayParams
		officialAccounts         []*OfficialAccounts = []*OfficialAccounts{}
		num                      int64
		WechatPayStatus          int16 = consts.TYPE__STORAGE__UNABLED
	)
	officialAccountSimpleInfos = []*OfficialAccountSimpleInfo{}
	o := orm.NewOrm()
	num, err = o.QueryTable((&OfficialAccounts{}).TableName()).Filter("company_id", companyId).Filter("status", consts.STATUS_VALID).All(&officialAccounts)
	if err != nil {
		err = errors.Wrap(err, "GetOfficialAccounts")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		for index := 0; index < int(num); index++ {
			// 1. 获取公众号ID、名称
			// 2. 获取是否公众号已开启支付
			officialAccountsPayParam, retcode, err = GetOfficialAccountPayParamByOfficialAccountId(officialAccounts[index].Id)
			if err != nil {
				err = errors.Wrap(err, "GetOfficialAccounts")
				return
			}
			if officialAccountsPayParam != nil && officialAccountsPayParam.MchId != "" && officialAccountsPayParam.Appkey != "" {
				WechatPayStatus = consts.TYPE_PAY_STATUS__WECHAT__STARTED
			} else {
				WechatPayStatus = consts.TYPE__STORAGE__UNABLED
			}
			officialAccountSimpleInfos = append(officialAccountSimpleInfos, &OfficialAccountSimpleInfo{
				Id:               officialAccounts[index].Id,
				Name:             officialAccounts[index].Nickname,
				IsStartWechatPay: WechatPayStatus,
			})
		}
	}
	return
}

// 4. 通过公司ID，获取公众号信息
func GetOfficialAccountByCompanyId(companyId int) (officialAccount *OfficialAccounts, retcode int, err error) {
	Logger.Info("[%v] enter GetOfficialAccountByCompanyId.", companyId)
	defer Logger.Info("[%v] left GetOfficialAccountByCompanyId.", companyId)
	var (
		officialAccounts []*OfficialAccounts
		num              int64
	)
	o := orm.NewOrm()
	num, err = o.QueryTable((&OfficialAccounts{}).TableName()).Filter("company_id", companyId).Filter("status", consts.STATUS_VALID).All(&officialAccounts)
	if err != nil {
		err = errors.Wrap(err, "GetOfficialAccountByCompanyId")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		officialAccount = officialAccounts[0]
	}
	return
}

func getOfficialAccountIdByCompanyId(companyId int) (officialAccountId int, retcode int, err error) {
	Logger.Info("[%v] enter getOfficialAccountIdByCompanyId.", companyId)
	defer Logger.Info("[%v] left getOfficialAccountIdByCompanyId.", companyId)
	var (
		officialAccount *OfficialAccounts
	)
	officialAccount, retcode, err = GetOfficialAccountByCompanyId(companyId)
	if err != nil {
		err = errors.Wrap(err, "getOfficialAccountIdByCompanyId")
		return
	}
	if officialAccount != nil {
		return officialAccount.Id, 0, nil
	}
	return
}
