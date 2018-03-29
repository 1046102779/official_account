package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/1046102779/official_account/common/consts"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
)

// OfficialAccountIndustryCodesController oprations for OfficialAccountIndustryCodes
type OfficialAccountIndustryCodesController struct {
	beego.Controller
}

// 获取设置的行业信息
// @router /:id/industry_codes [GET]
func (t *OfficialAccountIndustryCodesController) GetOfficialAccountIndustry() {
	type IndustryInfo struct {
		FirstIndustry string `json:"first_industry"`
		SecIndustry   string `json:"sec_industry"`
	}
	var (
		industryInfo                 *IndustryInfo                         = new(IndustryInfo)
		officialAccountIndustryCodes []models.OfficialAccountIndustryCodes = []models.OfficialAccountIndustryCodes{}
		retcode                      int
	)
	now := time.Now()
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
	o := orm.NewOrm()
	num, err := o.QueryTable((&models.OfficialAccountIndustryCodes{}).TableName()).Filter("official_account_id", id).Filter("status", consts.STATUS_VALID).All(&officialAccountIndustryCodes)
	if err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__DB__READ,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if num > 0 {
		if officialAccountIndustryCodes[0].IndustryId1 > 0 {
			industryCode := &models.IndustryCodeQuerys{
				Id: officialAccountIndustryCodes[0].IndustryId1,
			}
			if retcode, err = industryCode.GetIndustryCodeQueryNoLock(&o); err != nil {
				Logger.Error(err.Error())
				t.Data["json"] = map[string]interface{}{
					"err_code": retcode,
					"err_msg":  errors.Cause(err).Error(),
				}
				t.ServeJSON()
				return
			}
			industryInfo.FirstIndustry = industryCode.SecIndustryCode
		}
		if officialAccountIndustryCodes[0].IndustryId2 > 0 {
			industryCode := &models.IndustryCodeQuerys{
				Id: officialAccountIndustryCodes[0].IndustryId2,
			}
			if retcode, err = industryCode.GetIndustryCodeQueryNoLock(&o); err != nil {
				Logger.Error(err.Error())
				t.Data["json"] = map[string]interface{}{
					"err_code": retcode,
					"err_msg":  errors.Cause(err).Error(),
				}
				t.ServeJSON()
				return
			}
			industryInfo.SecIndustry = industryCode.SecIndustryCode
		}
	} else {
		// 1.从微信公众号获取设置的行业信息
		token, retcode, err := models.GetAuthorierAccessTokenById(id, &o)
		if err != nil {
			Logger.Error(err.Error())
			t.Data["json"] = map[string]interface{}{
				"err_code": retcode,
				"err_msg":  errors.Cause(err).Error(),
			}
			t.ServeJSON()
			return
		}
		weMessageTemplate := new(models.WeMessageTemplate)
		if industryInfoWechat, retcode, err := weMessageTemplate.GetOfficialAccountIndustry(token); err != nil {
			Logger.Error(err.Error())
			t.Data["json"] = map[string]interface{}{
				"err_code": retcode,
				"err_msg":  errors.Cause(err).Error(),
			}
			t.ServeJSON()
			return
		} else {
			industryId1, retcode, err := models.GetIndustryIdByName(industryInfoWechat.PrimaryIndustry.FirstClass, industryInfoWechat.PrimaryIndustry.SecondClass)
			if err != nil {
				Logger.Error(err.Error())
				t.Data["json"] = map[string]interface{}{
					"err_code": retcode,
					"err_msg":  errors.Cause(err).Error(),
				}
				t.ServeJSON()
				return
			}
			industryId2, retcode, err := models.GetIndustryIdByName(industryInfoWechat.SecIndustry.FirstClass, industryInfoWechat.SecIndustry.SecondClass)
			if err != nil {
				Logger.Error(err.Error())
				t.Data["json"] = map[string]interface{}{
					"err_code": retcode,
					"err_msg":  errors.Cause(err).Error(),
				}
				t.ServeJSON()
				return
			}
			// 2.添加行业信息记录
			officialAccountIndustryCode := &models.OfficialAccountIndustryCodes{
				OfficialAccountId: id,
				IndustryId1:       industryId1,
				IndustryId2:       industryId2,
				Status:            consts.STATUS_VALID,
				UpdatedAt:         now,
				CreatedAt:         now,
			}
			if retcode, err := officialAccountIndustryCode.InsertOfficialAccountIndustryCodesNoLock(&o); err != nil {
				Logger.Error(err.Error())
				t.Data["json"] = map[string]interface{}{
					"err_code": retcode,
					"err_msg":  errors.Cause(err).Error(),
				}
				t.ServeJSON()
				return
			}
			industryInfo.FirstIndustry = industryInfoWechat.PrimaryIndustry.SecondClass
			industryInfo.SecIndustry = industryInfoWechat.SecIndustry.SecondClass
		}
	}

	t.Data["json"] = map[string]interface{}{
		"err_code":      0,
		"err_msg":       "",
		"industry_info": *industryInfo,
	}
	t.ServeJSON()
	return
}

// 更新公众号所属行业
// @router /:id/industry_codes [PUT]
func (t *OfficialAccountIndustryCodesController) UpdateOfficialAccountIndustryCode() {
	type IndustryInfo struct {
		IndustryId1 int `json:"industry_id1"`
		IndustryId2 int `json:"industry_id2"`
	}
	var (
		industryInfo *IndustryInfo = new(IndustryInfo)
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
	if err := json.Unmarshal(t.Ctx.Input.RequestBody, industryInfo); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__JSON__PARSE_FAILED,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	if industryInfo.IndustryId1 <= 0 || industryInfo.IndustryId2 <= 0 {
		err := errors.New("param `industry_id1 | industry_id2` empty")
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": consts.ERROR_CODE__SOURCE_DATA__ILLEGAL,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	offAccInCode := &models.OfficialAccountIndustryCodes{
		OfficialAccountId: id,
	}
	if _, _, retcode, err := offAccInCode.GetOfficialAccountIndustryNoLock(); err != nil {
		Logger.Error(err.Error())
		t.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  errors.Cause(err).Error(),
		}
		t.ServeJSON()
		return
	}
	offAccInCode.IndustryId1 = industryInfo.IndustryId1
	offAccInCode.IndustryId2 = industryInfo.IndustryId2
	fmt.Println("hello,world")
	o := orm.NewOrm()
	if retcode, err := offAccInCode.UpdateOfficialAccountIndustryCodeNoLock(&o); err != nil {
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
