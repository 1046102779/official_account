package models

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/orm"

	"github.com/1046102779/common/consts"
	"github.com/1046102779/common/httpRequest"
	"github.com/1046102779/common/types"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"
)

// 公众号授权给公众号第三方平台的返回结果集, 平台通过这些信息，托管所有授权的公众号
type FuncScopeInfo struct {
	Id int `json:"id"`
}

type FuncInfo struct {
	FuncScope FuncScopeInfo `json:"funcscope_category"`
}
type AuthorizedInfo struct {
	Appid        string     `json:"authorizer_appid"`
	AccessToken  string     `json:"authorizer_access_token"`
	ExpiresIn    int        `json:"expires_in"`
	RefreshToken string     `json:"authorizer_refresh_token"`
	FuncInfos    []FuncInfo `json:"func_info"`
}

type AuthorizedInfoResp struct {
	AuthorizedInfo AuthorizedInfo `json:"authorization_info"`
}

type ServiceTypeInfo struct {
	Id int `json:"id"`
}

type VerifyTypeInfo struct {
	Id int `json:"id"`
}

type BusinessInfo struct {
	OpenStore int `json:"open_store"`
	OpenScan  int `json:"open_scan"`
	OpenPay   int `json:"open_pay"`
	OpenCard  int `json:"open_card"`
	OpenShake int `json:"open_shake"`
}

type AuthorizerInfo struct {
	Nickname        string          `json:"nick_name"`
	HeadImg         string          `json:"head_img"`
	ServiceTypeInfo ServiceTypeInfo `json:"service_type_info"`
	VerifyTypeInfo  VerifyTypeInfo  `json:"verify_type_info"`
	UserName        string          `json:"user_name"`
	PrincipalName   string          `json:"principal_name"`
	BusinessInfo    BusinessInfo    `json:"business_info"`
	Alias           string          `json:"alias"`
	QrcodeUrl       string          `json:"qrcode_url"`
}

type AuthorizationInfo struct {
	Appid     string     `json:"appid"`
	FuncInfos []FuncInfo `json:"func_info"`
}

type Authorizer struct {
	AuthorizerInfo    AuthorizerInfo    `json:"authorizer_info"`
	AuthorizationInfo AuthorizationInfo `json:"authorization_info"`
}

func GetAuthorierAccessTokenById(id int, o *orm.Ormer) (token string, retcode int, err error) {
	Logger.Info("[%v] enter GetAuthorierAccessTokenById.", id)
	defer Logger.Info("[%v] left GetAuthorierAccessTokenById.", id)
	offAcc := &OfficialAccounts{
		Id: id,
	}
	if retcode, err = offAcc.ReadOfficialAccountNoLock(o); err != nil {
		err = errors.Wrap(err, "GetAuthorierAccessTokenById")
		return
	}
	if offAcc.Appid != "" {
		var oa *types.OfficialAccount
		if oa, err = conf.WRServerRPC.GetOfficialAccountInfo(offAcc.Appid); err != nil {
			err = errors.Wrap(err, "GetAuthorierAccessTokenById")
		}
		token = oa.AuthorizerAccessToken
	}
	return
}

// 4、使用授权码换取公众号的接口调用凭据和授权信息
// DESC: 通过授权码和自己的接口调用凭据（component_access_token），换取公众号的接口调用凭据
//		（authorizer_access_token和用于前者快过期时用来刷新它的authorizer_refresh_token）和授权信息（授权了哪些权限等信息）
func GetAuthorierTokenInfo(componentAccessToken string, appid string, authorizationCode string) (authorizedInfoResp *AuthorizedInfoResp, retcode int, err error) {
	type ComponentData struct {
		ComponentAppid    string `json:"component_appid"`
		AuthorizationCode string `json:"authorization_code"`
	}
	var (
		componentData *ComponentData = new(ComponentData)
		retBody       []byte
	)
	authorizedInfoResp = new(AuthorizedInfoResp)
	Logger.Info("enter GetAuthorierTokenInfo.")
	defer Logger.Info("left GetAuthorierTokenInfo.")
	if "" == componentAccessToken || "" == authorizationCode {
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		err = errors.New("param `componentAccessToken | authorizationCode`  empty")
		return
	}
	componentData.ComponentAppid = appid
	componentData.AuthorizationCode = authorizationCode
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token=%s", componentAccessToken)
	bodyData, _ := json.Marshal(*componentData)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "GetAuthorierTokenInfo")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, authorizedInfoResp); err != nil {
		err = errors.Wrap(err, "GetAuthorierTokenInfo")
		retcode = consts.ERROR_CODE__JSON__PARSE_FAILED
		return
	}
	// ::TODO 返回的公众号Appid与AuthorizerAccessToken与AuthorizerRefreshToken 保存在etcd， TTL=ExpiresIn-1000， 定时刷新AuthorizerAccessToken
	// 看看公众号赋予了公众号第三方平台哪些权限
	fmt.Println("AppId: "+authorizedInfoResp.AuthorizedInfo.Appid+" allow accesses list: ", getOfficialAccountAccessList(authorizedInfoResp.AuthorizedInfo.FuncInfos))
	return
}

