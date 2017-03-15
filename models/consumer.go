// kafka消息队列
package models

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/1046102779/official_account/common/consts"
	"github.com/1046102779/official_account/common/utils"
	"github.com/1046102779/official_account/conf"
	pb "github.com/1046102779/official_account/igrpc"
	. "github.com/1046102779/official_account/logger"

	"github.com/pkg/errors"
)

var (
	wg sync.WaitGroup
)

type MessageInfoReq struct {
	TemplateId string          `json:"template_id"`
	ToUser     string          `json:"touser"`
	Content    json.RawMessage `json:"content"`
}

type Values struct {
	Value string `json:"value"`
	Color string `json:"color"`
}

// 通用模板消息内容
type Content struct {
	First    Values `json:"first"`
	Keyword1 Values `json:"keyword1"`
	Keyword2 Values `json:"keyword2"`
	Keyword3 Values `json:"keyword3"`
	Keyword4 Values `json:"keyword4"`
	Keyword5 Values `json:"keyword5"`
	Remark   Values `json:"remark"`
	//Keywords map[string]Values
}

// 封装消息为模板消息
// 输入：字段消息内容
// 输出：消息模板内容
func getTemplateMessageContent(first string, remarks string, keywords []string) (bts []byte, err error) {
	content := &Content{}
	content.First = Values{
		Value: first,
		Color: fmt.Sprintf("#173177"),
	}
	content.Remark = Values{
		Value: remarks,
		Color: fmt.Sprintf("#173177"),
	}
	//content.Keywords = make(map[string]Values, 0)
	for index := 0; keywords != nil && index < len(keywords); index++ {
		switch index + 1 {
		case 1:
			content.Keyword1 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		case 2:
			content.Keyword2 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		case 3:
			content.Keyword3 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		case 4:
			content.Keyword4 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		case 5:
			content.Keyword5 = Values{
				Value: keywords[index],
				Color: fmt.Sprintf("#173177"),
			}
		}
	}
	if bts, err = json.Marshal(content); err != nil {
		err = errors.Wrap(err, "getTemplateMessageContent")
		return
	}
	return
}

// 发送订单模板消息
func sendSaleOrderTemplateMessage(companyId int, userId int, customerId int, templateName string, bts []byte) (err error) {
	Logger.Info("[%v.%v] enter sendSaleOrderTemplateMessage.", companyId, userId)
	defer Logger.Info("[%v.%v] left sendSaleOrderTemplateMessage.", companyId, userId)
	var (
		templateId, wxOpenId string
		officialAccountId    int
	)
	// ::TODO get templateId
	templateId, _, err = getTemplateIdByCompanyIdAndName(companyId, templateName)
	if err != nil {
		err = errors.Wrap(err, "sendSaleOrderTemplateMessage")
		return
	}
	// ::TODO get wxOpenId
	if userId > 0 {
		wxOpenId, _, err = getWxOpenIdByCompanyIdAndUserId(companyId, userId)
	} else {
		wxOpenId, _, err = getWxOpenIdByCompanyIdAndCustomerId(companyId, customerId)
	}
	if err != nil {
		err = errors.Wrap(err, "sendSaleOrderTemplateMessage")
		return
	}
	officialAccountId, _, err = getOfficialAccountIdByCompanyId(companyId)
	if err != nil {
		err = errors.Wrap(err, "sendSaleOrderTemplateMessage")
		return
	}
	message := &MessageInfoReq{
		Content:    bts,
		TemplateId: templateId,
		ToUser:     wxOpenId,
	}
	if bts, err = json.Marshal(message); err != nil {
		err = errors.Wrap(err, "sendSaleOrderTemplateMessage")
		return
	}
	ConsumeAccountTemplateMessage(officialAccountId, bts)
	return
}

