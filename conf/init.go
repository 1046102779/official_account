package conf

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hprose/hprose-golang/rpc"

	common "github.com/1046102779/common/rpc"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

var (
	WRServerRPC   = common.WxRelayServer{}
	UserServerRPC = common.UserServer{}
	Servers       []string
	RpcAddr       string

	QueryAuthCodeTest string // 全网测试授权的 query_auth_code
	CertificationDir  string
	NotifyUrl         string // 微信支付回调通知
	HostName          string // 微信回调域名

	// db::mysql
	DBHost               string
	DBPort               int
	DBUser, DBPawd       string
	DBName, DBCharset    string
	DBTimeLoc            string
	DBMaxIdle, DBMaxConn int
	DBDebug              bool

	WechatOpenPlatformTestFeatureFlag bool = false

	// hprose clients
	WxRelayServer__Hprose__RPC_Client *rpc.HTTPClient
	User__Hprose__RPC_Client          *rpc.HTTPClient
)

func initHproseClient() {
	// wx_relay_server
	addrs := beego.AppConfig.Strings("rpc::wrs__hprose__rpc_address") // ;
	if len(addrs) <= 0 {
		panic("[config] param `rpc::wrs__hprose__rpc_address` empty")
	}
	WxRelayServer__Hprose__RPC_Client = rpc.NewHTTPClient(addrs...)
	WxRelayServer__Hprose__RPC_Client.UseService(&WRServerRPC)

	// user
	addrs = beego.AppConfig.Strings("rpc::user__hprose__rpc_address")
	if len(addrs) <= 0 {
		panic("[config] param `rpc::user__hprose__rpc_address` empty")
	}
	User__Hprose__RPC_Client = rpc.NewHTTPClient(addrs...)
	User__Hprose__RPC_Client.UseService(&UserServerRPC)
}

func initDBConn() {
	var (
		err error
	)
	DBHost = strings.TrimSpace(beego.AppConfig.String("db::host"))
	if "" == DBHost {
		panic("app parameter `db::host` empty")
	}

	DBPort, err = beego.AppConfig.Int("db::port")
	if err != nil {
		panic("app parameter `db::port` error")
	}
	DBUser = strings.TrimSpace(beego.AppConfig.String("db::user"))
	if "" == DBUser {
		panic("app parameter `db::user` empty")
	}

	DBPawd = strings.TrimSpace(beego.AppConfig.String("db::pawd"))
	if "" == DBPawd {
		panic("app parameter `db::pawd` empty")
	}

	DBName = strings.TrimSpace(beego.AppConfig.String("db::name"))
	if "" == DBName {
		panic("app parameter `db::name` empty")
	}

	DBCharset = strings.TrimSpace(beego.AppConfig.String("db::charset"))
	if "" == DBCharset {
		panic("app parameter `db::charset` empty")
	}

	DBTimeLoc = strings.TrimSpace(beego.AppConfig.String("db::time_loc"))
	if "" == DBTimeLoc {
		panic("app parameter `db::time_loc` empty")
	}

	DBMaxIdle, err = beego.AppConfig.Int("db::max_idle")
	if err != nil {
		panic("app parameter `db::max_idle` error")
	}
	DBMaxConn, err = beego.AppConfig.Int("db::max_conn")
	if err != nil {
		panic("app parameter `db::max_conn` error")
	}
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&loc=%s", DBUser, DBPawd, DBHost, DBPort, DBName, DBCharset, url.QueryEscape(DBTimeLoc))

	err = orm.RegisterDataBase("default", "mysql", dataSourceName, DBMaxIdle, DBMaxConn)
	if err != nil {
		panic(err)
	}
}

func init() {
	var err error
	// 初始化Hprose客户端
	initHproseClient()
	// 初始化DB连接
	initDBConn()
	RpcAddr = beego.AppConfig.String("rpc::official_account__hprose__rpc_address")
	if len(RpcAddr) <= 0 {
		panic("param `rpc::official_account__hprose__rpc_address` empty")
	}

	WechatOpenPlatformTestFeature := strings.TrimSpace(beego.AppConfig.String("wechat::wechat_open_platform_test_feture"))
	if WechatOpenPlatformTestFeature == "ON" {
		WechatOpenPlatformTestFeatureFlag = true
	}

	// 初始化微信公众号支付证书目录
	CertificationDir = beego.AppConfig.String("upload_certification_dir::certification_file")
	if "" == strings.TrimSpace(CertificationDir) {
		panic("param `upload_certification_dir` empty")
	}
	// orm debug
	DBDebug, err := beego.AppConfig.Bool("dev::debug")
	if err != nil {
		panic("app parameter `dev::debug` error:" + err.Error())
	}
	if DBDebug {
		orm.Debug = true
	}
}
