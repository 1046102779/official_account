package controllers

import (
	"strings"

	utils "github.com/1046102779/common"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/pkg/errors"
)

// OfficialAccountUsersController oprations for OfficialAccountUsers
type OfficialAccountUsersController struct {
	beego.Controller
}

// 普通用户微信授权获取access_token
// @router /:id/user/authorization [GET]
func (t *OfficialAccountUsersController) OfficialAccountAuthorizationUser() {
	id, _ := t.GetInt(":id")
	if id <= 0 {
		err := errors.New("param `:id` empty")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": utils.SOURCE_DATA_ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if httpStr, retcode, err := models.OfficialAccountAuthorizationUser(id); err != nil {
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
			"uri":      httpStr,
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
			"err_code": utils.SOURCE_DATA_ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if retcode, err := models.GetUserAccessToken(appid, code); err != nil {
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
