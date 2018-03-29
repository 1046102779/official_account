package models

import (
	"fmt"

	"github.com/1046102779/common/consts"
	. "github.com/1046102779/common/utils"
	"github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
	"gopkg.in/chanxuehong/wechat.v2/mch/core"
	"gopkg.in/chanxuehong/wechat.v2/mch/pay"
)

type WechatPayParamInfo struct {
	Appid    string // 公众号appid
	MchId    string // 商户号ID
	Appkey   string // 支付密钥
	CertFile string // 证书1
	KeyFile  string // 证书2
}

type BillInfo struct {
	Money              int64  `json:"money"`
	ActivitySubOrderId int64  `json:"activity_sub_order_id"`
	TradeNo            string `json:"trade_no"`
	TradeNoJsapi       string `json:"trade_no_jsapi"`
	TradeNoNative      string `json:"trade_no_native"`
	PrepayId           string `json:"prepay_id"`
	PayStatus          int16  `json:"pay_status"`
	Title              string `json:"title"`
}

func GetWechatPayParams(id int) (payParamInfo *WechatPayParamInfo, retcode int, err error) {
	Logger.Info("[%v] enter GetWechatPayParams.", id)
	defer Logger.Info("[%v] left GetWechatPayParams.", id)
	var (
		officialAccountsPayParams []OfficialAccountsPayParams = []OfficialAccountsPayParams{}
		num                       int64
	)
	if id <= 0 {
		err = errors.New("pay param `:id` empty")
		Logger.Error(err.Error())
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	payParamInfo = new(WechatPayParamInfo)
	o := orm.NewOrm()
	// 获取appid
	officialAccount := &OfficialAccounts{
		Id: id,
	}
	if retcode, err = officialAccount.ReadOfficialAccountNoLock(&o); err != nil {
		err = errors.Wrap(err, "GetWechatPayParams")
		return
	}
	payParamInfo.Appid = officialAccount.Appid
	// 获取公众号支付参数
	num, err = o.QueryTable((&OfficialAccountsPayParams{}).TableName()).Filter("official_account_id", id).Filter("status", consts.STATUS_VALID).All(&officialAccountsPayParams)
	if err != nil {
		err = errors.Wrap(err, "getWechatPayParams")
		retcode = consts.ERROR_CODE__DB__READ
		return
	}
	if num > 0 {
		payParamInfo.Appkey = officialAccountsPayParams[0].Appkey
		payParamInfo.MchId = officialAccountsPayParams[0].MchId
		/*::TODO 建议用微信支付安全证书
		payParamInfo.CertFile = fmt.Sprintf("%s/%s/%s", conf.CertificationDir, payParamInfo.Appid, "apiclient_cert.pem")
		payParamInfo.KeyFile = fmt.Sprintf("%s/%s/%s", conf.CertificationDir, payParamInfo.Appid, "apiclient_key.pem")
		*/
	} else {
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		err = errors.New("pay params `appkey | mch_id | appid | certification_files` not exist")
		return
	}
	return
}

// @params id: 内部公众号ID
// @params attach: 附加数据，在查询API和支付通知中原样返回，可作为自定义参数使用。
// 附加数据，目前主要用于短信交易类型：短信充值、购买布匹交易
func UnifiedOrder(id int, bill *BillInfo, openId string, tradeType string, attach string) (payParamInfo *WechatPayParamInfo, unifiedOrderRespMap map[string]string, retcode int, err error) {
	Logger.Info("[%v] enter unifiedOrder.", id)
	defer Logger.Info("[%v] left unifiedOrder.", id)
	if payParamInfo, retcode, err = GetWechatPayParams(id); err != nil {
		err = errors.Wrap(err, "unifiedOrder")
		return
	}
	//get unified order
	unifiedOrderReqMap := make(map[string]string)
	if len(bill.Title) > 110 {
		unifiedOrderReqMap["body"] = SubString(bill.Title, 0, 30) + "..."
	} else {
		unifiedOrderReqMap["body"] = bill.Title
	}
	unifiedOrderReqMap["appid"] = payParamInfo.Appid
	unifiedOrderReqMap["mch_id"] = payParamInfo.MchId
	unifiedOrderReqMap["device_info"] = ""
	unifiedOrderReqMap["nonce_str"] = "nonce_str"
	unifiedOrderReqMap["attach"] = attach // 附加数据，目前主要用于短信交易类型：短信充值、购买布匹交易
	if tradeType == consts.TYPE_PAY__WECHAT__JSAPI {
		unifiedOrderReqMap["out_trade_no"] = bill.TradeNoJsapi
	} else {
		unifiedOrderReqMap["out_trade_no"] = bill.TradeNoNative
	}
	unifiedOrderReqMap["total_fee"] = fmt.Sprintf("%d", bill.Money)
	unifiedOrderReqMap["spbill_create_ip"] = "127.0.0.1"
	unifiedOrderReqMap["time_start"] = ""
	unifiedOrderReqMap["time_expire"] = ""
	unifiedOrderReqMap["goods_tag"] = ""
	unifiedOrderReqMap["notify_url"] = conf.NotifyUrl
	unifiedOrderReqMap["trade_type"] = tradeType
	unifiedOrderReqMap["openid"] = openId
	unifiedOrderReqMap["product_id"] = ""
	unifiedOrderReqMap["sign"] = core.Sign(unifiedOrderReqMap, payParamInfo.Appkey, nil)

	Logger.Debug("Get unified order:: request:: [%v]", unifiedOrderReqMap)
	/* ::TODO 建议用微信支付安全证书
	client, err := mch.NewTLSHttpClient(payParamInfo.CertFile, payParamInfo.KeyFile)
	if err != nil {
		err = errors.Wrap(err, "unifiedOrder")
		retcode = 32122
		return
	}
	proxy := mch.NewProxy(payParamInfo.Appid, payParamInfo.MchId, payParamInfo.Appkey, client)
	*/
	proxy := core.NewClient(payParamInfo.Appid, payParamInfo.MchId, payParamInfo.Appkey, nil)
	unifiedOrderRespMap = make(map[string]string)
	unifiedOrderRespMap, err = pay.UnifiedOrder(proxy, unifiedOrderReqMap)
	if err != nil {
		retcode = 32122
		err = errors.Wrap(err, "unifiedOrder")
		return
	}
	return
}
