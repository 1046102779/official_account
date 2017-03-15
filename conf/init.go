package conf

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/1046102779/official_account/common/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/clientselector"
	"github.com/smallnest/rpcx/codec"
)

var (
	AccountClient, SmsClient        *rpcx.Client
	SaleClient, WxRelayServerClient *rpcx.Client
	Servers                         []string
	RpcAddr, EtcdAddr               string

	QueryAuthCodeTest string // 全网测试授权的 query_auth_code
	CertificationDir  string
	NotifyUrl         string // 微信支付回调通知
	HostName          string // 微信回调域名
	//FrontendWechatCallUrl string // 前端接收微信的请求访问地址，接收appid和code
	//FrontendOfficialAccountCallUrl string // 前端接收微信公众号的请求访问地址，接收query_auth_code

	// db::mysql
	DBHost    string
	DBPort    int
	DBUser    string
	DBPawd    string
	DBName    string
	DBCharset string
	DBTimeLoc string
	DBMaxIdle int
	DBMaxConn int
	DBDebug   bool
)

func initEtcdClient() {
	RpcAddr = strings.TrimSpace(beego.AppConfig.String("rpc::address"))
	EtcdAddr = strings.TrimSpace(beego.AppConfig.String("etcd::address"))
	if "" == EtcdAddr || "" == RpcAddr {
		panic("param `etcd::address || etcd::address` empty")
	}
	serverTemp := beego.AppConfig.String("rpc::servers")
	Servers = strings.Split(serverTemp, ",")
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

func connRpcClient(appName string) (client *rpcx.Client) {
	s := clientselector.NewEtcdClientSelector([]string{EtcdAddr}, fmt.Sprintf("/%s/%s/%s", beego.BConfig.RunMode, "rpcx", appName), time.Minute, rpcx.RandomSelect, time.Minute)
	client = rpcx.NewClient(s)
	client.FailMode = rpcx.Failover
	client.ClientCodecFunc = codec.NewProtobufClientCodec
	return
}

func init() {
	var (
		err   error
		exist bool = false
		name  string
	)
	// 初始化Etcd客户端连接
	initEtcdClient()
	// 初始化DB连接
	initDBConn()

	// 获取RPC client
	if name, exist = utils.FindServer("sms", Servers); !exist {
		panic("params `sms` service not exist")
	}
	SmsClient = connRpcClient(name)
	if name, exist = utils.FindServer("accounts", Servers); !exist {
		panic("params `accounts` service not exist")
	}
	AccountClient = connRpcClient(name)

	if name, exist = utils.FindServer("sales", Servers); !exist {
		panic("params `sales` service not exist")
	}
	SaleClient = connRpcClient(name)
	if name, exist = utils.FindServer("wx_relay_server", Servers); !exist {
		panic("param `wx_relay_server` service not exist")
	}
	WxRelayServerClient = connRpcClient(name)
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
