package main

import (
	"flag"

	"github.com/golang/glog"

	"g.sugon.com/das/dcgm-dcu/pkg/dcgm"
)

func main() {
	glog.Infof("go-dcgm start ...")
	flag.Parse()
	defer glog.Flush()
	glog.Info("go-dcgm start ...")
	//初始化dcgm服务
	dcgm.Init()
	defer dcgm.ShutDown()
}
