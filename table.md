# 创建公众号第三方平台库
```
CREATE DATABASE IF NOT EXISTS ycfm_wechat_official_accounts_platforms DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```
## 创建公众号消息发送记录表
```
CREATE TABLE IF NOT EXISTS `account_message_send_records` (
  `account_message_send_record_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `official_account_id` int(11) DEFAULT NULL COMMENT '公众号ID',
  `template_id` varchar(100) DEFAULT NULL COMMENT '微信消息模板ID',
  `content` varchar(2000) DEFAULT NULL COMMENT '发送消息内容',
  `receiver_user` varchar(100) DEFAULT NULL COMMENT '消息接受者',
  `status` smallint(6) DEFAULT NULL COMMENT '状态：-2：逻辑删除；1：有效',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`account_message_send_record_id`)
) ENGINE=InnoDB AUTO_INCREMENT=30 DEFAULT CHARSET=utf8mb4
```

### 创建公众号消息模板表
```
CREATE TABLE IF NOT EXISTS `account_message_templates` (
  `account_message_template_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `official_account_id` int(11) DEFAULT NULL COMMENT '公众号ID',
  `template_id` varchar(100) NOT NULL COMMENT '微信消息模板ID',
  `system_message_template_id` int(11) DEFAULT NULL COMMENT '系统微信消息模板ID',
  `status` smallint(6) DEFAULT NULL COMMENT '状态：-2：逻辑删除；1：有效',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`account_message_template_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4
```

###  创建公众平台行业信息表
```
CREATE TABLE IF NOT EXISTS `industry_code_querys` (
  `industry_code_query_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `main_type` int(11) DEFAULT NULL COMMENT '主行业类型',
  `main_industry_code` varchar(30) DEFAULT NULL COMMENT '主行业',
  `sec_industry_code` varchar(30) DEFAULT NULL COMMENT '副行业',
  `code_num` int(11) DEFAULT NULL COMMENT '代码',
  `status` smallint(6) DEFAULT NULL COMMENT '-20：逻辑删除；10：有效',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`industry_code_query_id`)
) ENGINE=InnoDB AUTO_INCREMENT=42 DEFAULT CHARSET=utf8mb4
```

### 创建公众号行业信息表
```
CREATE TABLE IF NOT EXISTS `official_account_industry_codes` (
  `official_account_industry_code_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `official_account_id` int(11) DEFAULT NULL COMMENT '内部系统公众号ID',
  `industry_id1` int(11) DEFAULT NULL COMMENT '公众号模板消息所属行业编号,每个月只允许修改一次',
  `industry_id2` int(11) DEFAULT NULL COMMENT '公众号模板消息所属行业编号,每个月只允许修改一次',
  `status` smallint(6) DEFAULT NULL COMMENT '-2: 逻辑删除；1：有效',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`official_account_industry_code_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4
```

###  创建所有公众号的用户表
```
CREATE TABLE IF NOT EXISTS `official_account_users` (
  `official_account_user_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `official_account_id` int(11) DEFAULT NULL COMMENT '公众号ID',
  `user_wx_info_id` int(11) DEFAULT NULL COMMENT '微信ID',
  `openid` varchar(50) DEFAULT NULL COMMENT '用户ID在该公众号下的唯一标识',
  `status` smallint(6) DEFAULT NULL COMMENT '状态：-2：逻辑删除；1：有效',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`official_account_user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4
```

###  创建公众号授权信息表
```
CREATE TABLE IF NOT EXISTS `official_accounts` (
  `official_account_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `nickname` varchar(50) DEFAULT NULL COMMENT '公众号微信昵称',
  `avartar_url` varchar(300) DEFAULT NULL COMMENT '公众号微信头像链接',
  `service_type_id` smallint(6) DEFAULT NULL COMMENT '公众号类型，0代表订阅号，1代表由历史老帐号升级后的订阅号，2代
表服务号',
  `verify_type_id` smallint(6) DEFAULT NULL COMMENT '认证类型，-1代表未认证，0代表微信认证，1代表新浪微博认证，2代
表腾讯微博认证，3代表已资质认证通过但还未通过名称认证，4代表已资质认证通过、还未通过名称认证，但通过了新浪微博认证
，5代表已资质认证通过、还未通过名称认证，但通过了腾讯微博认证',
  `original_id` varchar(40) DEFAULT NULL COMMENT '公众号的原始ID',
  `principal_name` varchar(300) DEFAULT NULL COMMENT '公众号的主体名称',
  `alias` varchar(100) DEFAULT NULL COMMENT '公众号所设置的微信号，可能为空',
  `business_info_open_store` smallint(6) DEFAULT NULL COMMENT '是否开通微信门店功能: 1: 未开通；2：已开通',
  `business_info_open_scan` smallint(6) DEFAULT NULL COMMENT '是否开通微信扫商品功能: 1: 未开通；2：已开通',
  `business_info_open_pay` smallint(6) DEFAULT NULL COMMENT '是否开通微信支付功能: 1: 未开通；2：已开通',
  `business_info_open_card` smallint(6) DEFAULT NULL COMMENT '是否开通微信卡券功能: 1: 未开通；2：已开通',
  `business_info_open_shake` smallint(6) DEFAULT NULL COMMENT '是否开通微信摇一摇功能: 1: 未开通；2：已开通',
  `qrcode_url` varchar(300) DEFAULT NULL COMMENT '二维码图片的URL',
  `appid` varchar(100) DEFAULT NULL COMMENT '公众号appid',
  `func_ids` varchar(100) DEFAULT NULL COMMENT '公众号开放给第三方平台的权限列表ID',
  `status` smallint(6) DEFAULT NULL COMMENT '-2: 逻辑删除；1：有效',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`official_account_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4
```

### 创建公众号支付参数表
```
CREATE TABLE IF NOT EXISTS `official_accounts_pay_params` (
  `official_accounts_pay_param_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `official_account_id` int(11) DEFAULT NULL COMMENT '公众号ID',
  `mch_id` varchar(30) DEFAULT NULL COMMENT '微信商户号ID',
  `name` varchar(100) DEFAULT NULL COMMENT '公众号公司名称',
  `appkey` varchar(50) DEFAULT NULL COMMENT '商户支付密钥key',
  `status` smallint(6) DEFAULT NULL COMMENT '状态：-2：逻辑删除；1：有效',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`official_accounts_pay_param_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4
```

### 创建公众号平台消息模板表
```
CREATE TABLE IF NOT EXISTS `system_message_templates` (
  `system_message_template_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `code` varchar(20) DEFAULT NULL COMMENT '消息编号',
  `title` varchar(100) DEFAULT NULL COMMENT '消息标题',
  `industry_code_query_id` int(11) DEFAULT NULL COMMENT '行业ID',
  `content` varchar(1000) NOT NULL COMMENT '模板内容',
  `status` smallint(6) DEFAULT NULL COMMENT '-2:逻辑删除；1：有效',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`system_message_template_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4
```

### 创建C端用户微信信息表
```
CREATE TABLE IF NOT EXISTS `user_wx_infos` (
  `user_wx_info_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `user_id` int(11) DEFAULT NULL COMMENT '用户ID，系统唯一标识',
  `nickname` varchar(50) DEFAULT NULL COMMENT '微信昵称',
  `sex` smallint(6) DEFAULT NULL COMMENT '性别：1.男性；2.女性; 3:未知',
  `province` varchar(100) DEFAULT NULL COMMENT '省份名称',
  `city` varchar(100) DEFAULT NULL COMMENT '城市名称',
  `country` varchar(30) DEFAULT NULL COMMENT '国家',
  `headimgurl` varchar(200) DEFAULT NULL COMMENT '头像url',
  `privilege` varchar(100) DEFAULT NULL COMMENT '用户特权信息，json 数组，如微信沃卡用户为（chinaunicom）',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`user_wx_info_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4
```

### 创建公众号微信消息接收记录表
```
CREATE TABLE IF NOT EXISTS `wechat_message_receipt_records` (
  `wechat_message_receipt_record_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `appid` varchar(50) DEFAULT NULL COMMENT '托管的公众号ID',
  `to_user_name` varchar(50) DEFAULT NULL COMMENT '公众号微信号',
  `from_user_name` varchar(50) DEFAULT NULL COMMENT '接收模板消息的用户的openid',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `msg_type` varchar(50) DEFAULT NULL COMMENT '消息类型是事件',
  `event` varchar(50) DEFAULT NULL COMMENT '事件为模板消息发送结束',
  `content` varchar(1000) NOT NULL COMMENT '接收文本内容',
  `msg_id` varchar(50) DEFAULT NULL COMMENT '消息id',
  `status` smallint(6) DEFAULT NULL COMMENT '消息发送状态',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`wechat_message_receipt_record_id`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8mb4
```
