package consts

const (
	CODE__SUCCESS = 0

	CODE__FIRST = "1"

	// position
	CODE__POSITION__SELLER           = "seller"
	CODE__POSITION__CAHIER           = "cashier"
	CODE__POSITION__TREASURER        = "treasurer"
	CODE__POSITION__WAREHOUSE_KEEPER = "warehouse_keeper"
	CODE__POSITION__SHOP_MANAGER     = "shop_manager"
	CODE__POSITION__GENERAL_MANAGER  = "general_manager"
	CODE__POSITION__COMPANY_ADMIN    = "company_admin"

	// system  template message
	// 出库提醒
	CODE__TEMPLATE_MESSAGE__WAREHOUSE_MANAGER__WAITING_OUTBOUNDS = "出库提醒"
	// 销售员或店长-接收已出库提醒 , 销售员或店长-接收已物流提醒
	CODE__TEMPLATE_MESSAGE__SALE_OR_SHOPMANAGER__WAITING_OUTBOUNDS_OR_LOGISTAICS = "货物出库提醒"
	// 门店订购提醒通知
	CODE__TEMPLATE_MESSAGE__CUSTOMER__PLACE_ORDER_SUCCESS = "门店订购提醒通知"
	// 客户-总价格确定时接收提醒
	CODE__TEMPLATE_MESSAGE__CUSTOMER__PRICE_CLAIM = "订单状态更新通知"
	// 客户-待客户验收状态时接收提醒
	CODE__TEMPLATE_MESSAGE__CUSTOMER__TO_BE_ACCEPTION = "商品自提通知"
	// 总经理-每晚8点接收当天营业额统计情况的提醒
	CODE__TEMPLATE_MESSAGE__SHOP_MANAGER__STATISTICS = "营业情况明细通知"
)