// 获取仓管员待出库消息提醒的内容
func sendWaitingToOutboundsMessage(saleOrderItemId int) (err error) {
	Logger.Info("[%v] enter sendWaitingToOutboundsMessage.", saleOrderItemId)
	defer Logger.Info("[%v] left sendWaitingToOutboundsMessage.", saleOrderItemId)
	var (
		companyId, userId int
		bts               []byte
	)
	// 获取模板ID
	saleOrder := &pb.SaleOrder{
		SaleOrderItemId: int64(saleOrderItemId),
	}
	now := time.Now()
	warehouseManager := &pb.WarehouseManagers{}
	conf.SaleClient.Call(fmt.Sprintf("%s.%s", "sales", "GetSaleOrderItemsWaitingToOuts"), saleOrder, warehouseManager)
	if warehouseManager.WarehouseManagerMessages == nil && len(warehouseManager.WarehouseManagerMessages) <= 0 {
		return
	}
	for index := 0; index < len(warehouseManager.WarehouseManagerMessages); index++ {
		first := fmt.Sprintf("%s仓管, 您有一个新订单需要您进行出库操作！", warehouseManager.WarehouseManagerMessages[index].Name)
		var value string
		tempColorPatterns := warehouseManager.WarehouseManagerMessages[index].ColorPatternRelationships
		for subIndex := 0; tempColorPatterns != nil && subIndex < len(tempColorPatterns); subIndex++ {
			if tempColorPatterns[subIndex].UnitType == 10 { // UNIT_TYPE_CLOTH
				if value == "" {
					value = fmt.Sprintf("%s%s%d%s", tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "匹")
				} else {
					value = fmt.Sprintf("%s, %s%s%d%s", value, tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "匹")
				}
			} else {
				if value == "" {
					value = fmt.Sprintf("%s%s%d%s", tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "米")
				} else {
					value = fmt.Sprintf("%s, %s%s%d%s", value, tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "米")
				}
			}
			companyId = int(tempColorPatterns[subIndex].CompanyId)
		}
		keywords := []string{}
		keywords = append(keywords, warehouseManager.WarehouseManagerMessages[index].SaleOrderNo)
		keywords = append(keywords, value)
		keywords = append(keywords, now.Format("2006-01-02 15:04:05"))
		remark := fmt.Sprintf("请及时处理，辛苦了！")
		bts, err = getTemplateMessageContent(first, remark, keywords)
		if err != nil {
			err = errors.Wrap(err, "sendWaitingToOutboundsMessage")
			return
		}
		userId = int(warehouseManager.WarehouseManagerMessages[index].Id)
		if err = sendSaleOrderTemplateMessage(companyId, userId, 0, consts.CODE__TEMPLATE_MESSAGE__WAREHOUSE_MANAGER__WAITING_OUTBOUNDS, bts); err != nil {
			err = errors.Wrap(err, "sendWaitingToOutboundsMessage")
			return
		}
	}
	return
}

// 销售员或店长-接收已出库提醒(货物出库提醒)
func sendOutboundsMessage(saleOrderItemId int) (err error) {
	var (
		companyId, userId int = 0, 0
		keywords          []string
		remark            string
		bts               []byte
	)
	// 模板内容
	saleOrder := &pb.SaleOrder{
		SaleOrderItemId: int64(saleOrderItemId),
	}
	shopMessage := &pb.ShopMessage{}
	conf.SaleClient.Call(fmt.Sprintf("%s.%s", "sales", "GetSaleOrderItemOutbounds"), saleOrder, shopMessage)
	first := fmt.Sprintf("商品已从%s出库，待物流！", shopMessage.OutbundedAtPlaceName)
	keywords = append(keywords, time.Now().String())
	keywords = append(keywords, shopMessage.OutbundedAtPlaceName)
	keywords = append(keywords, shopMessage.SaleOrderNo)
	if shopMessage.ColorPatternRelationships == nil || len(shopMessage.ColorPatternRelationships) <= 0 {
		err = errors.New("sale order item  not exist for color-pattern")
		return
	}
	companyId = int(shopMessage.CompanyId)
	userId = int(shopMessage.Id)
	if shopMessage.ColorPatternRelationships[0].UnitType == 10 { // CLOTH
		remark = fmt.Sprintf("%s%s%d%s", shopMessage.ColorPatternRelationships[0].Name, shopMessage.ColorPatternRelationships[0].ColorName, shopMessage.ColorPatternRelationships[0].Amount, "匹")
	} else {
		remark = fmt.Sprintf("%s%s%d%s", shopMessage.ColorPatternRelationships[0].Name, shopMessage.ColorPatternRelationships[0].ColorName, shopMessage.ColorPatternRelationships[0].Amount, "米")
	}
	bts, err = getTemplateMessageContent(first, remark, keywords)
	if err != nil {
		err = errors.Wrap(err, "sendOutboundsMessage")
		return
	}
	err = sendSaleOrderTemplateMessage(companyId, userId, 0, consts.CODE__TEMPLATE_MESSAGE__SALE_OR_SHOPMANAGER__WAITING_OUTBOUNDS_OR_LOGISTAICS, bts)
	if err != nil {
		err = errors.Wrap(err, "sendOutboundsMessage")
		return
	}
	return
}

