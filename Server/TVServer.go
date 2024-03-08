package Server

import (
	"TradingViewDemo/Context"
	"TradingViewDemo/Utilities"
	"encoding/json"
	"fmt"
	"net/http"

	// "os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"go.uber.org/zap"
)

var logger = Utilities.ZapLogger

type TVSignalRequest struct {
	Broker        string `json:"broker"`
	Secret        string `json:"secret"`
	Symbol        string `json:"symbol"`
	BuySell       string `json:"BuySell"`
	OrderType     string `json:"OrderType"`
	PositionSize  int    `json:"PositionSize"`
	Strategy      string `json:"Strategy"`
	EachSLPercent string `json:"eachslPercent"`
	DaySLPercent  string `json:"dayslPercent"`
	Inverse       bool   `json:"inverse"` // Added Inverse Flag
	Product       string `json:"product"`
	Exchange      string `json:"exchange"`
}

type TVServer struct {
	port    string
	context *Context.TVContext
}

// Receives a HTTP request for a place order from TradingView webhook
// Adds the received order to place order queue to be sent to the broker
func (server *TVServer) handleTVSignal(res http.ResponseWriter, req *http.Request) {
	logger.Info("New signal received!")
	// receive and add an order to context
	var reqData TVSignalRequest

	// err := json.NewDecoder(req.Body).Decode(&reqData)
	// if err != nil {
	// 	logger.Error("Failed to decode request", zap.String("ENDPOINT", "handleTVSignal"), zap.Error(err))
	// 	res.Header().Set("Content-Type", "application/json")
	// 	res.WriteHeader(http.StatusBadRequest)
	// 	err := json.NewEncoder(res).Encode("Bad request! Failed to decode the request!")
	// 	if err != nil {
	// 		logger.Error("Failed to encode response", zap.String("ENDPOINT", "handleTVSignal"), zap.Error(err))
	// 	}
	// 	return
	// }

	recvTime := time.Now().UTC()

	order := &Utilities.Order{
		Signal:        strings.TrimSpace(reqData.BuySell),
		State:         Utilities.RECEIVED,
		Symbol:        strings.TrimSpace(reqData.Symbol),
		OrderType:     strings.TrimSpace(reqData.OrderType),
		PositionSize:  reqData.PositionSize,
		Strategy:      strings.TrimSpace(reqData.Strategy),
		EachSLPercent: strings.TrimSpace(reqData.EachSLPercent),
		DaySLPercent:  strings.TrimSpace(reqData.DaySLPercent),
		Broker:        strings.TrimSpace(reqData.Broker),
		Secret:        strings.TrimSpace(reqData.Secret),
		RecvTime:      recvTime,
		Product:       strings.TrimSpace(reqData.Product),
		Exchange:      reqData.Exchange,
	}

	if reqData.Inverse { // If Inverse is true we check the position and place the order.

		positions, err := Utilities.GetPositions()
		if err != nil {
			fmt.Println("Error fetching positions:", err)
			http.Error(res, "Authentication failed. Check your API key and access token.", http.StatusForbidden)
			return
		}

		// var positionData Utilities.PositionData
		// if err := json.Unmarshal(positions, &positionData); err != nil {
		// 	fmt.Println("Error unmarshaling positions:", err)
		// 	fmt.Println("Raw JSON data during unmarshal attempt:", string(positions))
		// 	http.Error(res, "Failed to parse positions data", http.StatusInternalServerError)
		// 	return
		// }

		// var netTradingSymbols []string
		// for _, symbol := range positionData.Data.Net {
		// 	netTradingSymbols = append(netTradingSymbols, symbol.TradingSymbol)
		// }

		// var dayTradingSymbols []string
		// for _, symbol := range positionData.Data.Day {
		// 	dayTradingSymbols = append(dayTradingSymbols, symbol.TradingSymbol)
		// }

		var positionDatasy string
		var positionNetsy string

		for _, posDay := range positions.Day {
			for _, posNet := range positions.Net {
				if reqData.Symbol == posDay.Tradingsymbol {
					positionDatasy = posDay.Tradingsymbol
				}
				if reqData.Symbol == posNet.Tradingsymbol {
					positionNetsy = posNet.Tradingsymbol
				}
			}
		}

		if !Utilities.IsExistingPosition(positionDatasy, order.Symbol) && !Utilities.IsExistingPosition(positionNetsy, order.Symbol) { //For checking position is available or Not
			logger.Info("Order details",
				zap.String("Signal", order.Signal),
				zap.String("Symbol", order.Symbol),
			)
			//If the position is empty place order once
			server.context.AddOrder(order)
		} else {
			// The Inverse flag is true and the position is avaiable place order twice
			for i := 0; i < int(Utilities.EXECUTIONCOUNT); i++ {

				logger.Info("Order details",
					zap.String("Signal", order.Signal),
					zap.String("Symbol", order.Symbol),
				)
				server.context.AddOrder(order)
			}
		}
	} else {
		// If Inverse is false, execute this logic once
		logger.Info("Order details",
			zap.String("Signal", order.Signal),
			zap.String("Symbol", order.Symbol),
		)
		server.context.AddOrder(order)
	}
	var orderDetials kiteconnect.Order
	var stopLossPrice float64
	var sellStopLossPrice float64

	// select {
	// case orderId := <-Context.OrderIdChannel:
	// 	orderDetials, err = server.GetOrderDetailsUsingOrderId(orderId)
	// 	if err != nil {
	// 		return
	// 	}
	// }

	if reqData.BuySell == "buy" {
		stopLossPrice = CalculateStopLoss(Utilities.StopLoss{
			BuyPrice:        orderDetials.AveragePrice,
			StopLossPercent: reqData.EachSLPercent,
		})
	} else if reqData.BuySell == "sell" {
		sellStopLossPrice = CalculateSellStopLoss(Utilities.SellStopLoss{
			SellPrice:       orderDetials.AveragePrice,
			StopLossPercent: reqData.EachSLPercent,
		})
		stopLossPrice = sellStopLossPrice
	}

	fmt.Println(sellStopLossPrice)

	fmt.Println(orderDetials)

	getAllGTTOrder, err := Utilities.GetAllGTTOrder()
	if err != nil {
		return
	}

	for _, gttOrderCheck := range getAllGTTOrder.Data {
		if gttOrderCheck.Condition.TradingSymbol == reqData.Symbol {
			Utilities.DeleteGTTOrderById(strconv.Itoa(gttOrderCheck.ID))
		}
	}

	gttOrderResp, err := Utilities.PlaceGTTOrder(orderDetials, stopLossPrice, reqData.Product, reqData.BuySell) //Pass order details here
	if err != nil {
		return
	}

	fmt.Println(gttOrderResp)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	err = json.NewEncoder(res).Encode("Order has been added to the queue successfully!")
	if err != nil {
		logger.Error("Failed to encode response", zap.String("ENDPOINT", "handleTVSignal"))
	}
}

