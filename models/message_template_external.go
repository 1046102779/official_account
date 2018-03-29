package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/1046102779/common/consts"
	"github.com/1046102779/common/httpRequest"
	. "github.com/1046102779/official_account/logger"
	"github.com/pkg/errors"
)

// 1.设置所属行业
// 2.获取设置的行业信息
// 3.获得模板ID
// 4.获取模板列表
// 5.删除模板
// 6.发送模板消息
// 7.事件推送
type WeMessageTemplate struct{}

// 1.设置所属行业
func (t *WeMessageTemplate) SetOfficialAccountIndustry(firstIndustry int, secIndustry int, accessToken string) (retcode int, err error) {
	type IndustryData struct {
		FirstIndustry int `json:"industry_id1"`
		SecIndustry   int `json:"industry_id2"`
	}
	Logger.Info("enter SetOfficialAccountIndustry.")
	defer Logger.Info("left SetOfficialAccountIndustry.")
	var (
		retBody      []byte
		industryData *IndustryData = new(IndustryData)
	)
	if "" == strings.TrimSpace(accessToken) || firstIndustry <= 0 || secIndustry <= 0 {
		err = errors.New("param `access_token | industry_id1 | industry_id2` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	industryData.FirstIndustry = firstIndustry
	industryData.SecIndustry = secIndustry
	bodyData, _ := json.Marshal(*industryData)
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/template/api_set_industry?access_token=%s", accessToken)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "SetOfficialAccountIndustry")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	fmt.Println("Set Industry result: ", string(retBody))
	return
}

type ClassIndustry struct {
	FirstClass  string `json:"first_class"`
	SecondClass string `json:"second_class"`
}

type IndustryInfo struct {
	PrimaryIndustry ClassIndustry `json:"primary_industry"`
	SecIndustry     ClassIndustry `json:"secondary_industry"`
}

// 2.获取设置的行业信息
func (t *WeMessageTemplate) GetOfficialAccountIndustry(accessToken string) (industryInfo *IndustryInfo, retcode int, err error) {
	Logger.Info("[%v] enter GetOfficialAccountIndustry.", accessToken)
	defer Logger.Info("[%v] left GetOfficialAccountIndustry.", accessToken)
	var (
		retBody []byte
	)
	industryInfo = new(IndustryInfo)
	if "" == strings.TrimSpace(accessToken) {
		err = errors.New("param `access_token` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/template/get_industry?access_token=%s", accessToken)
	retBody, err = httpRequest.HttpGetBody(httpStr)
	if err != nil {
		err = errors.Wrap(err, "GetOfficialAccountIndustry")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, industryInfo); err != nil {
		err = errors.Wrap(err, "GetOfficialAccountIndustry")
		retcode = consts.ERROR_CODE__JSON__PARSE_FAILED
		return
	}
	return
}

// 3.获得模板ID
// @param: templateCode模板编号
func (t *WeMessageTemplate) GetTemplateId(accessToken string, templateCode string) (templateId string, retcode int, err error) {
	type TemplateInfo struct {
		TemplateCode string `json:"template_id_short"`
	}
	Logger.Info("[%v] enter GetTemplateId.", templateCode)
	defer Logger.Info("[%v] left GetTemplateId.", templateCode)
	var (
		retJson      map[string]interface{} = map[string]interface{}{}
		templateInfo *TemplateInfo          = new(TemplateInfo)
	)
	if "" == strings.TrimSpace(accessToken) || "" == strings.TrimSpace(templateCode) {
		err = errors.New("param `access_token | message_template_code` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	templateInfo.TemplateCode = templateCode
	bodyData, _ := json.Marshal(*templateInfo)
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/template/api_add_template?access_token=%s", accessToken)
	retJson, err = httpRequest.HttpPostJson(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "GetTemplateId")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	templateId = retJson["template_id"].(string)
	return
}

type TemplateInfo struct {
	TemplateId      string `json:"template_id"`
	Title           string `json:"title"`
	PrimaryIndustry string `json:"primary_industry"`
	DeputyIndustry  string `json:"deputy_industry"`
	Content         string `json:"content"`
	Example         string `json:"example"`
}
type TemplateInfos struct {
	TemplateList []TemplateInfo `json:"template_list"`
}

// 4.获取模板列表
func (t *WeMessageTemplate) GetTemplateList(accessToken string) (templateInfos *TemplateInfos, retcode int, err error) {
	Logger.Info("[%v] enter GetTemplateList.", accessToken)
	defer Logger.Info("[%v] left GetTemplateList.", accessToken)
	var (
		retBody []byte
	)
	templateInfos = new(TemplateInfos)
	if "" == strings.TrimSpace(accessToken) {
		err = errors.New("param `access_token` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/template/get_all_private_template?access_token=%s", accessToken)
	retBody, err = httpRequest.HttpGetBody(httpStr)
	if err != nil {
		err = errors.Wrap(err, "GetTemplateList")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, templateInfos); err != nil {
		err = errors.Wrap(err, "GetTemplateList")
		retcode = consts.ERROR_CODE__JSON__PARSE_FAILED
		return
	}
	return
}

// 5.删除模板
func (t *WeMessageTemplate) DeleteMessageTemplate(accessToken string, templateId string) (retcode int, err error) {
	Logger.Info("[%v] enter DeleteMessageTemplate.", templateId)
	defer Logger.Info("[%v] left DeleteMessageTemplate.", templateId)
	type TemplateInfo struct {
		TemplateId string `json:"template_id"`
	}
	var (
		templateInfo *TemplateInfo          = new(TemplateInfo)
		retJson      map[string]interface{} = map[string]interface{}{}
	)
	if "" == strings.TrimSpace(accessToken) || "" == strings.TrimSpace(templateId) {
		err = errors.New("param `access_token | template_id`  empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	templateInfo.TemplateId = templateId
	bodyData, _ := json.Marshal(*templateInfo)
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/template/del_private_template?access_token=%s", accessToken)
	retJson, err = httpRequest.HttpPostJson(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "DeleteMessageTemplate")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	retcode = int(retJson["errcode"].(float64))
	err = errors.New(retJson["errmsg"].(string))
	if err.Error() == "ok" && retcode == 0 {
		err = nil
	}
	return
}

// 6.发送模板消息
// @params: toUser是指接收消息的用户openid
func (t *WeMessageTemplate) SendMessage(accessToken string, templateId string, content json.RawMessage, toUser string) (msgId string, retcode int, err error) {
	Logger.Info("[%v] enter SendMessage.", templateId)
	defer Logger.Info("[%v] left SendMessage.", templateId)
	type MessageInfo struct {
		ToUser     string           `json:"touser"`
		TemplateId string           `json:"template_id"`
		Data       *json.RawMessage `json:"data"`
	}
	var (
		retJson map[string]interface{} = map[string]interface{}{}
	)
	if "" == strings.TrimSpace(accessToken) || "" == strings.TrimSpace(templateId) || "" == strings.TrimSpace(string(content)) {
		err = errors.New("param `access_token | template_id | message_content` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	messageInfo := &MessageInfo{
		ToUser:     toUser,
		TemplateId: templateId,
		Data:       &content,
	}
	bodyData, err := json.Marshal(*messageInfo)
	if err != nil {
		Logger.Error(err.Error())
		retcode = consts.ERROR_CODE__JSON__PARSE_FAILED
		return
	}
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken)
	retJson, err = httpRequest.HttpPostJson(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "SendMessage")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	if retJson["errmsg"].(string) == "ok" {
		msgId = fmt.Sprintf("%d", int(retJson["msgid"].(float64)))
	} else {
		err = errors.New(retJson["errmsg"].(string))
		retcode = int(retJson["errcode"].(float64))
	}
	return
}

// 7.事件推送