// 销售员或店长-接收已物流提醒()
func sendLogisticsMessage(saleOrderItemId int) (err error) {
	var (
		companyId, userId int = 0, 0
		keywords          []string
		remark            string
		bts               []byte
	)
	// 模板内容
	saleOrder := &pb.SaleOrder{
		SaleOrderItemId: int64(saleOrderItemId),
	}
	shopMessage := &pb.ShopMessage{}
	conf.SaleClient.Call(fmt.Sprintf("%s.%s", "sales", "GetSaleOrderItemOutbounds"), saleOrder, shopMessage)
	first := fmt.Sprintf("商品已从%s出库，发往%s，请注意验收！", shopMessage.OutbundedAtPlaceName, shopMessage.DeliveriedAtPlaceName)
	keywords = append(keywords, time.Now().String())
	keywords = append(keywords, shopMessage.DeliveriedAtPlaceName)
	keywords = append(keywords, shopMessage.SaleOrderNo)
	if shopMessage.ColorPatternRelationships == nil || len(shopMessage.ColorPatternRelationships) <= 0 {
		err = errors.New("sale order item  not exist for color-pattern")
		return
	}
	companyId = int(shopMessage.CompanyId)
	userId = int(shopMessage.Id)
	if shopMessage.ColorPatternRelationships[0].UnitType == 10 { // CLOTH
		remark = fmt.Sprintf("%s%s%d%s", shopMessage.ColorPatternRelationships[0].Name, shopMessage.ColorPatternRelationships[0].ColorName, shopMessage.ColorPatternRelationships[0].Amount, "匹")
	} else {
		remark = fmt.Sprintf("%s%s%d%s", shopMessage.ColorPatternRelationships[0].Name, shopMessage.ColorPatternRelationships[0].ColorName, shopMessage.ColorPatternRelationships[0].Amount, "米")
	}
	bts, err = getTemplateMessageContent(first, remark, keywords)
	if err != nil {
		err = errors.Wrap(err, "sendLogisticsMessage")
		return
	}
	err = sendSaleOrderTemplateMessage(companyId, userId, 0, consts.CODE__TEMPLATE_MESSAGE__SALE_OR_SHOPMANAGER__WAITING_OUTBOUNDS_OR_LOGISTAICS, bts)
	if err != nil {
		err = errors.Wrap(err, "sendLogisticsMessage")
		return
	}
	return
}

// 客户-下单成功时提醒(门店订购提醒通知)
func sendPlaceOrderSuccessMessage(saleOrderId int) (err error) {
	var (
		companyId, userId int = 0, 0
		keywords          []string
		bts               []byte
	)
	// 模板内容
	saleOrder := &pb.SaleOrder{
		SaleOrderId: int64(saleOrderId),
	}
	shopMessage := &pb.ShopMessage{}
	conf.SaleClient.Call(fmt.Sprintf("%s.%s", "sales", "GetSaleOrderPlaceOrderSuccess"), saleOrder, shopMessage)
	first := fmt.Sprintf("您的订单已经生成！")
	remark := fmt.Sprintf("如有疑问,具体请拨打订购门店的电话!")
	if shopMessage.ColorPatternRelationships == nil || len(shopMessage.ColorPatternRelationships) <= 0 {
		err = errors.New("sale order item  not exist for color-pattern")
		return
	}
	companyId = int(shopMessage.CompanyId)
	userId = int(shopMessage.Id)
	tempColorPatterns := shopMessage.ColorPatternRelationships
	var value string
	for subIndex := 0; tempColorPatterns != nil && subIndex < len(tempColorPatterns); subIndex++ {
		if tempColorPatterns[subIndex].UnitType == 10 { // UNIT_TYPE_CLOTH
			if value == "" {
				value = fmt.Sprintf("%s%s%d%s", tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "匹")
			} else {
				value = fmt.Sprintf("%s, %s%s%d%s", value, tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "匹")
			}
		} else {
			if value == "" {
				value = fmt.Sprintf("%s%s%d%s", tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "米")
			} else {
				value = fmt.Sprintf("%s, %s%s%d%s", value, tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "米")
			}
		}
	}
	keywords = append(keywords, shopMessage.DeliveriedAtPlaceName)
	keywords = append(keywords, shopMessage.SaleOrderNo)
	keywords = append(keywords, value)
	keywords = append(keywords, fmt.Sprintf("%d", 1))
	keywords = append(keywords, time.Now().String())
	bts, err = getTemplateMessageContent(first, remark, keywords)
	if err != nil {
		err = errors.Wrap(err, "sendPlaceOrderSuccessMessage")
		return
	}
	err = sendSaleOrderTemplateMessage(companyId, 0, userId, consts.CODE__TEMPLATE_MESSAGE__CUSTOMER__PLACE_ORDER_SUCCESS, bts)
	if err != nil {
		err = errors.Wrap(err, "sendPlaceOrderSuccessMessage")
		return
	}
	return
}

