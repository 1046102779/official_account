// 公众号服务列表
// 1. 获取已托管的公众号列表
// 2. 获取公众号基本信息
package controllers

import (
	"github.com/1046102779/official_account/common/utils"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/pkg/errors"
)

// OfficialAccountsController oprations for OfficialAccounts
type OfficialAccountsController struct {
	beego.Controller
}

// 1. 获取已托管的公众号列表
// @router /official_accounts [GET]
func (t *OfficialAccountsController) GetOfficialAccounts() {
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
	officialAccounts, retcode, err := models.GetOfficialAccounts(companyId)
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
		"err_code":          0,
		"err_msg":           "",
		"official_accounts": officialAccounts,
	}
	t.ServeJSON()
	return
}
