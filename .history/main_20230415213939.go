package main

import (
	"fmt"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"githun.com/twilight2017/oauthsso/config"
)

var srv *server.Server
var mgr *manage.Manager

func main() {
	//config init
	config.Setup()
}
