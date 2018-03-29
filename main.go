package main

import (
	"fmt"
	"net/http"

	"github.com/1046102779/official_account/conf"
	"github.com/1046102779/official_account/models"
	_ "github.com/1046102779/official_account/routers"
	"github.com/hprose/hprose-golang/rpc"

	"github.com/astaxie/beego"
)

func startHproseService(rpcAddr string) {
	service := rpc.NewHTTPService()
	service.AddAllMethods(&models.OfficialAccountServer{})
	http.ListenAndServe(rpcAddr, service)
}

func main() {
	fmt.Println("main starting...")
	go startHproseService(conf.RpcAddr)

	beego.Run()
}
