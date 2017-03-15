/* 服务列表：
*	> 1. 微信JSAPI短信充值支付
*	> 2. 微信Native二维码短信充值支付
*	> 3. 获取公众号下的用户openid
*   > 4. C端客户账号与微信绑定，目前只与盈创丰茂公众号绑定
*   > 5. B端账号与微信绑定，目前只与盈创丰茂公众号绑定
*   > 6. 通过公司ID，获取内部公众号ID
*   > 7. 获取面料订单支付二维码
**/
package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/chanxuehong/wechat/mch"

	"github.com/1046102779/official_account/common/consts"
	pb "github.com/1046102779/official_account/igrpc"
	. "github.com/1046102779/official_account/logger"
)

type OfficialAccountServer struct{}

// 1. 微信JSAPI短信充值支付
func (t *OfficialAccountServer) GetSmsRechargePayJsapiParams(in *pb.SmsRechargeInfo, out *pb.WechatJSAPIParamInfo) (err error) {
	Logger.Info("[%v] enter GetSmsRechargePayJsapiParams.", in.TradeNo)
	defer Logger.Info("[%v] left GetSmsRechargePayJsapiParams.", in.TradeNo)
	var (
		prepayId            string
		payParamInfo        *WechatPayParamInfo
		unifiedOrderRespMap map[string]string
	)
	defer func() {
		err = nil
	}()
	bill := &BillInfo{
		Money:        in.Money,
		Title:        in.Title,
		TradeNoJsapi: in.TradeNo,
	}
	fmt.Println("Openid=", in.Openid)
	payParamInfo, unifiedOrderRespMap, _, err = UnifiedOrder(int(in.OfficialAccountId), bill, in.Openid, consts.TYPE_PAY__WECHAT__JSAPI, consts.TYPE_PAY_ENV__WECHAT__SMS_RECHARGE)
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
	wcPayParams["paySign"] = mch.Sign(wcPayParams, payParamInfo.Appkey, nil)
	// 输出结果:
	out.AppId = wcPayParams["appId"]
	out.TimeStamp = wcPayParams["timeStamp"]
	out.NonceStr = wcPayParams["nonceStr"]
	out.Package = wcPayParams["package"]
	out.SignType = wcPayParams["signType"]
	out.PaySign = wcPayParams["paySign"]
	return
}

// 3. 获取面料订单支付二维码
func (t *OfficialAccountServer) GetSaleOrderPayNativeParams(in *pb.FabricPatternPayInfo, out *pb.WechatNativeParamInfo) (err error) {
	Logger.Info("[%v] enter GetSaleOrderPayNativeParams.", in.TradeNo)
	defer Logger.Info("[%v] left GetSaleOrderPayNativeParams.", in.TradeNo)
	var (
		unifiedOrderRespMap map[string]string
	)
	defer func() {
		err = nil
	}()
	bill := &BillInfo{
		Money:         in.Money,
		Title:         in.Title,
		TradeNoNative: in.TradeNo,
	}
	//get unified order
	_, unifiedOrderRespMap, _, err = UnifiedOrder(int(in.OfficialAccountId), bill, "", consts.TYPE_PAY__WECHAT__NATIVE, consts.TYPE_PAY_ENV__WECHAT__FABRIC_ORDER)
	if err != nil {
		Logger.Error("Get unified order error:%v", err.Error())
		return
	} else {
		if unifiedOrderRespMap["return_code"] == "SUCCESS" && unifiedOrderRespMap["result_code"] == "SUCCESS" {
			out.CodeUrl = unifiedOrderRespMap["code_url"]
		} else {
			Logger.Error("unifiedOrderRespMap:[%v]", unifiedOrderRespMap)
			return
		}
	}
	return
}

