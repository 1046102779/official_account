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

	"github.com/1046102779/common/consts"
	"github.com/1046102779/common/types"
	"github.com/1046102779/common/utils"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
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
	cvt := &types.ComponentVerifyTicket{
		TimeStamp: t.GetString("timestamp"),
		Nonce:     t.GetString("nonce"),
		MsgSign:   t.GetString("msg_signature"),
		Bts:       t.Ctx.Input.RequestBody,
	}
	conf.WRServerRPC.RefreshComponentVerifyTicket(cvt)
	t.Ctx.Output.Body([]byte("success"))
	return
}

// 引入用户进入授权页
// @router /authorization/loginpage [GET]
func (t *Authorized) GetComponentLoginPage() {
	redirectUrl := t.GetString("callback_url")
	var oap *types.OfficialAccountPlatform
	oap, _ = conf.WRServerRPC.GetOfficialAccountPlatformInfo()
	httpStr := fmt.Sprintf(
		"https://mp.weixin.qq.com/cgi-bin/componentloginpage?component_appid=%s&pre_auth_code=%s&redirect_uri=%s",
		oap.Appid,
		oap.PreAuthCode,
		url.QueryEscape(redirectUrl),
	)
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
		offAcc             *models.OfficialAccounts = new(models.OfficialAccounts)
		companyId, retCode int                      = 1, 0
		err                error
	)
	if !conf.WechatOpenPlatformTestFeatureFlag {
		// 获取登录态的公司ID
		companyId, retCode, err = utils.GetCompanyIdFromHeader(t.Ctx.Request)
		if err != nil {
			Logger.Error(err.Error())
			t.Data["json"] = map[string]interface{}{
				"err_code": retCode,
				"err_msg":  errors.Cause(err).Error(),
			}
			t.ServeJSON()
			return
		}
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
	var oap *types.OfficialAccountPlatform
	if oap, err = conf.WRServerRPC.GetOfficialAccountPlatformInfo(); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__GRPC__FAILED,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	platformAppid := oap.Appid
	authorizedInfoResp, retCode, err := models.GetAuthorierTokenInfo(oap.ComponentAccessToken, oap.Appid, code)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retCode,
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
		oa := &types.OfficialAccount{
			Appid: authorizedInfoResp.AuthorizedInfo.Appid,
			AuthorizerAccessToken:          authorizedInfoResp.AuthorizedInfo.AccessToken,
			AuthorizerAccessTokenExpiresIn: int64(authorizedInfoResp.AuthorizedInfo.ExpiresIn),
			AuthorizerRefreshToken:         authorizedInfoResp.AuthorizedInfo.RefreshToken,
		}
		conf.WRServerRPC.StoreOfficialAccountInfo(oa)
	}
	// 获取公众号基本信息
	offAcc, retCode, err = models.GetOfficialAccountBaseInfo(platformAppid, authorizedInfoResp.AuthorizedInfo.Appid)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retCode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	fmt.Println("authorizedInfoResp: ", *authorizedInfoResp)
	// 公众号与公司ID的绑定

	retCode, err = models.BindingCompanyAndOfficialAccount(offAcc.Id, companyId)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retCode,
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
