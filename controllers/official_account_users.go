// 公众号下的用户服务列表
// 1. 前端获取微信公众号需要的有效用户授权url
// 2. 微信用户是否已经绑定(备注：前端获取有效url后，访问微信公众号。若有效，则微信公众号会回调, 前端解析url，会拿到appid和拿AccessToken所需要的code)
package controllers

import (
	"strings"

	"github.com/1046102779/official_account/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/pkg/errors"
)

// OfficialAccountUsersController oprations for OfficialAccountUsers
type OfficialAccountUsersController struct {
	beego.Controller
}

// 自测使用，获取微信回调的appid和code
// @router /callback [GET]
func (t *OfficialAccountUsersController) GetWeChatCallback() {
	appid := t.GetString("appid")
	code := t.GetString("code")
	t.Data["json"] = map[string]interface{}{
		"err_code": 0,
		"err_msg":  "",
		"appid":    appid,
		"code":     code,
	}
	t.ServeJSON()
	return
}

// 1. 前端获取微信公众号需要的有效用户授权url
// @router /:id/user/authorization [GET]
func (t *OfficialAccountUsersController) OfficialAccountAuthorizationUser() {
	id, _ := t.GetInt(":id")
	if id <= 0 {
		err := errors.New("param `:id` empty")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__SOURCE_DATA__ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	callbackUrl := t.GetString("callback_url")
	if callbackUrl == "" {
		err := errors.New("param `callback_url` empty")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__PARAM__ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if httpUri, retcode, err := models.OfficialAccountAuthorizationUser(id, callbackUrl); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	} else {
		t.Data["json"] = map[string]interface{}{
			"err_code": 0,
			"err_msg":  "",
			"uri":      httpUri,
		}
	}
	t.ServeJSON()
	return
}

// 2. 微信用户是否已经绑定(备注：前端获取有效url后，访问微信公众号。若有效，则微信公众号会回调, 前端解析url，会拿到appid和拿AccessToken所需要的code)
// @router /users/binding [GET]
func (t *OfficialAccountUsersController) GetWechatUserBinding() {
	appid := t.GetString("appid")
	code := t.GetString("code")
	if "" == strings.TrimSpace("code") || "" == strings.TrimSpace("appid") {
		err := errors.New("param `code | appid` empty")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__SOURCE_DATA__ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if bindingStatus, user, customer, openid, retcode, err := models.GetUserAccessToken(appid, code); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	} else {
		t.Data["json"] = map[string]interface{}{
			"err_code":       0,
			"err_msg":        "",
			"open_id":        openid,
			"binding_status": bindingStatus,
			"user":           user,
			"customer":       customer,
		}
	}
	t.ServeJSON()
	return
}

// 通过code换取access_token
// @router /user/authorization/callback [GET]
func (t *OfficialAccountUsersController) GetUserAccessToken() {
	code := t.GetString("code")
	appid := t.GetString("appid")
	if "" == strings.TrimSpace("code") || "" == strings.TrimSpace("appid") {
		err := errors.New("param `code | appid` empty")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__SOURCE_DATA__ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if _, _, _, _, retcode, err := models.GetUserAccessToken(appid, code); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	t.Data["json"] = map[string]interface{}{
		"err_code": 0,
		"err_msg":  "",
	}
	t.ServeJSON()
	return
}