// 客户-总价格确定时接收提醒(订单状态更新通知)
func sendDetermineTotalPrice(saleOrderId int) (err error) {
	var (
		userId, companyId int
		keywords          []string
		bts               []byte
	)
	saleOrder := &pb.SaleOrder{
		SaleOrderId: int64(saleOrderId),
	}
	shopMessage := &pb.ShopMessage{}
	conf.SaleClient.Call(fmt.Sprintf("%s.%s", "sales", "GetDetermineTotalPrice"), saleOrder, shopMessage)
	first := fmt.Sprintf("您的订单金额已生成")
	keywords = append(keywords, shopMessage.DeliveriedAtPlaceName)
	keywords = append(keywords, shopMessage.Tel)
	keywords = append(keywords, shopMessage.SaleOrderNo)
	keywords = append(keywords, "配送中")
	keywords = append(keywords, fmt.Sprintf("%d", shopMessage.TotalAmount/100))
	tempColorPatterns := shopMessage.ColorPatternRelationships
	var value string
	for subIndex := 0; tempColorPatterns != nil && subIndex < len(tempColorPatterns); subIndex++ {
		if tempColorPatterns[subIndex].UnitType == 10 { // UNIT_TYPE_CLOTH
			if value == "" {
				value = fmt.Sprintf("%s%s%d%s", tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "匹")
			} else {
				value = fmt.Sprintf("%s, %s%s%d%s", value, tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "匹")
			}
		} else {
			if value == "" {
				value = fmt.Sprintf("%s%s%d%s", tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "米")
			} else {
				value = fmt.Sprintf("%s, %s%s%d%s", value, tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "米")
			}
		}
	}
	remark := value
	companyId = int(shopMessage.CompanyId)
	userId = int(shopMessage.Id)
	if bts, err = getTemplateMessageContent(first, remark, keywords); err != nil {
		err = errors.Wrap(err, "sendDetermineTotalPrice")
		return
	}
	err = sendSaleOrderTemplateMessage(companyId, 0, userId, consts.CODE__TEMPLATE_MESSAGE__CUSTOMER__PRICE_CLAIM, bts)
	if err != nil {
		err = errors.Wrap(err, "sendDetermineTotalPrice")
		return
	}
	return
}

