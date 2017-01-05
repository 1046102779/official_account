package controllers

import (
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/pkg/errors"
)

// OfficialAccountsController oprations for OfficialAccounts
type OfficialAccountsController struct {
	beego.Controller
}

// 获取公众账号基本信息
// @router /baseinfo/:appid [GET]
func (t *OfficialAccountsController) GetOfficialAccountBaseInfo() {
	appid := t.GetString(":appid")
	offAcc, retcode, err := models.GetOfficialAccountBaseInfo(appid)
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
