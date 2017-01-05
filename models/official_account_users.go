package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	utils "github.com/1046102779/common"
	"github.com/1046102779/common/httpRequest"
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
		retcode = utils.DB_INSERT_ERROR
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
func OfficialAccountAuthorizationUser(id int) (httpStr string, retcode int, err error) {
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
	httpStr = fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&component_appid=%s#wechat_redirect", appid, url.QueryEscape(fmt.Sprintf("%s%s", conf.HostName, "/v1/wechats/user/authorization/callback")), "snsapi_userinfo", conf.WechatParam.AppId)
	return
}

// 通过code换取access_token
func GetUserAccessToken(appid string, code string) (retcode int, err error) {
	Logger.Info("[%v] enter GetUserAccessToken.", appid)
	defer Logger.Info("[%v] left GetUserAccessToken.", appid)
	var (
		retJson map[string]interface{} = map[string]interface{}{}
		retBody []byte
	)
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/component/access_token?appid=%s&code=%s&grant_type=authorization_code&component_appid=%s&component_access_token=%s", appid, code, conf.WechatParam.AppId, conf.WechatAuthTTL.ComponentAccessToken)
	if retJson, err = httpRequest.HttpGetJson(httpStr); err != nil {
		err = errors.Wrap(err, "GetUserAccessToken")
		retcode = utils.HTTP_CALL_FAILD_EXTERNAL
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
	openid := retJson["openid"].(string)
	httpStr = fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN", accessToken, openid)
	if retBody, err = httpRequest.HttpGetBody(httpStr); err != nil {
		err = errors.Wrap(err, "GetUserAccessToken")
		retcode = utils.HTTP_CALL_FAILD_EXTERNAL
		return
	}
	wxBaseInfoResp := new(WxBaseInfoResp)
	if err = json.Unmarshal(retBody, wxBaseInfoResp); err != nil {
		err = errors.Wrap(err, "GetUserAccessToken")
		retcode = utils.JSON_PARSE_FAILED
		return
	}
	fmt.Println("wxBaseInfoResp: ", *wxBaseInfoResp)
	// 1. 判断微信用户是否已经存在
	// 2. 判断微信用户在该公众号appid下是否已存在
	// 3. 否则，添加相关记录
	if retcode, err = AddAuthorizationUserWxInfo(wxBaseInfoResp, appid); err != nil {
		err = errors.Wrap(err, "GetUserAccessToken")
		return
	}
	return
}

func AddAuthorizationUserWxInfo(base *WxBaseInfoResp, appid string) (retcode int, err error) {
	Logger.Info("[%v] enter AddAuthorizationUserWxInfo.", appid)
	defer Logger.Info("[%v] left AddAuthorizationUserWxInfo.", appid)
	var (
		userWxInfos          []UserWxInfos          = []UserWxInfos{}
		officialAccountUsers []OfficialAccountUsers = []OfficialAccountUsers{}
		num                  int64
		userWxInfoId         int
	)
	o := orm.NewOrm()
	now := time.Now()
	// 1. 判断微信用户是否已经存在
	num, err = o.QueryTable((&UserWxInfos{}).TableName()).Filter("nickname", base.Nickname).Filter("headimgurl", base.Headimgurl).All(&userWxInfos)
	if err != nil {
		err = errors.Wrap(err, "AddAuthorizationUserWxInfo")
		retcode = utils.DB_READ_ERROR
		return
	}
	if num > 0 {
		// 微信用户已存在
		userWxInfoId = userWxInfos[0].Id
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
			UserId:     1, // ::TODO
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
		retcode = utils.DB_READ_ERROR
		return
	}
	if num <= 0 {
		// 新增公众账号的用户
		officialAccounts := []OfficialAccounts{}
		num, err = o.QueryTable((&OfficialAccounts{}).TableName()).Filter("appid", appid).Filter("status", utils.STATUS_VALID).All(&officialAccounts)
		if err != nil {
			err = errors.Wrap(err, "AddAuthorizationUserWxInfo")
			retcode = utils.DB_READ_ERROR
			return
		}
		if num > 0 {
			officialAccount := &OfficialAccountUsers{
				OfficialAccountId: officialAccounts[0].Id,
				UserWxInfoId:      userWxInfoId,
				Openid:            base.Openid,
				Status:            utils.STATUS_VALID,
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
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	o := orm.NewOrm()
	num, err = o.QueryTable((&OfficialAccountUsers{}).TableName()).Filter("official_account_id", officialAccountId).Filter("user_wx_info_id", userWxInfoId).Filter("status", utils.STATUS_VALID).All(&officialAccountUsers)
	if err != nil {
		err = errors.Wrap(err, "GetopenidByCond")
		retcode = utils.DB_READ_ERROR
		return
	}
	if num > 0 {
		openid = officialAccountUsers[0].Openid
	}
	return
}

// GetAllOfficialAccountUsers retrieves all OfficialAccountUsers matches certain condition. Returns empty list if
// no records exist
func GetAllOfficialAccountUsers(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(OfficialAccountUsers))
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

	var l []OfficialAccountUsers
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