func getOfficialAccountAccessList(funcInfos []FuncInfo) (allowAccesses []string) {
	Logger.Info("enter getOfficialAccountAccessList.")
	defer Logger.Info("left getOfficialAccountAccessList.")
	if funcInfos == nil || len(funcInfos) <= 0 {
		return
	}
	for index := 0; index < len(funcInfos); index++ {
		switch funcInfos[index].FuncScope.Id {
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__MESSAGE_MANAGEMENT:
			allowAccesses = append(allowAccesses, "消息管理权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__USER_MANAGEMENT:
			allowAccesses = append(allowAccesses, "用户管理权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__ACCOUNT_SERVICE:
			allowAccesses = append(allowAccesses, "帐号服务权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__WEB_SERVICE:
			allowAccesses = append(allowAccesses, "网页服务权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__WECHAT_SHOP:
			allowAccesses = append(allowAccesses, "微信小店权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__MULTI_CUSTOMER_SERVICE:
			allowAccesses = append(allowAccesses, "微信多客服权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__NOTIFICATION:
			allowAccesses = append(allowAccesses, "群发与通知权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__CARDS:
			allowAccesses = append(allowAccesses, "微信卡券权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__SCANNING:
			allowAccesses = append(allowAccesses, "微信扫一扫权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__WECHAT_WIFI:
			allowAccesses = append(allowAccesses, "微信连WIFI权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__MATERIAL_MANAGEMENT:
			allowAccesses = append(allowAccesses, "素材管理权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__SHAKE_AROUND:
			allowAccesses = append(allowAccesses, "微信摇周边权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__WECHAT_STORE:
			allowAccesses = append(allowAccesses, "微信门店权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__WECHAT_PAY:
			allowAccesses = append(allowAccesses, "微信支付权限")
		case consts.TYPE_PERMISSION__OFFICIAL_ACCOUNT__CUSTOMIZE_MENU:
			allowAccesses = append(allowAccesses, "自定义菜单权限")
		}
	}
	return
}

// 6、获取授权公众号帐号基本信息
// DESC: 在需要的情况下，第三方平台可以获取公众号的帐号基本信息，
//		包括头像、昵称、帐号名、帐号类型、认证类型、微信号、原始ID和二维码图片URL。
func GetOfficialAccountBaseInfoExternal(componentAccessToken string, appid string, authorizedAppid string) (authorizer *Authorizer, retcode int, err error) {
	type ComponentData struct {
		ComponentAppid  string `json:"component_appid"`
		AuthorizerAppid string `json:"authorizer_appid"`
	}
	var (
		componentData *ComponentData = new(ComponentData)
		retBody       []byte
	)
	Logger.Info("enter GetOfficialAccountBaseInfoExternal.")
	defer Logger.Info("left GetOfficialAccountBaseInfoExternal.")
	if "" == appid || "" == authorizedAppid || "" == componentAccessToken {
		err = errors.New("params `component_appid | authorizer_appid | component_access_token`  empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	authorizer = new(Authorizer)
	componentData.AuthorizerAppid = authorizedAppid
	componentData.ComponentAppid = appid
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_info?component_access_token=%s", componentAccessToken)
	bodyData, _ := json.Marshal(*componentData)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "GetOfficialAccountBaseInfoExternal")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, authorizer); err != nil {
		err = errors.Wrap(err, "GetOfficialAccountBaseInfoExternal")
		retcode = consts.ERROR_CODE__JSON__PARSE_FAILED
		return
	}
	authorizer.AuthorizationInfo.Appid = authorizedAppid
	return
}

// 7、获取授权方的选项设置信息
// DESC: 在需要的情况下，第三方平台可以获取公众号的选项设置，包括地理位置上报设置、语音识别开关设置、微信多客服功能开关设置

// 8、设置授权方的选项信息
// DESC: 在需要的情况下，第三方平台可以修改上述公众号的选项设置，包括地理位置上报设置、语音识别开关设置、微信多客服功能开关设置

// 9、推送授权相关通知
// DESC: 当公众号对第三方进行授权、取消授权、更新授权时，将通过事件推送告诉开发者

// 接下来：代替公众号调用接口
// DESC: 取在完成授权后，第三方平台可通过公众号的接口调用凭据（authorizer_access_token）来代替它调用接口，具体请见“代公众号实现业务”文件夹中的内容
