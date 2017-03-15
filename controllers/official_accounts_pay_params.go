// 微信支付参数服务列表
// 1. 新增公众号支付开发参数
// 2. 获取公众号支付开发参数
// 3. 上传证书 证书包括apiclient_key.pem和apiclient_cert.pem
package controllers

import (
	"github.com/1046102779/official_account/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
)

// OfficialAccountsPayParamsController oprations for OfficialAccountsPayParams
type OfficialAccountsPayParamsController struct {
	beego.Controller
}

// 上传证书 证书包括apiclient_key.pem和apiclient_cert.pem
// @params :id 表示内部公众号ID
// @router /:id/certification [POST]
func (t *OfficialAccountsPayParamsController) UploadCertification() {
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
	if retcode, err := models.UploadCertification(id, t.Ctx.Request); err != nil {
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

type PayParamInfo struct {
	Appkey string `json:"appkey"`
	MchId  string `json:"mch_id"`
	Name   string `json:"name"`
}

// 1. 新增公众号支付开发参数
// @router /:id/payparams [POST]
func (t *OfficialAccountsPayParamsController) ModifyWechatParams() {
	var (
		payParamInfo *PayParamInfo = new(PayParamInfo)
	)
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
	if err := jsoniter.Unmarshal(t.Ctx.Input.RequestBody, payParamInfo); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__JSON__PARSE_FAILED,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if retcode, err := models.ModifyWechatParams(id, payParamInfo.Appkey, payParamInfo.MchId, payParamInfo.Name); err != nil {
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

// 2. 获取公众号支付开发参数
// @router /:id/payparams [GET]
func (t *OfficialAccountsPayParamsController) GetOfficialAccountPayParamByOfficialAccountId() {
	type PayParamInfo struct {
		Appkey string `json:"appkey"`
		MchId  string `json:"mch_id"`
		Name   string `json:"name"`
	}
	var (
		payParamInfo *PayParamInfo = new(PayParamInfo)
	)
	id, _ := t.GetInt(":id")
	if id <= 0 {
		err := errors.New("param `:id` empty")
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__SOURCE_DATA__ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	payParam, retcode, err := models.GetOfficialAccountPayParamByOfficialAccountId(id)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if payParam != nil {
		payParamInfo = &PayParamInfo{
			Appkey: payParam.Appkey,
			MchId:  payParam.MchId,
			Name:   payParam.Name,
		}
	}
	t.Data["json"] = map[string]interface{}{
		"err_code":          0,
		"err_msg":           "",
		"wechat_pay_params": *payParamInfo,
	}
	t.ServeJSON()
	return
}
