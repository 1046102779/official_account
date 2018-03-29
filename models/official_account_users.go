// 公众号用户组服务列表
// 1. 通过OfficialAccountId和UserWxInfoId，获取openid
// 2. 通过code换取access_token
// 3. 获取可以得到请求CODE的url
// 4. 通过open_id，获取user_wx_info的主键ID
package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/1046102779/common/consts"
	"github.com/1046102779/common/httpRequest"
	"github.com/1046102779/common/types"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
)

type WxBaseInfoResp struct {
	Openid     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int16    `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	Headimgurl string   `json:"headimgurl"`
	Privileges []string `json:"privileges"`
	Unionid    string   `json:"unionid"`
}

type OfficialAccountUsers struct {
	Id                int       `orm:"column(official_account_user_id);auto"`
	OfficialAccountId int       `orm:"column(official_account_id);null"`
	UserWxInfoId      int       `orm:"column(user_wx_info_id);null"`
	Openid            string    `orm:"column(openid);size(50);null"`
	Status            int16     `orm:"column(status);null"`
	CreatedAt         time.Time `orm:"column(created_at);type(datetime);null"`
}

func (t *OfficialAccountUsers) TableName() string {
	return "official_account_users"
}

func (t *OfficialAccountUsers) InsertOfficialAccountNoLock(o *orm.Ormer) (retcode int, err error) {
	Logger.Info("[%v] enter InsertOfficialAccountNoLock.", t.OfficialAccountId)
	defer Logger.Info("[%v] left InsertOfficialAccountNoLock.", t.OfficialAccountId)
	if o == nil {
		err = errors.New("param `orm.Ormer` ptr empty")
		retcode = consts.ERROR_CODE__DB__INSERT
		return
	}
	if _, err = (*o).Insert(t); err != nil {
		err = errors.Wrap(err, "InsertOfficialAccountNoLock")
		return
	}
	return
}

func init() {
	orm.RegisterModel(new(OfficialAccountUsers))
}

// 获取可以得到请求CODE的url
func OfficialAccountAuthorizationUser(id int, callbackUrl string) (httpStr string, retcode int, err error) {
	Logger.Info("[%v] enter OfficialAccountAuthorizationUser.", id)
	defer Logger.Info("[%v] left OfficialAccountAuthorizationUser.", id)
	o := orm.NewOrm()
	officialAccount := &OfficialAccounts{
		Id: id,
	}
	if retcode, err = officialAccount.ReadOfficialAccountNoLock(&o); err != nil {
		err = errors.Wrap(err, "OfficialAccountAuthorizationUser")
		return
	}
	appid := officialAccount.Appid
	var oap *types.OfficialAccountPlatform
	if oap, err = conf.WRServerRPC.GetOfficialAccountPlatformInfo(); err != nil {
		err = errors.Wrap(err, "OfficialAccountAuthorizationUser")
		return
	}
	httpStr = fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&component_appid=%s#wechat_redirect", appid, url.QueryEscape(callbackUrl), "snsapi_userinfo", oap.Appid)
	return
}

// 绑定要返回的object对象信息{user，customer}
type BindingSimpleInfo struct {
	Id     int    `json:"id"`
	Mobile string `json:"mobile"`
}

// param bindingStatus // 微信是否已经绑定过用户
// param user, customer // 如果已绑定，则返回id等信息
// 通过code换取access_token
func GetUserAccessToken(appid string, code string) (
	bindingStatus int16,
	user *BindingSimpleInfo,
	customer *BindingSimpleInfo,
	openid string,
	retcode int, err error) {
	Logger.Info("[%v] enter GetUserAccessToken.", appid)
	defer Logger.Info("[%v] left GetUserAccessToken.", appid)
	var (
		retJson            map[string]interface{} = map[string]interface{}{}
		retBody            []byte
		userId, customerId int = 0, 0 // 如果微信已绑定用户，则返回userId或者customerID
	)
	user = new(BindingSimpleInfo)
	customer = new(BindingSimpleInfo)
	var oap *types.OfficialAccountPlatform
	if oap, err = conf.WRServerRPC.GetOfficialAccountPlatformInfo(); err != nil {
		err = errors.Wrap(err, "get wechat platform token failed rpc")
		return
	}
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/component/access_token?appid=%s&code=%s&grant_type=authorization_code&component_appid=%s&component_access_token=%s", appid, code, oap.Appid, oap.ComponentAccessToken)
	if retJson, err = httpRequest.HttpGetJson(httpStr); err != nil {
		err = errors.Wrap(err, "GetUserAccessToken")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	if _, ok := retJson["errcode"]; ok {
		err = errors.New(retJson["errmsg"].(string))
		err = errors.Wrap(err, "GetUserAccessToken")
		retcode = int(retJson["errcode"].(float64))
		return
	}
	// 获取用户授权信息, 包括token和refresh_token
	accessToken := retJson["access_token"].(string)
	openid = retJson["openid"].(string)
	httpStr = fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN", accessToken, openid)
	if retBody, err = httpRequest.HttpGetBody(httpStr); err != nil {
		err = errors.Wrap(err, "GetUserAccessToken")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	wxBaseInfoResp := new(WxBaseInfoResp)
	if err = json.Unmarshal(retBody, wxBaseInfoResp); err != nil {
		err = errors.Wrap(err, "GetUserAccessToken")
		retcode = consts.ERROR_CODE__JSON__PARSE_FAILED
		return
	}
	fmt.Println("wxBaseInfoResp: ", *wxBaseInfoResp)
	// 1. 判断微信用户是否已经存在
	// 2. 判断微信用户在该公众号appid下是否已存在
	// 3. 否则，添加相关记录
	if bindingStatus, userId, customerId, retcode, err = AddAuthorizationUserWxInfo(wxBaseInfoResp, appid); err != nil {
		err = errors.Wrap(err, "GetUserAccessToken")
		return
	}
	// 填充绑定要返回的信息
	switch bindingStatus {
	case consts.TYPE__WECHAT_USER_BINDING__USER:
		// rpc 获取B端用户信息
		user.Id = userId
		user.Mobile, err = conf.UserServerRPC.GetWechatBindingUserInfo(userId, consts.TYPE__WECHAT_USER_BINDING__USER)
		if err != nil {
			err = errors.Wrap(err, "rpc get b-end user info failed")
			return
		}
	case consts.TYPE__WECHAT_USER_BINDING__CUSTOMER:
		//  rpc 获取客户信息
		customer.Id = customerId
		customer.Mobile, err =
			conf.UserServerRPC.GetWechatBindingUserInfo(customerId, consts.TYPE__WECHAT_USER_BINDING__CUSTOMER)
		if err != nil {
			err = errors.Wrap(err, "rpc get c-end user info failed.")
		}
	default:
		bindingStatus = consts.TYPE__WECHAT_USER_BINDING__NO
	}
	return
}

func AddAuthorizationUserWxInfo(base *WxBaseInfoResp, appid string) (
	bindingStatus int16,
	userId int,
	customerId int,
	retcode int, err error) {
	Logger.Info("[%v] enter AddAuthorizationUserWxInfo.", appid)
	defer Logger.Info("[%v] left AddAuthorizationUserWxInfo.", appid)
	var (
		userWxInfos          = []UserWxInfos{}
		officialAccountUsers = []OfficialAccountUsers{}
		num                  int64
		userWxInfoId         int
	)
	o := orm.NewOrm()
	now := time.Now()
	// 1. 判断微信用户是否已经存在
	num, err = o.QueryTable((&UserWxInfos{}).TableName()).Filter("nickname", base.Nickname).Filter("headimgurl", base.Headimgurl).All(&userWxInfos)
	if err != nil {
		err = errors.Wrap(err, "AddAuthorizationUserWxInfo")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		// 微信用户已存在
		userWxInfoId = userWxInfos[0].Id
		userId = userWxInfos[0].UserId
		customerId = userWxInfos[0].CustomerId
		if userId <= 0 && customerId <= 0 {
			bindingStatus = consts.TYPE__WECHAT_USER_BINDING__NO
		} else if userId > 0 {
			bindingStatus = consts.TYPE__WECHAT_USER_BINDING__USER
		} else if customerId > 0 {
			bindingStatus = consts.TYPE__WECHAT_USER_BINDING__CUSTOMER
		}
	} else {
		// 新增微信用户
		var privilege string
		for index := 0; base.Privileges != nil && index < len(base.Privileges); index++ {
			if privilege == "" {
				privilege = fmt.Sprintf("%s", base.Privileges[index])
			} else {
				privilege = fmt.Sprintf("%s,%s", privilege, base.Privileges[index])
			}
		}
		userWxInfo := &UserWxInfos{
			Nickname:   base.Nickname,
			Sex:        base.Sex,
			Province:   base.Province,
			City:       base.City,
			Country:    base.Country,
			Headimgurl: base.Headimgurl,
			Privilege:  privilege,
			CreatedAt:  now,
		}
		if retcode, err = userWxInfo.InsertUserWxInfoNoLock(&o); err != nil {
			err = errors.Wrap(err, "AddAuthorizationUserWxInfo")
			return
		}
		userWxInfoId = userWxInfo.Id
	}
	// 2. 判断微信用户在该公众号appid下是否已存在
	num, err = o.QueryTable((&OfficialAccountUsers{}).TableName()).Filter("user_wx_info_id", userWxInfoId).Filter("openid", base.Openid).All(&officialAccountUsers)
	if err != nil {
		err = errors.Wrap(err, "AddAuthorizationUserWxInfo")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num <= 0 {
		// 新增公众账号的用户
		officialAccounts := []OfficialAccounts{}
		num, err = o.QueryTable((&OfficialAccounts{}).TableName()).Filter("appid", appid).Filter("status", consts.STATUS_VALID).All(&officialAccounts)
		if err != nil {
			err = errors.Wrap(err, "AddAuthorizationUserWxInfo")
			retcode = consts.ERROR_CODE__DB__READ
			return
		}
		if num > 0 {
			officialAccount := &OfficialAccountUsers{
				OfficialAccountId: officialAccounts[0].Id,
				UserWxInfoId:      userWxInfoId,
				Openid:            base.Openid,
				Status:            consts.STATUS_VALID,
				CreatedAt:         now,
			}
			if retcode, err = officialAccount.InsertOfficialAccountNoLock(&o); err != nil {
				err = errors.Wrap(err, "AddAuthorizationUserWxInfo")
				return
			}
		}
	}
	return
}

// 通过OfficialAccountId和UserWxInfoId，获取openid
func GetopenidByCond(officialAccountId int, userWxInfoId int) (openid string, retcode int, err error) {
	Logger.Info("[%v.%v] enter GetopenidByCond.", officialAccountId, userWxInfoId)
	defer Logger.Info("[%v.%v] left GetopenidByCond.", officialAccountId, userWxInfoId)
	var (
		officialAccountUsers []OfficialAccountUsers = []OfficialAccountUsers{}
		num                  int64
	)
	if officialAccountId <= 0 || userWxInfoId <= 0 {
		err = errors.New("param `officialAccountId|| userWxInfoId` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	o := orm.NewOrm()
	num, err = o.QueryTable((&OfficialAccountUsers{}).TableName()).Filter("official_account_id", officialAccountId).Filter("user_wx_info_id", userWxInfoId).Filter("status", consts.STATUS_VALID).All(&officialAccountUsers)
	if err != nil {
		err = errors.Wrap(err, "GetopenidByCond")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		openid = officialAccountUsers[0].Openid
	}
	return
}

// 4. 通过open_id，获取user_wx_info的主键ID
func GetUserWxInfoIdByOpenid(officialAccountId int, openid string) (id int, retcode int, err error) {
	Logger.Info("[%v] enter GetUserWxInfoIdByOpenid.", openid)
	defer Logger.Info("[%v] left GetUserWxInfoIdByOpenid.", openid)
	var (
		num                  int64
		officialAccountUsers []*OfficialAccountUsers = []*OfficialAccountUsers{}
	)
	o := orm.NewOrm()
	num, err = o.QueryTable((&OfficialAccountUsers{}).TableName()).Filter("official_account_id", officialAccountId).Filter("openid", openid).All(&officialAccountUsers)
	if err != nil {
		err = errors.Wrap(err, "GetUserWxInfoIdByOpenid")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		id = officialAccountUsers[0].UserWxInfoId
	}
	return
}

// 5. 通过公司ID和用户ID，获取公众号下的用户openid
func getWxOpenIdByCompanyIdAndUserId(companyId int, userId int) (openid string, retcode int, err error) {
	Logger.Info("[%v.%v] enter getWxOpenIdByCompanyIdAndUserId.", companyId, userId)
	defer Logger.Info("[%v.%v] left getWxOpenIdByCompanyIdAndUserId.", companyId, userId)
	var (
		officialAccountId int = 0
		userWxId          int
	)
	if officialAccountId, retcode, err = getOfficialAccountIdByCompanyId(companyId); err != nil {
		err = errors.Wrap(err, "getWxOpenIdByCompanyIdAndUserId")
		return
	}
	if userWxId, retcode, err = GetUserWxIdByUserId(userId); err != nil {
		err = errors.Wrap(err, "getWxOpenIdByCompanyIdAndUserId")
		return
	}
	if openid, retcode, err = GetopenidByCond(officialAccountId, userWxId); err != nil {
		err = errors.Wrap(err, "getWxOpenIdByCompanyIdAndUserId")
		return
	}
	return
}

// 5. 通过公司ID和客户ID，获取公众号下的用户openid
func getWxOpenIdByCompanyIdAndCustomerId(companyId int, customerId int) (openid string, retcode int, err error) {
	Logger.Info("[%v.%v] enter getWxOpenIdByCompanyIdAndCustomerId.", companyId, customerId)
	defer Logger.Info("[%v.%v] left getWxOpenIdByCompanyIdAndCustomerId.", companyId, customerId)
	var (
		officialAccountId int = 0
		userWxId          int
	)
	if officialAccountId, retcode, err = getOfficialAccountIdByCompanyId(companyId); err != nil {
		err = errors.Wrap(err, "getWxOpenIdByCompanyIdAndUserId")
		return
	}
	if userWxId, retcode, err = GetUserWxIdByCustomerId(customerId); err != nil {
		err = errors.Wrap(err, "getWxOpenIdByCompanyIdAndUserId")
		return
	}
	if openid, retcode, err = GetopenidByCond(officialAccountId, userWxId); err != nil {
		err = errors.Wrap(err, "getWxOpenIdByCompanyIdAndUserId")
		return
	}
	return
}
