package main

import "os"

var a App
var appConf AppConf

func main() {
	appConf.Initialize()
	a.AppConf = &appConf
	a.Initialize()

	a.Run()

	defer a.MQChannel.Close()
	defer a.MQConnection.Close()

	os.Exit(0)
}