// 2. 微信Native二维码短信充值支付
func (t *OfficialAccountServer) GetSmsRechargePayNativeParams(in *pb.SmsRechargeInfo, out *pb.WechatNativeParamInfo) (err error) {
	Logger.Info("[%v] enter GetSmsRechargePayNativeParams.", in.TradeNo)
	defer Logger.Info("[%v] left GetSmsRechargePayNativeParams.", in.TradeNo)
	var (
		unifiedOrderRespMap map[string]string
	)
	defer func() {
		err = nil
	}()
	bill := &BillInfo{
		Money:         in.Money,
		Title:         in.Title,
		TradeNoNative: in.TradeNo,
	}
	//get unified order
	_, unifiedOrderRespMap, _, err = UnifiedOrder(int(in.OfficialAccountId), bill, "", consts.TYPE_PAY__WECHAT__NATIVE, consts.TYPE_PAY_ENV__WECHAT__SMS_RECHARGE)
	if err != nil {
		Logger.Error("Get unified order error:%v", err.Error())
		return
	} else {
		if unifiedOrderRespMap["return_code"] == "SUCCESS" && unifiedOrderRespMap["result_code"] == "SUCCESS" {
			out.CodeUrl = unifiedOrderRespMap["code_url"]
		} else {
			Logger.Error("unifiedOrderRespMap:[%v]", unifiedOrderRespMap)
			return
		}
	}
	return
}

func (t *OfficialAccountServer) GetOpenid(in *pb.UserOpenidInfo, out *pb.UserOpenidInfo) (err error) {
	Logger.Info("[%v.%v] enter GetOpenid.", in.CompanyId, in.UserId)
	defer Logger.Info("[%v.%v] left GetOpenid.", in.CompanyId, in.UserId)
	var (
		info *UserWxInfos
	)
	if in.CompanyId == 1 {
		// 盈创丰茂公众号，用于短信充值
		// 用户微信ID
		info, _, err = GetUserWxInfoByUserId(int(in.UserId))
		if info != nil {
			// 默认盈创丰茂公众号ID=1
			out.Openid, _, err = GetopenidByCond(1, info.Id)
			return
		}
	} else {
		// 其他业务场景
		// ::TODO
	}
	return
}

// 4. 客户账号与微信绑定，目前只与盈创丰茂公众号绑定
func (t *OfficialAccountServer) BindingCustomerAndWxInfo(in *pb.UserOpenidInfo, out *pb.UserOpenidInfo) (err error) {
	Logger.Info("[%v.%v] enter BindingCustomerAndWxInfo.", in.UserId, in.Openid)
	defer Logger.Info("[%v.%v] left BindingCustomerAndWxInfo.", in.UserId, in.Openid)
	var (
		id int // userWxInfoId
	)
	defer func() {
		err = nil
	}()
	// 通过openid，获取user_wx_info_id
	// 1 : 表示盈创丰茂ID
	id, _, err = GetUserWxInfoIdByOpenid(1, in.Openid)
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	// 读取和更新UserWxInfo表记录
	o := orm.NewOrm()
	//now := time.Now()
	userWxInfo := &UserWxInfos{
		Id: id,
	}
	if _, err = userWxInfo.ReadUserWxInfoNoLock(&o); err != nil {
		Logger.Error(err.Error())
		return
	}
	//userWxInfo.UpdatedAt = now
	userWxInfo.CustomerId = int(in.UserId)
	if _, err = userWxInfo.UpdateUserWxInfoNoLock(&o); err != nil {
		Logger.Error(err.Error())
		return
	}
	return
}

//  5. B端账号与微信绑定，目前只与盈创丰茂公众号绑定
func (t *OfficialAccountServer) BindingUserAndWxInfo(in *pb.UserOpenidInfo, out *pb.UserOpenidInfo) (err error) {
	Logger.Info("[%v.%v] enter BindingUserAndWxInfo.", in.UserId, in.Openid)
	defer Logger.Info("[%v.%v] left BindingUserAndWxInfo.", in.UserId, in.Openid)
	var (
		id int // userWxInfoId
	)
	defer func() {
		err = nil
	}()
	// 通过openid，获取user_wx_info_id
	// 1 : 表示盈创丰茂ID
	id, _, err = GetUserWxInfoIdByOpenid(1, in.Openid)
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	// 读取和更新UserWxInfo表记录
	o := orm.NewOrm()
	//now := time.Now()
	userWxInfo := &UserWxInfos{
		Id: id,
	}
	if _, err = userWxInfo.ReadUserWxInfoNoLock(&o); err != nil {
		Logger.Error(err.Error())
		return
	}
	userWxInfo.UserId = int(in.UserId)
	if _, err = userWxInfo.UpdateUserWxInfoNoLock(&o); err != nil {
		Logger.Error(err.Error())
		return
	}
	return
}

