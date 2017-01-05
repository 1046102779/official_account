package models

import (
	"strings"
	"time"

	utils "github.com/1046102779/common"
	. "github.com/1046102779/official_account/conf"
	. "github.com/1046102779/official_account/logger"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// ETCD读写、监听相关操作
type EtcdClient struct{}

func (t *EtcdClient) ForceMKDir(dir string) (retcode int, err error) {
	Logger.Info("[%v] enter ForceMKDir.", dir)
	defer Logger.Info("[%v] left ForceMKDir.", dir)
	if "" == strings.TrimSpace(dir) {
		err = errors.New("param `dir` empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	_, err = KApi.Set(context.Background(), dir, "", &client.SetOptions{
		PrevExist: client.PrevIgnore,
		Dir:       true,
	})
	if err != nil {
		err = errors.Wrap(err, "ForceMKDir ")
		retcode = utils.ETCD_CREATE_DIR_ERROR
		return
	}
	return
}

func (t *EtcdClient) Put(key string, value string, ttl int) (retcode int, err error) {
	Logger.Info("[%v] enter Put.", key)
	defer Logger.Info("[%v] left Put.", key)
	if "" == strings.TrimSpace(key) {
		err = errors.New("params `key` empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	var (
		opt *client.SetOptions
	)
	if ttl > 0 {
		opt = &client.SetOptions{
			TTL: time.Duration(ttl) * time.Second,
		}
	}
	_, err = KApi.Set(context.Background(), key, value, opt)
	if err != nil {
		err = errors.Wrap(err, "etcdclient Put ")
		retcode = utils.ETCD_CREATE_KEY_ERROR
		return
	}
	return
}

// 读取键值对
func (t *EtcdClient) Get(key string) (pairs map[string]string, retcode int, err error) {
	Logger.Info("[%v] enter Get.", key)
	defer Logger.Info("[%v] left Get.", key)
	var (
		response *client.Response
	)
	pairs = map[string]string{}
	if "" == strings.TrimSpace(key) {
		err = errors.New("param `key` empty")
		retcode = utils.SOURCE_DATA_ILLEGAL
		return
	}
	response, err = KApi.Get(context.Background(), key, nil)
	if err != nil {
		err = errors.Wrap(err, "etcdclient Get")
		retcode = utils.ETCD_READ_KEY_ERROR
		return
	} else {
		retcode, err = recursiveNodes(t, response.Node, pairs)
		if err != nil {
			err = errors.Wrap(err, "etcdclient Get ")
			return
		}
	}
	return
}

func recursiveNodes(c *EtcdClient, node *client.Node, pairs map[string]string) (retcode int, err error) {
	var (
		resp *client.Response
	)
	if !node.Dir {
		pairs[node.Key] = node.Value
		return
	}
	for _, subnode := range node.Nodes {
		if !subnode.Dir {
			pairs[subnode.Key] = subnode.Value
		} else {
			resp, err = KApi.Get(context.Background(), subnode.Key, nil)
			if err != nil {
				retcode = utils.ETCD_READ_KEY_ERROR
				return
			} else {
				retcode, err = recursiveNodes(c, resp.Node, pairs)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (t *EtcdClient) Watch(key string) {
	Logger.Info("[%v] enter Watch.", key)
	defer Logger.Info("[%v] left Watch.", key)
	var (
		fields              []string
		lastField           string // SRC: /wechats/thirdplatform/ComponentVerifyTicket  RESULT: ComponentVerifyTicket
		appid               string // 公众号appid
		token, refreshToken string
		expiresIn           int
	)
	watcher := KApi.Watcher(key, &client.WatcherOptions{
		Recursive: true,
	})
	go func() {
		for {
			response, err := watcher.Next(context.Background())
			if err != nil {
				Logger.Error(err.Error())
				continue
			}
			switch response.Action {
			case "expire":
				fields = strings.Split(response.PrevNode.Key, "/")
				if fields != nil && len(fields) > 0 {
					lastField = fields[len(fields)-1]
				}
				switch lastField {
				case "ComponentAccessToken":
					WechatAuthTTL.ComponentAccessToken, WechatAuthTTL.ComponentAccessTokenExpiresIn, _, err = GetComponentAccessToken(WechatParam.AppId, WechatParam.AppSecret, WechatAuthTTL.ComponentVerifyTicket)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
					_, err = t.Put(response.PrevNode.Key, WechatAuthTTL.ComponentAccessToken, WechatAuthTTL.ComponentAccessTokenExpiresIn-1200)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
				case "PreAuthCode":
					WechatAuthTTL.PreAuthCode, WechatAuthTTL.PreAuthCodeExpiresIn, _, err = GetPreAuthCode(WechatAuthTTL.ComponentAccessToken)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
					_, err = t.Put(response.PrevNode.Key, WechatAuthTTL.PreAuthCode, WechatAuthTTL.PreAuthCodeExpiresIn-300)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
				case "AuthorizerAccessToken":
					appid = fields[len(fields)-2]
					token, expiresIn, refreshToken, _, err = RefreshToken(WechatAuthTTL.ComponentAccessToken, WechatAuthTTL.AuthorizerMap[appid].AuthorizerRefreshToken, appid)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
					WechatAuthTTL.AuthorizerMap[appid] = AuthorizerManagementInfo{
						AuthorizerAccessToken:          token,
						AuthorizerAccessTokenExpiresIn: expiresIn,
						AuthorizerRefreshToken:         refreshToken,
					}
					_, err = t.Put(response.PrevNode.Key, WechatAuthTTL.AuthorizerMap[appid].AuthorizerAccessToken, WechatAuthTTL.AuthorizerMap[appid].AuthorizerAccessTokenExpiresIn-1200)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
				}
			}
		}
	}()
	return
}
