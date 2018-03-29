package controllers

import (
	"github.com/1046102779/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/pkg/errors"
)

// IndustryCodeQuerysController oprations for IndustryCodeQuerys
type IndustryCodeQuerysController struct {
	beego.Controller
}

// 获取主行业列表
// @router /main_industrys [GET]
func (t *IndustryCodeQuerysController) GetAllMainIndutry() {
	if mainIndustrys, retcode, err := models.GetAllMainIndutryNoLocks(); err != nil {
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
			"industry_infos": mainIndustrys,
		}
	}
	t.ServeJSON()
	return
}

// 获取副行业列表
// @router /:id/deputy_industrys [GET]
func (t *IndustryCodeQuerysController) GetSecIndutryListNoLocks() {
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
	industryCodeQuery := &models.IndustryCodeQuerys{
		MainType: id,
	}
	if secIndustrys, retcode, err := industryCodeQuery.GetSecIndutryListNoLocks(); err != nil {
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
			"industry_infos": secIndustrys,
		}
	}
	t.ServeJSON()
	return
}
