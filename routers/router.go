// @APIVersion 1.0.0
// @Title beego official_account API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"git.kissdata.com/ycfm/official_account/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/wechats",
			beego.NSInclude(
				&controllers.Authorized{},
				&controllers.OfficialAccountsController{},
				&controllers.MessageController{},
				&controllers.OfficialAccountsPayParamsController{},
				&controllers.OfficialAccountUsersController{},
				&controllers.IndustryCodeQuerysController{},
				&controllers.OfficialAccountIndustryCodesController{},
				&controllers.MessageTemplatesController{},
				&controllers.AccountMessageTemplatesController{},
				&controllers.AccountMessageSendRecordsController{},
				&controllers.WechatPayController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
