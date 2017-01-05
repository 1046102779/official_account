package models

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/orm"

	utils "github.com/1046102779/common"
	"github.com/1046102779/common/httpRequest"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"
)

const (
	OFFICIAL_ACCOUNT_ACCESS_MESSAGE_MANAGEMENT     = iota + 1 // 消息管理权限
	OFFICIAL_ACCOUNT_ACCESS_USER_MANAGEMENT                   // 用户管理权限
	OFFICIAL_ACCOUNT_ACCESS_ACCOUNT_SERVICE                   // 帐号服务权限
	OFFICIAL_ACCOUNT_ACCESS_WEB_SERVICE                       // 网页服务权限
	OFFICIAL_ACCOUNT_ACCESS_WECHAT_SHOP                       // 微信小店权限
	OFFICIAL_ACCOUNT_ACCESS_MULTI_CUSTOMER_SERVICE            // 微信多客服权限
	OFFICIAL_ACCOUNT_ACCESS_NOTIFICATION                      // 群发与通知权限
	OFFICIAL_ACCOUNT_ACCESS_CARDS                             // 微信卡券权限
	OFFICIAL_ACCOUNT_ACCESS_SCANNING                          // 微信扫一扫权限
	OFFICIAL_ACCOUNT_ACCESS_WECHAT_WIFI                       // 微信连WIFI权限
	OFFICIAL_ACCOUNT_ACCESS_MATERIAL_MANAGEMENT               // 素材管理权限
	OFFICIAL_ACCOUNT_ACCESS_SHAKE_AROUND                      // 微信摇周边权限
	OFFICIAL_ACCOUNT_ACCESS_WECHAT_STORE                      // 微信门店权限
	OFFICIAL_ACCOUNT_ACCESS_WECHAT_PAY                        // 微信支付权限
	OFFICIAL_ACCOUNT_ACCESS_CUSTOMIZE_MENU                    // 自定义菜单权限
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

/*
func init() {
	// 创建相关目录，并监听公众号第三方平台的相关token参数，达到定时更新授权token的目的
	// 监听的参数包括：
	// 1. ComponentVerifyTicket 用于获取第三方平台接口调用凭
	// 2. ComponentAccessToken  第三方平台的下文中接口的调用凭据
	// 3. PreAuthCode           预授权码, 获取公众号第三方平台授权页面
	for index := 0; index < len(conf.ListenPaths); index++ {
		if "" != conf.ListenPaths[index] {
			go etcdClient.Watch(conf.ListenPaths[index])
		}
	}
}
*/

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
		if _, ok := conf.WechatAuthTTL.AuthorizerMap[offAcc.Appid]; !ok {
			err = errors.New("param `appid` not exists in maps")
			retcode = utils.SOURCE_DATA_ILLEGAL
			return
		}
		token = conf.WechatAuthTTL.AuthorizerMap[offAcc.Appid].AuthorizerAccessToken
	}
	return
}

// 2、获取第三方平台component_access_token
// DESC: 第三方平台通过自己的component_appid和component_appsecret
//		（即在微信开放平台管理中心的第三方平台详情页中的AppID和AppSecret），以及component_verify_ticket
//		 来获取自己的接口调用凭据（component_access_token）
func GetComponentAccessToken(componentAppid string, componentAppseret string, componentVerifyTicket string) (componentAccessToken string, expiresIn int, retcode int, err error) {
	type ComponentData struct {
		ComponentAppid        string `json:"component_appid"`
		ComponentAppsecret    string `json:"component_appsecret"`
		ComponentVerifyTicket string `json:"component_verify_ticket"`
	}
	type ComponentAccessTokenResp struct {
		ComponentAccessToken string `json:"component_access_token"`
		ExpiresIn            int    `json:"expires_in"`
	}
	var (
		componentData *ComponentData            = new(ComponentData)
		tokenResp     *ComponentAccessTokenResp = new(ComponentAccessTokenResp)
		retBody       []byte
	)
	Logger.Info("enter GetComponentAccessToken.")
	defer Logger.Info("left GetComponentAccessToken.")
	if "" == componentAppid || "" == componentAppseret || "" == componentVerifyTicket {
		retcode = utils.SOURCE_DATA_ILLEGAL
		err = errors.New("param `componentAppid | componentAppseret | componentVerifyTicket` empty")
		return
	}
	componentData.ComponentAppid = componentAppid
	componentData.ComponentAppsecret = componentAppseret
	componentData.ComponentVerifyTicket = componentVerifyTicket
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_component_token")
	bodyData, _ := json.Marshal(*componentData)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "GetComponentAccessToken")
		retcode = utils.HTTP_CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, tokenResp); err != nil {
		err = errors.Wrap(err, "GetComponentAccessToken")
		retcode = utils.JSON_PARSE_FAILED
		return
	}
	// expires_in: 两小时
	return tokenResp.ComponentAccessToken, tokenResp.ExpiresIn, 0, nil
}

