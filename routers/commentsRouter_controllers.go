package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:AccountMessageSendRecordsController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:AccountMessageSendRecordsController"],
		beego.ControllerComments{
			Method: "SendAccountMessage",
			Router: `/:id/message`,
			AllowHTTPMethods: []string{"POST"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:AccountMessageTemplatesController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:AccountMessageTemplatesController"],
		beego.ControllerComments{
			Method: "LogicDeleteAccountMessageTemplate",
			Router: `/accounts/:id/message_templates/:message_id/invalid`,
			AllowHTTPMethods: []string{"PUT"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:AccountMessageTemplatesController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:AccountMessageTemplatesController"],
		beego.ControllerComments{
			Method: "AddAccountMessageTemplate",
			Router: `/:id/message_template`,
			AllowHTTPMethods: []string{"POST"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:AccountMessageTemplatesController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:AccountMessageTemplatesController"],
		beego.ControllerComments{
			Method: "GetAccountMessageTemplates",
			Router: `/:id/message_templates`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:Authorized"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:Authorized"],
		beego.ControllerComments{
			Method: "ComponentVerifyTicket",
			Router: `/`,
			AllowHTTPMethods: []string{"POST"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:Authorized"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:Authorized"],
		beego.ControllerComments{
			Method: "Authorization",
			Router: `/authorization`,
			AllowHTTPMethods: []string{"POST"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:Authorized"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:Authorized"],
		beego.ControllerComments{
			Method: "GetComponentLoginPage",
			Router: `/authorization/loginpage`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:Authorized"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:Authorized"],
		beego.ControllerComments{
			Method: "GetAuthorizedCode",
			Router: `/authorization/code`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:IndustryCodeQuerysController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:IndustryCodeQuerysController"],
		beego.ControllerComments{
			Method: "GetAllMainIndutry",
			Router: `/main_industrys`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:IndustryCodeQuerysController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:IndustryCodeQuerysController"],
		beego.ControllerComments{
			Method: "GetSecIndutryListNoLocks",
			Router: `/:id/deputy_industrys`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:MessageController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:MessageController"],
		beego.ControllerComments{
			Method: "WechatCallback",
			Router: `/:appid/callback`,
			AllowHTTPMethods: []string{"POST"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:MessageTemplatesController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:MessageTemplatesController"],
		beego.ControllerComments{
			Method: "GetMessageTemplates",
			Router: `/system/message_templates`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:MessageTemplatesController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:MessageTemplatesController"],
		beego.ControllerComments{
			Method: "LogicDeleteSystemMessageTemplate",
			Router: `/system/message_templates/:id/invalid`,
			AllowHTTPMethods: []string{"PUT"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountIndustryCodesController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountIndustryCodesController"],
		beego.ControllerComments{
			Method: "GetOfficialAccountIndustry",
			Router: `/:id/industry_codes`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountIndustryCodesController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountIndustryCodesController"],
		beego.ControllerComments{
			Method: "UpdateOfficialAccountIndustryCode",
			Router: `/:id/industry_codes`,
			AllowHTTPMethods: []string{"PUT"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountUsersController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountUsersController"],
		beego.ControllerComments{
			Method: "OfficialAccountAuthorizationUser",
			Router: `/:id/user/authorization`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountUsersController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountUsersController"],
		beego.ControllerComments{
			Method: "GetUserAccessToken",
			Router: `/user/authorization/callback`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountsController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountsController"],
		beego.ControllerComments{
			Method: "GetOfficialAccountBaseInfo",
			Router: `/baseinfo/:appid`,
			AllowHTTPMethods: []string{"GET"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountsPayParamsController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountsPayParamsController"],
		beego.ControllerComments{
			Method: "UploadCertification",
			Router: `/:id/certification`,
			AllowHTTPMethods: []string{"POST"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountsPayParamsController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:OfficialAccountsPayParamsController"],
		beego.ControllerComments{
			Method: "ModifyWechatParams",
			Router: `/:id/payparams`,
			AllowHTTPMethods: []string{"POST"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:WechatPayController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:WechatPayController"],
		beego.ControllerComments{
			Method: "GetPayJsapiParams",
			Router: `/:id/pay/jsapi_params/:bill_id/open_id/:open_id`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:WechatPayController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:WechatPayController"],
		beego.ControllerComments{
			Method: "GetPayNativeParams",
			Router: `/:id/pay/native_params/:bill_id`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:WechatPayController"] = append(beego.GlobalControllerRouter["github.com/1046102779/official_account/controllers:WechatPayController"],
		beego.ControllerComments{
			Method: "NotifyUrl",
			Router: `/notification`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

}
