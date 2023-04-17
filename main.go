package main

import (
	"net/url"
	"net/http"
	"github.com/golang-jwt/jwt"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/store"
	"fmt"
	"oauthsso/config"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"oauthsso/pkg/session"
	//"github.com/go-redis/redis"
)

var srv *server.Server
var mgr *manage.Manager

func main() {
	//config init
	config.Setup()
	// init db connection
	// congifure db in app.yaml then uncomment
	session.SetUp()

	//manager config 默认OAuth 2.0管理器
	mgr = manage.NewDefaultManager()
	mgr.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp: time.Hour * time.Duration(config.Get().OAuth2.AccessTokenExp),
		RefreshTokenExp: time.Hour*24*3,
		IsGenerateRefresh:true
	})

	//提供token的存储方式，直接存储在内存中，这里也可以选择用redis存储
	mgr.MustTokenStorage(store.NewMemoryTokenStore)
	// or use redis token store
	/*
	mgr.MapTokenStorage(oredis.NewRedisStore(&redis.Options{
		Addr: config.Get().Redis.Default.Addr,
		DB: config.Get().Redis.Default.DB,
	}))
	*/

	//access token generate method: jwt指定:token的生成方式-jwt
	mgr.MapAccessGenerate(generates.JWTAccessGenerate("", []byte(config.Get().OAuth2.JWTSignedKey), jwt.SigningMethodES512))

	// 提取Client配置
	clientStore := store.NewClientStore()
	for _, v := range config.Get().OAuth2.Client{
		clientStore.Set(v.ID, &models.Client{
			ID: v.ID,
			Secret: v.Secret,
			Domain: v.Domain,
		})
	}

	//client注册到oauth2服务器中
	mgr.MapClientStorage(clientStore)
	//config oauth2 server
	srv = server.NewServer(server.NewConfig(), mgr)

	//http server
	http.HandlerFunc("/authorize", authorizeHandler) // 该接口用于获取授权code

	//首先进入执行
	func authorizeHandler(w http.ResponseWriter, r *http.Request){
		var form url.Values
		if v, _ := session.Get(r, "RequestForm");v!= nil{
			r.ParseForm()
			if r.Form.Get("client_id") == ""{
				form = v.(url.Values)
			}
		}
		r.Form = form
		//该函数尝试从会话中删除之前保存的表单数据，以避免表单重复提交等问题。
		//如果删除过程中出现错误，则会报服务器500错误
		if err := session.Delete(w, r, "RequestForm"); err != nil{
			errorHandler(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := srv.HandleAuthorizeRequest(w, r); err != nil{
			errorHandler(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
