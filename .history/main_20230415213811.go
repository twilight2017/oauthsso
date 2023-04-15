package main

import (
	"fmt"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
)

var srv *server.Server
var mgr *manage.Manager

func main() {
	fmt.Println(srv, mgr)
}
