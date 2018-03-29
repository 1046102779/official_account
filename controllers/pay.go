package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1046102779/common/types"
	"github.com/1046102779/official_account/common/consts"
	. "github.com/1046102779/official_account/common/utils"
	. "github.com/1046102779/official_account/logger"
	"github.com/1046102779/official_account/models"
	"gopkg.in/chanxuehong/wechat.v2/mch/core"

	"github.com/astaxie/beego"
	"github.com/chanxuehong/util"
)

const (
	//error code:
	BILL_MONEY_ERROR  = 7806
	BILL_STATUS_ERROR = 7807
)

type WechatPayController struct {
	beego.Controller
}

// @params :id 内部公众号ID
// @params :bill_id 订单ID
// @params :openid 用户在公众号下的openid
// @router /:id/pay/jsapi_params/:bill_id/open_id/:open_id [get]
func (w *WechatPayController) GetPayJsapiParams() {
	var (
		prepayId string
	)

	Logger.Informational("enter GetPayJsapiParams.")
	defer Logger.Informational("left GetPayJsapiParams.")

	id, _ := w.GetInt(":id")
	billId, _ := w.GetInt(":bill_id")
	openId := w.GetString(":open_id")

	//get unified order
	Logger.Informational("Get unified order for bill:[%v].", billId)
	// 自己测试微信支付, DEMO
	bill := &models.BillInfo{
		Money:        10,
		Title:        "hello,world",
		TradeNoJsapi: "YL-20170101122055",
	}
	payParamInfo, unifiedOrderRespMap, retcode, err := models.UnifiedOrder(id, bill, openId, consts.TYPE_PAY__WECHAT__JSAPI, consts.TYPE_PAY_ENV__WECHAT__FABRIC_ORDER)
	if err != nil {
		Logger.Error("Get unified order for bill:[%v] error:%v", billId, err.Error())
		w.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  "pay unifiedorder error: " + err.Error(),
		}
		w.ServeJSON()
		return
	}
	if unifiedOrderRespMap["return_code"] == "SUCCESS" && unifiedOrderRespMap["result_code"] == "SUCCESS" {
		prepayId = unifiedOrderRespMap["prepay_id"]
	} else {
		Logger.Error("Get unified order for bill:[%v], response error", billId)
		Logger.Error("unifiedOrderRespMap:[%v]", unifiedOrderRespMap)
		w.Data["json"] = map[string]interface{}{
			"err_code": 32122,
			"err_msg":  unifiedOrderRespMap["err_code_des"],
		}
		w.ServeJSON()
		return
	}

	wcPayParams := make(map[string]string)
	wcPayParams["appId"] = payParamInfo.Appid
	wcPayParams["timeStamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	wcPayParams["nonceStr"] = "nonce_str"
	wcPayParams["package"] = "prepay_id=" + prepayId
	wcPayParams["signType"] = "MD5"
	wcPayParams["paySign"] = core.Sign(wcPayParams, payParamInfo.Appkey, nil)
	w.Data["json"] = map[string]interface{}{
		"wc_pay_params": wcPayParams,
		"err_code":      0,
		"err_msg":       "",
	}
	w.ServeJSON()
	return
}

// @params :id 内部公众号ID
// @params :bill_id 订单ID
// 说明: 二维码支付不需要openid
// @router /:id/pay/native_params/:bill_id [get]
func (w *WechatPayController) GetPayNativeParams() {
	Logger.Informational("enter GetPayNativeParams.")
	defer Logger.Informational("left GetPayNativeParams.")

	id, _ := w.GetInt(":id")
	billId, _ := w.GetInt(":bill_id")

	// 自己测试微信支付, DEMO
	bill := &models.BillInfo{
		Money:         10,
		Title:         "hello,world",
		TradeNoNative: fmt.Sprintf("YL-%s", time.Now().Format("20060102150405")),
	}
	//get unified order
	_, unifiedOrderRespMap, retcode, err := models.UnifiedOrder(id, bill, "", consts.TYPE_PAY__WECHAT__NATIVE, consts.TYPE_PAY_ENV__WECHAT__FABRIC_ORDER)
	if err != nil {
		Logger.Error("Get unified order for bill:[%v] error:%v", billId, err.Error())
		w.Data["json"] = map[string]interface{}{
			"err_code": retcode,
			"err_msg":  "pay unifiedorder error: " + err.Error(),
		}
	} else {
		if unifiedOrderRespMap["return_code"] == "SUCCESS" && unifiedOrderRespMap["result_code"] == "SUCCESS" {
			w.Data["json"] = map[string]interface{}{
				"err_code": 0,
				"err_msg":  "",
				"code_url": unifiedOrderRespMap["code_url"],
			}
		} else {
			Logger.Error("Get unified order for bill:[%v], response error", billId)
			Logger.Error("unifiedOrderRespMap:[%v]", unifiedOrderRespMap)
			w.Data["json"] = map[string]interface{}{
				"err_code": 7808,
				"err_msg":  unifiedOrderRespMap["err_code_des"],
			}
		}
	}
	w.ServeJSON()
	return
}

// @router /notification [post]
func (w *WechatPayController) NotifyUrl() {
	reader := strings.NewReader(string(w.Ctx.Input.RequestBody))
	reqMap, err := util.DecodeXMLToMap(reader)
	if err != nil {
		Logger.Error("request body::", string(w.Ctx.Input.RequestBody))
	} else {
		if reqMap["return_code"] == "SUCCESS" && reqMap["result_code"] == "SUCCESS" {
			fmt.Println("reqMap: ", reqMap)
			money := ConvertStrToInt(reqMap["total_fee"])
			switch reqMap["attach"] {
			case consts.TYPE_PAY_ENV__WECHAT__FABRIC_ORDER:
				// SaaS平台商家商品交易业务
				ppi := &types.ProductPayInfo{
					TradeNo:       reqMap["out_trade_no"],
					TransactionId: reqMap["transaction_id"],
					Money:         int64(money),
				}
				fmt.Printf("%v\n", ppi)
				//conf.SaleClient.Call(fmt.Sprintf("%s.%s", "sales", "UpdateSaleOrderPayInfo"), in, in)
			}
		}
	}
	w.Ctx.Output.Body([]byte("success"))
	return
}
