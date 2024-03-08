package Context

import (
	"TradingViewDemo/Utilities"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"go.uber.org/zap"
)

var logger = Utilities.ZapLogger

type TVContext struct {
	OrderQueue     chan *Utilities.Order
	ProcessedQueue chan *Utilities.Order
	CancelDuration time.Duration
	KiteConn       *kiteconnect.Client
	KiteLock       sync.Mutex
}

var OrderIdChannel = make(chan string)

// Place the order using Kite as the broker
func (context *TVContext) processOrder(order *Utilities.Order) (error, string) {
	fmt.Println("processOrderssssss")
	context.KiteLock.Lock()
	defer context.KiteLock.Unlock()

	logger.Info("Processing place order!")

	TVOrder := kiteconnect.OrderParams{
		Exchange:        strings.ToUpper(order.Exchange),
		Tradingsymbol:   strings.ToUpper(order.Symbol),
		TransactionType: strings.ToUpper(order.Signal),
		OrderType:       strings.ToUpper(order.OrderType),
		Quantity:        order.PositionSize,
		Product:         strings.ToUpper(order.Product),
		Validity:        "DAY",
	}
	// position size
	// lopping
	res, err := context.KiteConn.PlaceOrder("regular", TVOrder)

	if err != nil {
		logger.Error("Error while placing order!", zap.Error(err))
		return err, ""
	} else if res.OrderID == "" {
		logger.Error("Error while placing order (Response orderId empty)!")
		return errors.New("response orderId is empty"), ""
	}

	return nil, res.OrderID
}

func processOrder() {
	panic("unimplemented")
}

// Cancel the order using Kite as the broker
func (context *TVContext) cancelOrder(order *Utilities.Order) error {
	context.KiteLock.Lock()
	defer context.KiteLock.Unlock()

	logger.Info("Processing cancel order!")

	res, err := context.KiteConn.CancelOrder("regular", order.Id, nil)

	if err != nil {
		logger.Error("Error while cancelling order!", zap.Error(err))
		return err
	} else if res.OrderID == "" {
		logger.Error("Error while cancelling order (Response orderId empty)!")
		return errors.New("response orderId is empty")
	}
	return nil
}

// Init Initializes broker connection
func (context *TVContext) Init() {
	logger.Info("Initializing Kite connection")
	// Create a new Kite connect instance
	context.KiteConn = kiteconnect.New(os.Getenv("API_KEY"))
	context.KiteConn.SetAccessToken(os.Getenv("ACCESS_TOKEN"))
}

// Thread function to cancel the placed orders after a given time
func (context *TVContext) runCancel() {
	logger.Info("Starting order cancel thread!")
	for {
		select {
		case order := <-context.ProcessedQueue:
			// sleep until the order cancel time
			time.Sleep(order.OrderedTime.Add(context.CancelDuration).Sub(time.Now().UTC()))

			// cancel order
			err := context.cancelOrder(order)

			processedTime := time.Now().UTC()

			if err != nil {
				logger.Error("Error while cancelling order with broker!",
					zap.String("Symbol", order.Symbol),
					zap.String("Broker", order.Broker),
					zap.Any("Action", order.Signal),
				)
			} else {
				logger.Error("Order cancelled successfully!",
					zap.String("Symbol", order.Symbol),
					zap.String("Broker", order.Broker),
					zap.Any("Action", order.Signal),
					zap.Time("Cancelled at", processedTime),
				)

				order.CancelledTime = processedTime
				order.State = Utilities.CANCELLED
			}
		}
	}
}

// Context run thread. Places new orders as received
func (context *TVContext) runContext() {
	logger.Info("TVContext started!")
	for {
		select {
		case order := <-context.OrderQueue:
			err, orderId := context.processOrder(order)

			OrderIdChannel <- orderId

			processedTime := time.Now().UTC()

			if err != nil {
				logger.Error("Error while placing order with broker!",
					zap.String("Symbol", order.Symbol),
					zap.String("Broker", order.Broker),
					zap.Any("Action", order.Signal),
				)
			} else {
				logger.Error("Order placed successfully!",
					zap.String("Symbol", order.Symbol),
					zap.String("Broker", order.Broker),
					zap.Any("Action", order.Signal),
					zap.Time("Ordered at", processedTime),
				)
				order.OrderedTime = processedTime
				order.State = Utilities.ORDERED
				order.Id = orderId

				context.ProcessedQueue <- order

			}

		}
	}
}

// Create context
func (context *TVContext) Create(cancelDuration time.Duration) *TVContext {
	context.CancelDuration = cancelDuration
	context.OrderQueue = make(chan *Utilities.Order, 5)
	context.ProcessedQueue = make(chan *Utilities.Order, 5)
	return context
}

// Start context
func (context *TVContext) Start() {
	logger.Info("Starting TVContext!")
	go context.runCancel()
	context.runContext()
	logger.Info("TVOrder context exiting!")
}
func (context *TVContext) AddOrder(order *Utilities.Order) {
	logger.Info("Adding new order to orderQueue!")
	context.OrderQueue <- order
}

// CancelExistingOrders calls GetOrder from the broker and cancels any ongoing trades
func (context *TVContext) CancelExistingOrders() {
	logger.Info("Checking for existing orders to cancel!")
	var orders []kiteconnect.Order
	var err error

	orders, err = context.KiteConn.GetOrders()

	if err != nil {
		logger.Error("Error retrieving past orders!", zap.Error(err))
		return
	}

	// iterate orders and cancel OPEN orders
	if len(orders) > 0 {
		for i := 0; i < len(orders); i++ {
			order := orders[i]

			if order.Status == "OPEN" {
				odr := &Utilities.Order{
					Id:    order.OrderID,
					State: Utilities.ORDERED,
				}

				logger.Info("Cancelling existing order", zap.String("ID", odr.Id))
				err = context.cancelOrder(odr)
				if err != nil {
					logger.Error("Error while cancelling existing order!",
						zap.String("ID", odr.Id),
						zap.Error(err),
					)
				}
			}
		}
	} else {
		logger.Info("No past orders to cancel!")
	}
}