// 客户-待客户验收状态时接收提醒(商铺自提通知(没有取货编码))
func sendAcceptionMessage(saleOrderId int) (err error) {
	var (
		companyId, userId int
		keywords          []string
		bts               []byte
	)
	saleOrder := &pb.SaleOrder{
		SaleOrderId: int64(saleOrderId),
	}
	shopMessage := &pb.ShopMessage{}
	conf.SaleClient.Call(fmt.Sprintf("%s.%s", "sales", "GetDetermineTotalPrice"), saleOrder, shopMessage)
	first := fmt.Sprintf("您好，您订购的商品已经到货！快去自提吧！")
	tempColorPatterns := shopMessage.ColorPatternRelationships
	var value string
	for subIndex := 0; tempColorPatterns != nil && subIndex < len(tempColorPatterns); subIndex++ {
		if tempColorPatterns[subIndex].UnitType == 10 { // UNIT_TYPE_CLOTH
			if value == "" {
				value = fmt.Sprintf("%s%s%d%s", tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "匹")
			} else {
				value = fmt.Sprintf("%s, %s%s%d%s", value, tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "匹")
			}
		} else {
			if value == "" {
				value = fmt.Sprintf("%s%s%d%s", tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "米")
			} else {
				value = fmt.Sprintf("%s, %s%s%d%s", value, tempColorPatterns[subIndex].Name, tempColorPatterns[subIndex].ColorName, tempColorPatterns[subIndex].Amount, "米")
			}
		}
	}
	keywords = append(keywords, shopMessage.SaleOrderNo)
	keywords = append(keywords, value)
	keywords = append(keywords, "已到货")
	keywords = append(keywords, shopMessage.DeliveriedAtPlaceName)
	keywords = append(keywords, shopMessage.Tel)
	remark := fmt.Sprintf("请及时上门验货")
	companyId = int(shopMessage.CompanyId)
	userId = int(shopMessage.Id)
	if bts, err = getTemplateMessageContent(first, remark, keywords); err != nil {
		err = errors.Wrap(err, "sendDetermineTotalPrice")
		return
	}
	err = sendSaleOrderTemplateMessage(companyId, 0, userId, consts.CODE__TEMPLATE_MESSAGE__CUSTOMER__TO_BE_ACCEPTION, bts)
	if err != nil {
		err = errors.Wrap(err, "sendAcceptionMessage")
		return
	}
	return
}

// 总经理-每晚8点接收当天营业额统计情况的提醒(营业情况明细通知)
func sendStatisticsMessage() (err error) {
	var (
		companyId, userId int
		bts               []byte
	)
	now := time.Now()
	earliesTime, _ := utils.GetEarliestDate(&now)
	managerMessage := &pb.GeneralMangerMessages{}
	conf.SaleClient.Call(fmt.Sprintf("%s.%s", "sales", "GetAllTurnoverStatistics"), managerMessage, managerMessage)
	for index := 0; managerMessage.GeneralMangerMessages != nil && index < len(managerMessage.GeneralMangerMessages); index++ {
		keywords := []string{}
		first := fmt.Sprintf("今天营业情况如下：")
		keywords = append(keywords, fmt.Sprintf("%s-%s", earliesTime.String(), now.String()))
		keywords = append(keywords, fmt.Sprintf("总订单数: %d,  总金额：%d, 实收总金额：%d",
			managerMessage.GeneralMangerMessages[index].TurnoverStatistics.TotalOrders,
			managerMessage.GeneralMangerMessages[index].TurnoverStatistics.TotalAmount,
			managerMessage.GeneralMangerMessages[index].TurnoverStatistics.ActualAmount))
		remark := fmt.Sprintf("感谢使用")
		companyId = int(managerMessage.GeneralMangerMessages[index].CompanyId)
		userId = int(managerMessage.GeneralMangerMessages[index].UserId)
		if bts, err = getTemplateMessageContent(first, remark, keywords); err != nil {
			err = errors.Wrap(err, "sendDetermineTotalPrice")
			return
		}
		err = sendSaleOrderTemplateMessage(companyId, userId, 0, consts.CODE__TEMPLATE_MESSAGE__SHOP_MANAGER__STATISTICS, bts)
		if err != nil {
			err = errors.Wrap(err, "sendAcceptionMessage")
			return
		}
	}
	return
}

func ConsumeAccountTemplateMessage(officialAccountId int, bts []byte) (err error) {
	Logger.Info("[%v] enter ConsumeAccountTemplateMessage.", string(bts))
	defer Logger.Info("[%v] left ConsumeAccountTemplateMessage.", string(bts))
	var (
		msgId string
	)

	var (
		messageInfo *MessageInfoReq = new(MessageInfoReq)
	)
	if err = json.Unmarshal(bts, messageInfo); err != nil {
		Logger.Error(err.Error())
		return
	}
	if msgId, _, err = SendAccountMessage(officialAccountId, messageInfo.ToUser, messageInfo.TemplateId, messageInfo.Content); err != nil {
		Logger.Error(err.Error())
		return
	}
	fmt.Println("msg_id: ", msgId)
	return
}
