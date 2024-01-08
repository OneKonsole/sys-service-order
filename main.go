package main

import "os"

var a App

func main() {
	a.Initialize()

	a.Run("8020")

	defer a.MQChannel.Close()
	defer a.MQConnection.Close()

	os.Exit(0)
}