// 3、获取预授权码pre_auth_code
// DESC: 第三方平台通过自己的接口调用凭据（component_access_token）来获取用于授权流程准备的预授权码（pre_auth_code）
func GetPreAuthCode(componentAccessToken string) (preAuthCode string, expiresIn int, retcode int, err error) {
	type PreAuthCodeInfo struct {
		PreAuthCode string `json:"pre_auth_code"`
		ExpiresIn   int    `json:"expires_in"`
	}
	type ComponentData struct {
		ComponentAppid string `json:"component_appid"`
	}
	var (
		componentData   *ComponentData   = new(ComponentData)
		preAuthCodeInfo *PreAuthCodeInfo = new(PreAuthCodeInfo)
		retBody         []byte
	)
	Logger.Info("enter GetPreAuthCode.")
	defer Logger.Info("left GetPreAuthCode.")
	if "" == componentAccessToken {
		retcode = utils.SOURCE_DATA_ILLEGAL
		err = errors.New("param `componentAccessToken` empty")
		return
	}
	componentData.ComponentAppid = conf.WechatParam.AppId
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token=%s", componentAccessToken)
	bodyData, _ := json.Marshal(*componentData)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "GetPreAuthCode")
		retcode = utils.HTTP_CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, preAuthCodeInfo); err != nil {
		err = errors.Wrap(err, "GetPreAuthCode")
		retcode = utils.JSON_PARSE_FAILED
		return
	}
	// expire_in: 20分钟
	return preAuthCodeInfo.PreAuthCode, preAuthCodeInfo.ExpiresIn, 0, nil
}

// 4、使用授权码换取公众号的接口调用凭据和授权信息
// DESC: 通过授权码和自己的接口调用凭据（component_access_token），换取公众号的接口调用凭据
//		（authorizer_access_token和用于前者快过期时用来刷新它的authorizer_refresh_token）和授权信息（授权了哪些权限等信息）
func GetAuthorierTokenInfo(componentAccessToken string, authorizationCode string) (authorizedInfoResp *AuthorizedInfoResp, retcode int, err error) {
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
		retcode = utils.SOURCE_DATA_ILLEGAL
		err = errors.New("param `componentAccessToken | authorizationCode`  empty")
		return
	}
	componentData.ComponentAppid = conf.WechatParam.AppId
	componentData.AuthorizationCode = authorizationCode
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_query_auth?component_access_token=%s", componentAccessToken)
	bodyData, _ := json.Marshal(*componentData)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "GetAuthorierTokenInfo")
		retcode = utils.HTTP_CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, authorizedInfoResp); err != nil {
		err = errors.Wrap(err, "GetAuthorierTokenInfo")
		retcode = utils.JSON_PARSE_FAILED
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
		case OFFICIAL_ACCOUNT_ACCESS_MESSAGE_MANAGEMENT:
			allowAccesses = append(allowAccesses, "消息管理权限")
		case OFFICIAL_ACCOUNT_ACCESS_USER_MANAGEMENT:
			allowAccesses = append(allowAccesses, "用户管理权限")
		case OFFICIAL_ACCOUNT_ACCESS_ACCOUNT_SERVICE:
			allowAccesses = append(allowAccesses, "帐号服务权限")
		case OFFICIAL_ACCOUNT_ACCESS_WEB_SERVICE:
			allowAccesses = append(allowAccesses, "网页服务权限")
		case OFFICIAL_ACCOUNT_ACCESS_WECHAT_SHOP:
			allowAccesses = append(allowAccesses, "微信小店权限")
		case OFFICIAL_ACCOUNT_ACCESS_MULTI_CUSTOMER_SERVICE:
			allowAccesses = append(allowAccesses, "微信多客服权限")
		case OFFICIAL_ACCOUNT_ACCESS_NOTIFICATION:
			allowAccesses = append(allowAccesses, "群发与通知权限")
		case OFFICIAL_ACCOUNT_ACCESS_CARDS:
			allowAccesses = append(allowAccesses, "微信卡券权限")
		case OFFICIAL_ACCOUNT_ACCESS_SCANNING:
			allowAccesses = append(allowAccesses, "微信扫一扫权限")
		case OFFICIAL_ACCOUNT_ACCESS_WECHAT_WIFI:
			allowAccesses = append(allowAccesses, "微信连WIFI权限")
		case OFFICIAL_ACCOUNT_ACCESS_MATERIAL_MANAGEMENT:
			allowAccesses = append(allowAccesses, "素材管理权限")
		case OFFICIAL_ACCOUNT_ACCESS_SHAKE_AROUND:
			allowAccesses = append(allowAccesses, "微信摇周边权限")
		case OFFICIAL_ACCOUNT_ACCESS_WECHAT_STORE:
			allowAccesses = append(allowAccesses, "微信门店权限")
		case OFFICIAL_ACCOUNT_ACCESS_WECHAT_PAY:
			allowAccesses = append(allowAccesses, "微信支付权限")
		case OFFICIAL_ACCOUNT_ACCESS_CUSTOMIZE_MENU:
			allowAccesses = append(allowAccesses, "自定义菜单权限")
		}
	}
	return
}

