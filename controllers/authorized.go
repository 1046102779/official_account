package controllers

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"

	utils "github.com/1046102779/common"
	"github.com/1046102779/common/httpRequest"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/chanxuehong/util"
	"github.com/gomydodo/wxencrypter"
	"github.com/pkg/errors"
)

type Authorized struct {
	beego.Controller
}

type XmlResp struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
}

var (
	firstInitital bool               = false
	etcdClient    *models.EtcdClient = new(models.EtcdClient)
)

// 授权事件接收URL
// DESC: 出于安全考虑，在第三方平台创建审核通过后，微信服务器每隔10分钟会向第三方的消息接收地址
//		 推送一次component_verify_ticket，用于获取第三方平台接口调用凭据
// example request:
// <xml>
// <AppId> </AppId>
// <CreateTime>1413192605 </CreateTime>
// <InfoType> </InfoType>
// <ComponentVerifyTicket> </ComponentVerifyTicket>
// </xml>
// @router / [POST]
func (t *Authorized) ComponentVerifyTicket() {
	type ComponentVerifyTicketReq struct {
		AppId                 string `xml:"AppId"`
		CreateTime            string `xml:"CreateTime"`
		InfoType              string `xml:"InfoType"`
		ComponentVerifyTicket string `xml:"ComponentVerifyTicket"`
	}
	var (
		req *ComponentVerifyTicketReq = new(ComponentVerifyTicketReq)
	)
	timestamp := t.GetString("timestamp")
	nonce := t.GetString("nonce")
	msgSignature := t.GetString("msg_signature")
	e, err := wxencrypter.NewEncrypter(conf.WechatParam.Token, conf.WechatParam.EncodingAesKey, conf.WechatParam.AppId)
	if err != nil {
		Logger.Error("NewEncrypter failed. " + err.Error())
	}
	b, err := e.Decrypt(msgSignature, timestamp, nonce, t.Ctx.Input.RequestBody)
	if err != nil {
		Logger.Error("Decrypt failed. " + err.Error())
	}
	fmt.Println("ticket body: ", string(b))
	if err := xml.Unmarshal(b, req); err != nil {
		Logger.Error(err.Error())
	}
	// 全网发布测试代码集合
	reader := strings.NewReader(string(b))
	reqMap, err := util.DecodeXMLToMap(reader)
	if reqMap["InfoType"] == "authorized" {
		//appid := reqMap["AuthorizerAppid"]
		conf.QueryAuthCodeTest = reqMap["AuthorizationCode"]
		httpStr := fmt.Sprintf("%s/v1/wechats/authorization/code?auth_code=%s", conf.HostName, conf.QueryAuthCodeTest)
		_, err = httpRequest.HttpGetBody(httpStr)
		if err != nil {
			Logger.Error("get authorizer access token failed. " + err.Error())
		}
		t.Ctx.Output.Body([]byte("success"))
		return
	} else if reqMap["InfoType"] == "unauthorized" {
		t.Ctx.Output.Body([]byte("success"))
		return
	}
	// end

	fmt.Println("wechat info: ", *req)
	conf.WechatAuthTTL.ComponentVerifyTicket = req.ComponentVerifyTicket
	if !firstInitital || conf.WechatAuthTTL.ComponentAccessToken == "" || conf.WechatAuthTTL.PreAuthCode == "" {
		firstInitital = true
		fmt.Println("appid:" + conf.WechatParam.AppId + " | appsecret:" + conf.WechatParam.AppSecret + " | ticket:" + conf.WechatAuthTTL.ComponentVerifyTicket)
		// Set Key: ComponentAccessToken
		conf.WechatAuthTTL.ComponentAccessToken, conf.WechatAuthTTL.ComponentAccessTokenExpiresIn, _, err = models.GetComponentAccessToken(conf.WechatParam.AppId, conf.WechatParam.AppSecret, conf.WechatAuthTTL.ComponentVerifyTicket)
		if err != nil {
			Logger.Error("get param `component_access_token | expires_in` failed. " + err.Error())
		} else {
			_, err = etcdClient.Put(fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, conf.ListenPaths[1]), conf.WechatAuthTTL.ComponentAccessToken, conf.WechatAuthTTL.ComponentAccessTokenExpiresIn)
			if err != nil {
				Logger.Error(err.Error())
			}
		}
		go etcdClient.Watch(fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, conf.ListenPaths[1])) // ComponentAccessToken

		conf.WechatAuthTTL.PreAuthCode, conf.WechatAuthTTL.PreAuthCodeExpiresIn, _, err = models.GetPreAuthCode(conf.WechatAuthTTL.ComponentAccessToken)
		if err != nil {
			Logger.Error("get param `pre_auth_code` failed. " + err.Error())
		} else {
			_, err = etcdClient.Put(fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, conf.ListenPaths[2]), conf.WechatAuthTTL.PreAuthCode, conf.WechatAuthTTL.PreAuthCodeExpiresIn)
			if err != nil {
				Logger.Error(err.Error())
			}
		}
		go etcdClient.Watch(fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, conf.ListenPaths[2])) // PreAuthCode
		fmt.Println("hello,world")
	}
	t.Ctx.Output.Body([]byte("success"))
	return
}

