package conf

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/1046102779/common/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/coreos/etcd/client"
	_ "github.com/go-sql-driver/mysql"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/clientselector"
	"github.com/smallnest/rpcx/codec"
)

var (
	SmsClient         *rpcx.Client
	Servers           []string
	WechatParam       *WechatParams
	WechatAuthTTL     *WechatAuthTTLInfo = new(WechatAuthTTLInfo)
	RpcAddr, EtcdAddr string
	KApi              client.KeysAPI
	ListenPaths       []string // 监听目录数组
	QueryAuthCodeTest string   // 全网测试授权的 query_auth_code
	CertificationDir  string
	NotifyUrl         string // 微信支付回调通知
	HostName          string // 微信回调域名

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

// 微信公众号开发第三方平台
type WechatParams struct {
	EncodingAesKey string
	Token          string
	AppId          string
	AppSecret      string
}

type AuthorizerManagementInfo struct {
	AuthorizerAccessToken          string // 授权方接口调用凭据（在授权的公众号具备API权限时，才有此返回值），也简称为令牌
	AuthorizerAccessTokenExpiresIn int
	AuthorizerRefreshToken         string // 刷新令牌主要用于公众号第三方平台获取和刷新已授权用户的access_token
}

type WechatAuthTTLInfo struct {
	ComponentVerifyTicket          string // 用于获取第三方平台接口调用凭据
	ComponentVerifyTicketExpiresIn int
	ComponentAccessToken           string // 第三方平台的下文中接口的调用凭据
	ComponentAccessTokenExpiresIn  int
	PreAuthCode                    string // 预授权码, 获取公众号第三方平台授权页面
	PreAuthCodeExpiresIn           int
	AuthorizerMap                  map[string]AuthorizerManagementInfo // 每一个公众号appid与自己的access_token和refresh_token的映射
}

func initWechatParams() {
	WechatParam = new(WechatParams)
	WechatParam.EncodingAesKey = beego.AppConfig.String("wechats::encodingAesKey")
	if "" == strings.TrimSpace(WechatParam.EncodingAesKey) {
		panic("param `wechats::encodingAesKey` empty")
	}
	WechatParam.Token = beego.AppConfig.String("wechats::token")
	if "" == strings.TrimSpace(WechatParam.Token) {
		panic("param `wechats::token`  empty")
	}
	WechatParam.AppId = beego.AppConfig.String("wechats::appid")
	if "" == strings.TrimSpace(WechatParam.AppId) {
		panic("param `wechats::appid` empty")
	}
	WechatParam.AppSecret = beego.AppConfig.String("wechats::appsecret")
	if "" == strings.TrimSpace(WechatParam.AppSecret) {
		panic("param `wechats::appsecret` empty")
	}
	NotifyUrl = beego.AppConfig.String("wechats::notify_url")
	if "" == strings.TrimSpace(NotifyUrl) {
		panic("param `wechats::notify_url` empty")
	}
	HostName = beego.AppConfig.String("wechats::hostname")
	if "" == strings.TrimSpace(HostName) {
		panic("param `wechats::hostname` empty")
	}
}

func initEtcdClient() {
	RpcAddr = strings.TrimSpace(beego.AppConfig.String("rpc::address"))
	EtcdAddr = strings.TrimSpace(beego.AppConfig.String("etcd::address"))
	if "" == EtcdAddr || "" == RpcAddr {
		panic("param `etcd::address || etcd::address` empty")
	}
	serverTemp := beego.AppConfig.String("rpc::servers")
	Servers = strings.Split(serverTemp, ",")
	// etcd 目录微信token过期监听
	c, err := client.New(client.Config{
		Endpoints: []string{EtcdAddr},
		Transport: client.DefaultTransport,
	})
	if err != nil {
		panic(err.Error())
	}
	KApi = client.NewKeysAPI(c)
	// 监听目录数据
	paths := beego.AppConfig.String("etcd::listenPaths")
	if "" != strings.TrimSpace(paths) {
		ListenPaths = strings.Split(paths, ",")
	}
}

func initDBConn() {
	var (
		err error
	)
	fmt.Println("hello,world")
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
	// 初始化公众号第三方平台开发参数
	initWechatParams()
	// 初始化Etcd客户端连接
	initEtcdClient()
	// 初始化DB连接
	initDBConn()
	// 获取RPC client
	if name, exist = utils.FindServer("sms", Servers); !exist {
		panic("params `sms` service not exist")
	}
	SmsClient = connRpcClient(name)
	// 初始化微信公众号支付证书目录
	CertificationDir = beego.AppConfig.String("upload_certification_dir::certification_file")
	if "" == strings.TrimSpace(CertificationDir) {
		panic("param `upload_certification_dir` empty")
	}
	WechatAuthTTL.AuthorizerMap = map[string]AuthorizerManagementInfo{}
	// orm debug
	DBDebug, err := beego.AppConfig.Bool("dev::debug")
	if err != nil {
		panic("app parameter `dev::debug` error:" + err.Error())
	}
	if DBDebug {
		orm.Debug = true
	}
}
