package main

import (
	"flag"
	"runtime"

	"github.com/wuyq101/smsverify/config"
	"github.com/wuyq101/smsverify/service"
)

func main() {
	var conf, port string
	flag.StringVar(&conf, "conf", "./sv.conf", "Configuration file path for sms verify")
	flag.StringVar(&port, "port", "8081", "Web address server listening on")
	flag.Parse()

	config.Init(conf)
	server := service.NewService()
	server.Run(port)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
