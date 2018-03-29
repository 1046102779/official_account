package models

import (
	"strconv"
	"time"

	"github.com/1046102779/common/consts"
	"github.com/1046102779/common/types"
	. "github.com/1046102779/official_account/logger"
	"github.com/astaxie/beego/orm"
	"gopkg.in/chanxuehong/wechat.v2/mch/core"
)

type OfficialAccountServer struct{}

// 1. 微信JSAPI短信充值支付
func (t *OfficialAccountServer) GetSmsRechargePayJsapiParams(sri *types.SmsRechargeInfo) (
	wjpi *types.WechatJSAPIParamInfo, err error) {
	Logger.Info("[%v] enter GetSmsRechargePayJsapiParams.", sri.TradeNo)
	defer Logger.Info("[%v] left GetSmsRechargePayJsapiParams.", sri.TradeNo)
	var (
		prepayId            string
		payParamInfo        *WechatPayParamInfo
		unifiedOrderRespMap map[string]string
	)
	defer func() {
		err = nil
	}()
	bill := &BillInfo{
		Money:        sri.Money,
		Title:        sri.Title,
		TradeNoJsapi: sri.TradeNo,
	}
	payParamInfo, unifiedOrderRespMap, _, err = UnifiedOrder(
		int(sri.OfficialAccountId),
		bill,
		sri.Openid,
		consts.TYPE_PAY__WECHAT__JSAPI,
		consts.TYPE_PAY_ENV__WECHAT__SMS_RECHARGE,
	)
	if err != nil {
		Logger.Error("Get unified order error:%v", err.Error())
		return
	}
	if unifiedOrderRespMap["return_code"] == "SUCCESS" && unifiedOrderRespMap["result_code"] == "SUCCESS" {
		prepayId = unifiedOrderRespMap["prepay_id"]
	} else {
		Logger.Error("unifiedOrderRespMap:[%v]", unifiedOrderRespMap)
		return
	}

	wcPayParams := make(map[string]string)
	wcPayParams["appId"] = payParamInfo.Appid
	wcPayParams["timeStamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	wcPayParams["nonceStr"] = "nonce_str"
	wcPayParams["package"] = "prepay_id=" + prepayId
	wcPayParams["signType"] = "MD5"
	wcPayParams["paySign"] = core.Sign(wcPayParams, payParamInfo.Appkey, nil)
	// 输出结果:
	wjpi = &types.WechatJSAPIParamInfo{
		AppId:     wcPayParams["appId"],
		TimeStamp: wcPayParams["timeStamp"],
		NonceStr:  wcPayParams["nonceStr"],
		Package:   wcPayParams["package"],
		SignType:  wcPayParams["signType"],
		PaySign:   wcPayParams["paySign"],
	}
	return
}

// 3. 获取商品订单支付二维码
func (t *OfficialAccountServer) GetSaleOrderPayNativeParams(ppi *types.ProductPayInfo) (codeUrl string, err error) {
	Logger.Info("[%v] enter GetSaleOrderPayNativeParams.", ppi.TradeNo)
	defer Logger.Info("[%v] left GetSaleOrderPayNativeParams.", ppi.TradeNo)
	var (
		unifiedOrderRespMap map[string]string
	)
	defer func() {
		err = nil
	}()
	bill := &BillInfo{
		Money:         ppi.Money,
		Title:         ppi.Title,
		TradeNoNative: ppi.TradeNo,
	}
	//get unified order
	_, unifiedOrderRespMap, _, err = UnifiedOrder(
		int(ppi.OfficialAccountId),
		bill,
		"",
		consts.TYPE_PAY__WECHAT__NATIVE,
		consts.TYPE_PAY_ENV__WECHAT__FABRIC_ORDER)
	if err != nil {
		Logger.Error("Get unified order error:%v", err.Error())
		return
	} else {
		if unifiedOrderRespMap["return_code"] == "SUCCESS" && unifiedOrderRespMap["result_code"] == "SUCCESS" {
			codeUrl = unifiedOrderRespMap["code_url"]
		} else {
			Logger.Error("unifiedOrderRespMap:[%v]", unifiedOrderRespMap)
			return
		}
	}
	return
}

// 2. 微信Native二维码短信充值支付
func (t *OfficialAccountServer) GetSmsRechargePayNativeParams(sri *types.SmsRechargeInfo) (codeUrl string, err error) {
	Logger.Info("[%v] enter GetSmsRechargePayNativeParams.", sri.TradeNo)
	defer Logger.Info("[%v] left GetSmsRechargePayNativeParams.", sri.TradeNo)
	var (
		unifiedOrderRespMap map[string]string
	)
	defer func() { err = nil }()
	bill := &BillInfo{
		Money:         sri.Money,
		Title:         sri.Title,
		TradeNoNative: sri.TradeNo,
	}
	//get unified order
	if _, unifiedOrderRespMap, _, err = UnifiedOrder(
		int(sri.OfficialAccountId),
		bill,
		"",
		consts.TYPE_PAY__WECHAT__NATIVE,
		consts.TYPE_PAY_ENV__WECHAT__SMS_RECHARGE,
	); err != nil {
		Logger.Error("Get unified order error:%v", err.Error())
		return
	} else {
		if unifiedOrderRespMap["return_code"] == "SUCCESS" && unifiedOrderRespMap["result_code"] == "SUCCESS" {
			codeUrl = unifiedOrderRespMap["code_url"]
		} else {
			Logger.Error("unifiedOrderRespMap:[%v]", unifiedOrderRespMap)
		}
	}
	return
}

// 4. 客户账号与微信绑定，目前只与盈创丰茂公众号绑定
func (t *OfficialAccountServer) BindingCustomerAndWxInfo(openid string, userId int) (err error) {
	Logger.Info("[%v.%v] enter BindingCustomerAndWxInfo.", userId, openid)
	defer Logger.Info("[%v.%v] left BindingCustomerAndWxInfo.", userId, openid)
	var (
		id int // userWxInfoId
	)
	defer func() { err = nil }()
	// 通过openid，获取user_wx_info_id
	// 1 : 表示盈创丰茂ID
	if id, _, err = GetUserWxInfoIdByOpenid(1, openid); err != nil {
		Logger.Error(err.Error())
		return
	}
	// 读取和更新UserWxInfo表记录
	o := orm.NewOrm()
	userWxInfo := &UserWxInfos{
		Id: id,
	}
	if _, err = userWxInfo.ReadUserWxInfoNoLock(&o); err != nil {
		Logger.Error(err.Error())
		return
	}
	userWxInfo.CustomerId = userId
	if _, err = userWxInfo.UpdateUserWxInfoNoLock(&o); err != nil {
		Logger.Error(err.Error())
		return
	}
	return
}

//  5. B端账号与微信绑定，目前只与盈创丰茂公众号绑定
func (t *OfficialAccountServer) BindingUserAndWxInfo(openid string, userId int) (err error) {
	Logger.Info("[%v.%v] enter BindingUserAndWxInfo.", userId, openid)
	defer Logger.Info("[%v.%v] left BindingUserAndWxInfo.", userId, openid)
	var (
		id int // userWxInfoId
	)
	defer func() { err = nil }()
	// 通过openid，获取user_wx_info_id
	// 1 : 表示盈创丰茂ID
	if id, _, err = GetUserWxInfoIdByOpenid(1, openid); err != nil {
		Logger.Error(err.Error())
		return
	}
	// 读取和更新UserWxInfo表记录
	o := orm.NewOrm()
	userWxInfo := &UserWxInfos{
		Id: id,
	}
	if _, err = userWxInfo.ReadUserWxInfoNoLock(&o); err != nil {
		Logger.Error(err.Error())
		return
	}
	userWxInfo.UserId = userId
	if _, err = userWxInfo.UpdateUserWxInfoNoLock(&o); err != nil {
		Logger.Error(err.Error())
		return
	}
	return
}

// 6. 通过公司ID，获取内部公众号ID
func (t *OfficialAccountServer) GetOfficialAccountId(companyId int) (officialAccountId int, err error) {
	Logger.Info("[%v] enter GetOfficialAccountId.", companyId)
	defer Logger.Info("[%v] left GetOfficialAccountId.", companyId)
	var (
		officialAccount *OfficialAccounts
	)
	defer func() { err = nil }()
	officialAccount, _, err = GetOfficialAccountByCompanyId(companyId)
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	if officialAccount != nil {
		officialAccountId = officialAccount.Id
	}
	return
}

func (t *OfficialAccountServer) SendMessage(soid int, soiid int, typ int16) (err error) {
	Logger.Info("[%v.%v.%v] enter SendMessage.", soid, soiid, typ)
	defer Logger.Info("[%v.%v.%v] left SendMessage.", soid, soiid, typ)
	defer func() {
		err = nil
	}()
	if (soid <= 0 && soiid <= 0 &&
		typ != consts.TYPE__MESSAGE_TEMPLATE__STATIS__GENERAL_MANAGER) || typ <= 0 {
		Logger.Error("param `saleOrderId | typ` illegal")
		return
	}
	/*
		switch typ {
		case consts.TYPE__MESSAGE_TEMPLATE__WAITING_TO_OUTS__WAREHOUSE_KEEPER:
			// 仓管员-接收待出库提醒消息
			// 订单ID 分别获取子订单ID列表，待选取库
			err = sendWaitingToOutboundsMessage(soiid)
		case consts.TYPE__MESSAGE_TEMPLATE__OUTS__SALER_OR_SHOP_MANAGER:
			// 销售员或店长-接收已出库提醒(货物出库提醒)
			// 子订单ID 出库
			err = sendOutboundsMessage(soiid)
		case consts.TYPE__MESSAGE_TEMPLATE__LOGISTICS__SALER_OR_SHOP_MANAGER:
			// 销售员或店长-接收已物流提醒()
			// 子订单ID，已发送物流
			err = sendLogisticsMessage(soiid)
		case consts.TYPE__MESSAGE_TEMPLATE__PLACE_ORDER__CUSTOMER:
			// 客户-下单成功时提醒(门店订购提醒通知)
			// 订单ID， 下单成功
			err = sendPlaceOrderSuccessMessage(soid)
		case consts.TYPE__MESSAGE_TEMPLATE__PRICE__CUSTOMER:
			// 客户-总价格确定时接收提醒(订单状态更新通知)
			// 订单ID， 总价格确定
			err = sendDetermineTotalPrice(soid)
		case consts.TYPE__MESSAGE_TEMPLATE__ACCEPTION__CUSTOMER:
			// 客户-待客户验收状态时接收提醒(商铺自提通知(没有取货编码))
			// 订单ID，验收
			err = sendAcceptionMessage(soid)
		case consts.TYPE__MESSAGE_TEMPLATE__STATIS__GENERAL_MANAGER:
			// 总经理-每晚8点接收当天营业额统计情况的提醒(营业情况明细通知)
			// 公司ID
			err = sendStatisticsMessage()
		default:
			err = errors.New("template message `typ` is illegal")
		}
	*/
	return
}