func (server *TVServer) Create(context *Context.TVContext, port string) *TVServer {
	server.context = context
	server.port = port
	return server
}

func (server *TVServer) runServer() error {
	fmt.Println("runServer")
	logger.Info("Server starting!", zap.String("Port", server.port))

	// create router
	router := mux.NewRouter()

	// add handlers
	router.HandleFunc("/addSignal", server.handleTVSignal).Methods("POST")

	//google auth
	http.HandleFunc("/google/login", server.GoogleLogin)
	http.HandleFunc("/google/callback", server.GoogleCallBack)

	http.ListenAndServe(":3000", nil)

	err := http.ListenAndServe(":9110", router)

	if err != nil {
		logger.Error("Error while starting the demo server!", zap.Error(err))
		return err
	} else {
		logger.Info("Demo server started!", zap.String("Port", server.port))
		return nil
	}

}

func (server *TVServer) Start() {
	err := server.runServer()
	if err != nil {
		logger.Fatal("Exiting due to demo server failure!")
		// os.Exit(1)
	}
}

func (server *TVServer) GetOrderDetailsUsingOrderId(orderId string) (kiteconnect.Order, error) {
	var orderBasedOnId kiteconnect.Order
	retrievedOrder, err := server.context.KiteConn.GetOrderHistory(orderId)
	if err != nil {
		return orderBasedOnId, err
	}
	for _, fetchOrderBasedOnId := range retrievedOrder {
		if fetchOrderBasedOnId.OrderID == orderId && fetchOrderBasedOnId.AveragePrice != 0 {
			orderBasedOnId = fetchOrderBasedOnId
		}
	}
	return orderBasedOnId, nil
}

func CalculateStopLoss(stopLoss Utilities.StopLoss) float64 {
	stopLoss.StopLossPercent = strings.TrimSuffix(stopLoss.StopLossPercent, "%")
	stopLossPercent, err := strconv.ParseFloat(stopLoss.StopLossPercent, 64)
	if err != nil {
		return 0
	}
	stopLoss.StopLossPrice = stopLoss.BuyPrice - stopLoss.BuyPrice*(stopLossPercent/100)
	return stopLoss.StopLossPrice
}

func CalculateSellStopLoss(sellStopLoss Utilities.SellStopLoss) float64 {
	sellStopLoss.StopLossPercent = strings.TrimSuffix(sellStopLoss.StopLossPercent, "%")
	stopLossPercentFloat, err := strconv.ParseFloat(sellStopLoss.StopLossPercent, 64)
	if err != nil {
		return 0
	}
	sellStopLossPrice := sellStopLoss.SellPrice + sellStopLoss.SellPrice*(stopLossPercentFloat/100)
	return sellStopLossPrice
}
