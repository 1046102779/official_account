// 公众号第三方平台服务列表
// 1. 使用授权码换取公众号的接口调用凭据和授权信息
// 2. 引入用户进入授权页
// 3. 授权事件接收URL
// 4.
package controllers

import (
	"fmt"
	"net/url"
	"strings"

	"git.kissdata.com/ycfm/common/consts"
	"git.kissdata.com/ycfm/common/utils"
	pb "git.kissdata.com/ycfm/igrpc"
	"git.kissdata.com/ycfm/official_account/conf"
	. "git.kissdata.com/ycfm/official_account/logger"
	"git.kissdata.com/ycfm/official_account/models"
	"github.com/astaxie/beego"
	"github.com/pkg/errors"
)

type Authorized struct {
	beego.Controller
}

type XmlResp struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
}

// 授权事件接收URL
// DESC: 出于安全考虑，在第三方平台创建审核通过后，微信服务器每隔10分钟会向第三方的消息接收地址
//       推送一次component_verify_ticket，用于获取第三方平台接口调用凭据
// example request:
// <xml>
// <AppId> </AppId>
// <CreateTime>1413192605 </CreateTime>
// <InfoType> </InfoType>
// <ComponentVerifyTicket> </ComponentVerifyTicket>
// </xml>
// @router / [POST]
func (t *Authorized) ComponentVerifyTicket() {
	in := &pb.ComponentVerifyTicket{
		TimeStamp: t.GetString("timestamp"),
		Nonce:     t.GetString("nonce"),
		MsgSign:   t.GetString("msg_signature"),
		Bts:       t.Ctx.Input.RequestBody,
	}
	conf.WxRelayServerClient.Call(fmt.Sprintf("%s.%s", "wx_relay_server", "RefreshComponentVerifyTicket"), in, in)
	t.Ctx.Output.Body([]byte("success"))
	return
}

// 引入用户进入授权页
// @router /authorization/loginpage [GET]
func (t *Authorized) GetComponentLoginPage() {
	redirectUrl := t.GetString("callback_url")
	in := &pb.OfficialAccountPlatform{}
	conf.WxRelayServerClient.Call(fmt.Sprintf("%s.%s", "wx_relay_server", "GetOfficialAccountPlatformInfo"), in, in)
	httpStr := fmt.Sprintf("https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=%s", in.Appid, in.PreAuthCode, url.QueryEscape(redirectUrl))
	t.Data["json"] = map[string]interface{}{
		"err_code": 0,
		"err_msg":  "",
		"uri":      httpStr,
	}
	t.ServeJSON()
	return
}

// 使用授权码换取公众号的接口调用凭据和授权信息
// @router /authorization/code [GET]
func (t *Authorized) GetAuthorizedCode() {
	var (
		offAcc    *models.OfficialAccounts = new(models.OfficialAccounts)
		companyId int                      = 0
	)
	// 获取登录态的公司ID
	companyId, retcode, err := utils.GetCompanyIdFromHeader(t.Ctx.Request)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	code := t.GetString("auth_code")
	if "" == strings.TrimSpace(code) {
		err := errors.New("param `auth_code` empty")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__SOURCE_DATA__ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	// ::TODO get
	in := &pb.OfficialAccountPlatform{}
	conf.WxRelayServerClient.Call(fmt.Sprintf("%s.%s", "wx_relay_server", "GetOfficialAccountPlatformInfo"), in, in)
	platformAppid := in.Appid
	authorizedInfoResp, retcode, err := models.GetAuthorierTokenInfo(in.ComponentAccessToken, in.Appid, code)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if authorizedInfoResp == nil || authorizedInfoResp.AuthorizedInfo.Appid == "" {
		Logger.Error("authorized code invalid. please authorized again!")
		t.Data["json"] = map[string]interface{}{
			"err_code": -1,
			"err_msg":  "authorized code invalid. please authorized again!",
		}
		t.ServeJSON()
		return
	} else {
		in := &pb.OfficialAccount{
			Appid: authorizedInfoResp.AuthorizedInfo.Appid,
			AuthorizerAccessToken:          authorizedInfoResp.AuthorizedInfo.AccessToken,
			AuthorizerAccessTokenExpiresIn: int64(authorizedInfoResp.AuthorizedInfo.ExpiresIn),
			AuthorizerRefreshToken:         authorizedInfoResp.AuthorizedInfo.RefreshToken,
		}
		conf.WxRelayServerClient.Call(fmt.Sprintf("%s.%s", "wx_relay_server", "StoreOfficialAccountInfo"), in, in)
	}
	// 获取公众号基本信息
	offAcc, retcode, err = models.GetOfficialAccountBaseInfo(platformAppid, authorizedInfoResp.AuthorizedInfo.Appid)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	fmt.Println("authorizedInfoResp: ", *authorizedInfoResp)
	// 公众号与公司ID的绑定

	retcode, err = models.BindingCompanyAndOfficialAccount(offAcc.Id, companyId)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	t.Data["json"] = map[string]interface{}{
		"err_code":  0,
		"err_msg":   "",
		"base_info": *offAcc,
	}
	t.ServeJSON()
	return
}
