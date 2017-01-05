package models

import (
	"fmt"
	"strconv"
	"time"

	utils "github.com/1046102779/common"
	"github.com/chanxuehong/wechat/mch"

	pb "github.com/1046102779/igrpc"
	. "github.com/1046102779/official_account/logger"
)

/* 服务列表：
*	> 1. 微信JSAPI短信充值支付
*	> 2. 微信Native二维码短信充值支付
*	> 3. 获取公众号下的用户openid
**/
type OfficialAccountServer struct{}

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
	payParamInfo, unifiedOrderRespMap, _, err = UnifiedOrder(int(in.OfficialAccountId), bill, in.Openid, utils.TRADE_TYPE_JSAPI, utils.WECHAT_PAY_BUSINESS_PLATFORM_SMS_RECHARGE)
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

func (t *OfficialAccountServer) GetSmsRechargePayNativeParams(in *pb.SmsRechargeInfo, out *pb.WechatNativeParamInfo) (err error) {
	Logger.Info("[%v] enter GetSmsRechargePayNativeParams.", in.TradeNo)
	defer Logger.Info("[%v] left GetSmsRechargePayNativeParams.", in.TradeNo)
	bill := &BillInfo{
		Money:         in.Money,
		Title:         in.Title,
		TradeNoNative: in.TradeNo,
	}
	//get unified order
	_, unifiedOrderRespMap, _, err := UnifiedOrder(int(in.OfficialAccountId), bill, "", utils.TRADE_TYPE_NATIVE, utils.WECHAT_PAY_BUSINESS_PLATFORM_SMS_RECHARGE)
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
