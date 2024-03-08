package main

import (
	"TradingViewDemo/Context"
	"TradingViewDemo/Server"
	"TradingViewDemo/Utilities"
	"fmt"
	"os"
	"strconv"
	"time"
)

var logger = Utilities.ZapLogger

func init() {
	logger.Info("TradingView-Demo started!")
	// set the configurations
	Utilities.SetConfigurations()
}

func main() {
	fmt.Println("hiii")
	port := os.Getenv("PORT")
	cancelMins, err := strconv.ParseInt(os.Getenv("CANCEL_MINS"), 10, 32)

	if err != nil {
		logger.Warn("Error while parsing order cancel(mins). Taking 5mins as cancel time!")
		cancelMins = 5
	}

	// create context
	var tvContext *Context.TVContext
	tvContext = new(Context.TVContext)

	tvContext = tvContext.Create(time.Duration(cancelMins) * time.Minute)

	// init context broker connection
	tvContext.Init()

	// cancel existing open trades
	tvContext.CancelExistingOrders()

	// create server
	var tvServer *Server.TVServer
	tvServer = new(Server.TVServer)

	tvServer.Create(tvContext, port)

	// start server
	go tvServer.Start()

	// start context
	tvContext.Start()
	fmt.Println("end")

}