// 6. 通过公司ID，获取内部公众号ID
func (t *OfficialAccountServer) GetOfficialAccountId(in *pb.CompanyAndOfficialAccountRel, out *pb.CompanyAndOfficialAccountRel) (err error) {
	Logger.Info("[%v] enter GetOfficialAccountId.", in.CompanyId)
	defer Logger.Info("[%v] left GetOfficialAccountId.", in.CompanyId)
	var (
		officialAccount *OfficialAccounts
	)
	defer func() {
		err = nil
	}()
	officialAccount, _, err = GetOfficialAccountByCompanyId(int(in.CompanyId))
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	if officialAccount != nil {
		out.CompanyId = int64(officialAccount.CompanyId)
		out.OfficialAccountId = int64(officialAccount.Id)
	}
	return
}

func (t *OfficialAccountServer) SendMessage(in *pb.SaleOrderTemplateMessage, out *pb.SaleOrderTemplateMessage) (err error) {
	Logger.Info("[%v.%v.%v] enter SendMessage.", in.SaleOrderId, in.SaleOrderItemId, in.Type)
	defer Logger.Info("[%v.%v.%v] left SendMessage.", in.SaleOrderId, in.SaleOrderItemId, in.Type)
	defer func() {
		if err != nil {
			Logger.Error(err.Error())
		}
		err = nil
	}()
	if (in.SaleOrderId <= 0 && in.SaleOrderItemId <= 0 &&
		in.Type != consts.TYPE__MESSAGE_TEMPLATE__STATIS__GENERAL_MANAGER) || in.Type <= 0 {
		Logger.Error("param `saleOrderId | typ` illegal")
		return
	}
	switch int16(in.Type) {
	case consts.TYPE__MESSAGE_TEMPLATE__WAITING_TO_OUTS__WAREHOUSE_KEEPER:
		// 仓管员-接收待出库提醒消息
		// 订单ID 分别获取子订单ID列表，待选取库
		err = sendWaitingToOutboundsMessage(int(in.SaleOrderItemId))
	case consts.TYPE__MESSAGE_TEMPLATE__OUTS__SALER_OR_SHOP_MANAGER:
		// 销售员或店长-接收已出库提醒(货物出库提醒)
		// 子订单ID 出库
		err = sendOutboundsMessage(int(in.SaleOrderItemId))
	case consts.TYPE__MESSAGE_TEMPLATE__LOGISTICS__SALER_OR_SHOP_MANAGER:
		// 销售员或店长-接收已物流提醒()
		// 子订单ID，已发送物流
		err = sendLogisticsMessage(int(in.SaleOrderItemId))
	case consts.TYPE__MESSAGE_TEMPLATE__PLACE_ORDER__CUSTOMER:
		// 客户-下单成功时提醒(门店订购提醒通知)
		// 订单ID， 下单成功
		err = sendPlaceOrderSuccessMessage(int(in.SaleOrderId))
	case consts.TYPE__MESSAGE_TEMPLATE__PRICE__CUSTOMER:
		// 客户-总价格确定时接收提醒(订单状态更新通知)
		// 订单ID， 总价格确定
		err = sendDetermineTotalPrice(int(in.SaleOrderId))
	case consts.TYPE__MESSAGE_TEMPLATE__ACCEPTION__CUSTOMER:
		// 客户-待客户验收状态时接收提醒(商铺自提通知(没有取货编码))
		// 订单ID，验收
		err = sendAcceptionMessage(int(in.SaleOrderId))
	case consts.TYPE__MESSAGE_TEMPLATE__STATIS__GENERAL_MANAGER:
		// 总经理-每晚8点接收当天营业额统计情况的提醒(营业情况明细通知)
		// 公司ID
		err = sendStatisticsMessage()
	default:
		err = errors.New("template message `typ` is illegal")
	}
	return
}