// @router  /authorization [POST]
func (t *Authorized) Authorization() {
	if "" == conf.WechatAuthTTL.ComponentVerifyTicket {
		Logger.Error("param `component_verify_ticket` empty, waiting for 10 mininutes")
		t.Data["json"] = map[string]interface{}{
			"err_code": -1,
			"err_msg":  "param `component_verify_ticket` empty, waiting for 10 mininutes",
		}
		t.ServeJSON()
		return
	}
	fmt.Println("hello,world")
	componentAccessToken, _, retcode, err := models.GetComponentAccessToken(conf.WechatParam.AppId, conf.WechatParam.AppSecret, conf.WechatAuthTTL.ComponentVerifyTicket)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	fmt.Println("token: " + componentAccessToken)
	t.Data["json"] = map[string]interface{}{
		"err_code": 0,
		"err_msg":  "",
	}
	t.ServeJSON()
	return
}

// 引入用户进入授权页
// @router /authorization/loginpage [GET]
func (t *Authorized) GetComponentLoginPage() {
	var (
		preAuthCode string                                                                        // 预授权码
		redirectUrl string = fmt.Sprintf("%s%s", conf.HostName, "/v1/wechats/authorization/code") // 回调URI
	)
	// 从etcd获取pre_auth_code预授权码
	maps, _, _ := etcdClient.Get(fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, conf.ListenPaths[2]))
	if maps != nil && len(maps) > 0 {
		for _, value := range maps {
			preAuthCode = value
		}
	}
	httpStr := fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=%s", conf.WechatParam.AppId, preAuthCode, url.QueryEscape(redirectUrl))
	/*
		t.Data["json"] = map[string]interface{}{
			"err_code": 0,
			"err_msg":  "",
			"uri":      fmt.Sprintf("<link>%s</link>", httpStr),
		}
		t.ServeJSON()
	*/
	t.Ctx.Output.Body([]byte(fmt.Sprintf("<a href=%s>Link text</a>", httpStr)))
	return
}

// 使用授权码换取公众号的接口调用凭据和授权信息
// @router /authorization/code [GET]
func (t *Authorized) GetAuthorizedCode() {
	var (
		offAcc *models.OfficialAccounts = new(models.OfficialAccounts)
	)
	code := t.GetString("auth_code")
	if "" == strings.TrimSpace(code) {
		err := errors.New("param `auth_code` empty")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": utils.SOURCE_DATA_ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if conf.WechatAuthTTL.ComponentAccessToken == "" || conf.WechatAuthTTL.PreAuthCode == "" {
		maps, _, _ := etcdClient.Get(fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, conf.ListenPaths[1]))
		for _, value := range maps {
			conf.WechatAuthTTL.ComponentAccessToken = value
		}
		maps, _, _ = etcdClient.Get(fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, conf.ListenPaths[2]))
		for _, value := range maps {
			conf.WechatAuthTTL.PreAuthCode = value
		}
	}
	authorizedInfoResp, retcode, err := models.GetAuthorierTokenInfo(conf.WechatAuthTTL.ComponentAccessToken, code)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if authorizedInfoResp != nil {
		appid := authorizedInfoResp.AuthorizedInfo.Appid
		conf.WechatAuthTTL.AuthorizerMap[appid] = conf.AuthorizerManagementInfo{
			AuthorizerAccessToken:          authorizedInfoResp.AuthorizedInfo.AccessToken,
			AuthorizerAccessTokenExpiresIn: authorizedInfoResp.AuthorizedInfo.ExpiresIn,
			AuthorizerRefreshToken:         authorizedInfoResp.AuthorizedInfo.RefreshToken,
		}
	}
	if authorizedInfoResp.AuthorizedInfo.Appid == "" {
		Logger.Error("authorized code invalid. please authorized again!")
		t.Data["json"] = map[string]interface{}{
			"err_code": -1,
			"err_msg":  "authorized code invalid. please authorized again!",
		}
		t.ServeJSON()
		return
	}
	// 查询appid key是否存在，不存在则set并watch
	fields := strings.Split(conf.ListenPaths[0], "/")
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", beego.BConfig.RunMode, fields[1], fields[2], authorizedInfoResp.AuthorizedInfo.Appid, fields[3])
	maps, _, _ := etcdClient.Get(key)
	retcode, err = etcdClient.Put(key, conf.WechatAuthTTL.AuthorizerMap[authorizedInfoResp.AuthorizedInfo.Appid].AuthorizerAccessToken, conf.WechatAuthTTL.AuthorizerMap[authorizedInfoResp.AuthorizedInfo.Appid].AuthorizerAccessTokenExpiresIn)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if maps == nil || len(maps) <= 0 {
		go etcdClient.Watch(key)
	}
	// 获取公众号基本信息
	offAcc, retcode, err = models.GetOfficialAccountBaseInfo(authorizedInfoResp.AuthorizedInfo.Appid)
	fmt.Println("authorizedInfoResp: ", *authorizedInfoResp)
	t.Data["json"] = map[string]interface{}{
		"err_code":  0,
		"err_msg":   "",
		"base_info": *offAcc,
	}
	t.ServeJSON()
	return
}