// 5、获取（刷新）授权公众号的接口调用凭据
// DESC: 通过authorizer_refresh_token来刷新公众号的接口调用凭据
func RefreshToken(componentAccessToken string, authorizerRefreshToken string, authorizerAppid string) (authorizerAccessToken string, expiresIn int, authorizerRefreshTokenNew string, retcode int, err error) {
	type ComponentData struct {
		ComponentAppid         string `json:"component_appid"`
		AuthorizerAppid        string `json:"authorizer_appid"`
		AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
	}
	type authorizedInfoResp struct {
		AccessToken  string `json:"authorizer_access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"authorizer_refresh_token"`
	}
	var (
		componentData  *ComponentData      = new(ComponentData)
		authorizerInfo *authorizedInfoResp = new(authorizedInfoResp)
		retBody        []byte
	)
	Logger.Info("enter RefreshToken.")
	defer Logger.Info("left RefreshToken.")
	if "" == authorizerAppid || "" == conf.WechatParam.AppId || "" == componentAccessToken || "" == authorizerRefreshToken {
		retcode = utils.SOURCE_DATA_ILLEGAL
		err = errors.New("params `authorizerAppid | component_appid | component_access_token | authorizer_refresh_token` empty")
		return
	}
	componentData.ComponentAppid = conf.WechatParam.AppId
	componentData.AuthorizerAppid = authorizerAppid
	componentData.AuthorizerRefreshToken = authorizerRefreshToken
	bodyData, _ := json.Marshal(*componentData)
	httpStr := fmt.Sprintf("https:// api.weixin.qq.com /cgi-bin/component/api_authorizer_token?component_access_token=%s", componentAccessToken)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "RefreshToken")
		retcode = utils.HTTP_CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, authorizerInfo); err != nil {
		err = errors.Wrap(err, "RefreshToken")
		retcode = utils.JSON_PARSE_FAILED
		return
	}
	authorizerAccessToken = authorizerInfo.AccessToken
	expiresIn = authorizerInfo.ExpiresIn
	authorizerRefreshTokenNew = authorizerInfo.RefreshToken
	return
}

// 6、获取授权公众号帐号基本信息
// DESC: 在需要的情况下，第三方平台可以获取公众号的帐号基本信息，
//		包括头像、昵称、帐号名、帐号类型、认证类型、微信号、原始ID和二维码图片URL。
func GetOfficialAccountBaseInfoExternal(componentAccessToken string, authorizedAppid string) (authorizer *Authorizer, retcode int, err error) {
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
	if "" == conf.WechatParam.AppId || "" == authorizedAppid || "" == componentAccessToken {
		err = errors.New("params `component_appid | authorizer_appid | component_access_token`  empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	authorizer = new(Authorizer)
	componentData.AuthorizerAppid = authorizedAppid
	componentData.ComponentAppid = conf.WechatParam.AppId
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_get_authorizer_info?component_access_token=%s", componentAccessToken)
	bodyData, _ := json.Marshal(*componentData)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "GetOfficialAccountBaseInfoExternal")
		retcode = utils.HTTP_CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, authorizer); err != nil {
		err = errors.Wrap(err, "GetOfficialAccountBaseInfoExternal")
		retcode = utils.JSON_PARSE_FAILED
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
