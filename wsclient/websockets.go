package wsclient

import (
	"context"
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/juju/errors"
	jsonrpc2 "github.com/sourcegraph/jsonrpc2"
	jsonrpc2ws "github.com/sourcegraph/jsonrpc2/websocket"
)

const wsAPIURL string = "wss://api.hitbtc.com/api/2/ws"

// responseChannels handles all incoming data from the hitbtc connection.
type responseChannels struct {
	notifications notificationChannels

	ErrorFeed chan error
}

// notificationChannels contains all the notifications from hitbtc for subscribed feeds.
type notificationChannels struct {
	TickerFeed map[string]chan WSNotificationTickerResponse
}

// Handle handles all incoming connections and fills the channels properly.
func (h *responseChannels) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Params != nil {
		message := *req.Params
		switch req.Method {
		case "ticker":
			var msg WSNotificationTickerResponse
			err := json.Unmarshal(message, &msg)
			if err != nil {
				h.ErrorFeed <- err
			} else {
				h.notifications.TickerFeed[msg.Symbol] <- msg
			}
		}
	}
}

// WSClient represents a JSON RPC v2 Connection over Websocket,
type WSClient struct {
	conn    *jsonrpc2.Conn
	updates *responseChannels
}

// NewWSClient creates a new WSClient
func NewWSClient() (*WSClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsAPIURL, nil)
	if err != nil {
		return nil, err
	}

	handler := responseChannels{
		notifications: notificationChannels{
			TickerFeed: make(map[string]chan WSNotificationTickerResponse),
		},
		ErrorFeed: make(chan error),
	}

	return &WSClient{
		conn:    jsonrpc2.NewConn(context.Background(), jsonrpc2ws.NewObjectStream(conn), jsonrpc2.AsyncHandler(&handler)),
		updates: &handler,
	}, nil
}

// Close closes the Websocket connected to the hitbtc api.
func (c *WSClient) Close() {
	c.conn.Close()

	for _, channel := range c.updates.notifications.TickerFeed {
		close(channel)
	}

	close(c.updates.ErrorFeed)

	c.updates.notifications.TickerFeed = make(map[string]chan WSNotificationTickerResponse)
	c.updates.ErrorFeed = make(chan error)
}

// WSGetCurrencyRequest is get currency request type on websocket
type WSGetCurrencyRequest struct {
	Currency string `json:"currency,required"`
}

// WSGetCurrencyResponse is get currency response type on websocket
type WSGetCurrencyResponse struct {
	ID                 string `json:"id"`
	FullName           string `json:"fullname"`
	Crypto             bool   `json:"crypto"`
	PayinEnabled       bool   `json:"payinEnabled"`
	PayinPaymentID     bool   `json:"payinPaymentId"`
	PayinConfirmations int    `json:"payinConfirmations"`
	PayoutEnabled      bool   `json:"payoutEnabled"`
	PayoutIsPaymentID  bool   `json:"payoutIsPaymentId"`
	TransferEnabled    bool   `json:"transferEnabled"`
	Delisted           bool   `json:"delisted"`
	PayoutFee          string `json:"payoutFee"`
}

// GetCurrencyInfo get the info about a currency.
func (c *WSClient) GetCurrencyInfo(symbol string) (*WSGetCurrencyResponse, error) {
	var request = WSGetCurrencyRequest{Currency: symbol}
	var response WSGetCurrencyResponse

	err := c.conn.Call(context.Background(), "getCurrency", request, &response)
	if err != nil {
		return nil, errors.Annotate(err, "Hitbtc GetCurrency")
	}
	return &response, nil
}

// WSGetSymbolRequest is get symbols request type on websocket
type WSGetSymbolRequest struct {
	Symbol string `json:"symbol,required"`
}

// WSGetSymbolResponse is get symbols response type on websocket
type WSGetSymbolResponse struct {
	ID                   string `json:"id,required"`
	BaseCurrency         string `json:"baseCurrency,required"`
	QuoteCurrency        string `json:"quoteCurrency,required"`
	QuantityIncrement    string `json:"quantityIncrement,required"`
	TickSize             string `json:"tickSize,required"`
	TakeLiquidityRate    string `json:"takeLiquidityRate,required"`
	ProvideLiquidityRate string `json:"provideLiquidityRate,required"`
	FeeCurrency          string `json:"feeCurrency,required"`
}

// GetSymbol obtains the data of a market.
func (c *WSClient) GetSymbol(symbol string) (*WSGetSymbolResponse, error) {
	var request = WSGetSymbolRequest{Symbol: symbol}
	var response WSGetSymbolResponse

	err := c.conn.Call(context.Background(), "getSymbol", request, &response)
	if err != nil {
		return nil, errors.Annotate(err, "Hitbtc GetSymbol")
	}
	return &response, nil
}

// WSNotificationTickerResponse is notification response type on websocket
type WSNotificationTickerResponse struct {
	Ask         string `json:"ask,required"`         // Best ask price
	Bid         string `json:"bid,required"`         // Best bid price
	Last        string `json:"last,required"`        // Last trade price
	Open        string `json:"open,required"`        // Last trade price 24 hours ago
	Low         string `json:"low,required"`         // Lowest trade price within 24 hours
	High        string `json:"high,required"`        // Highest trade price within 24 hours
	Volume      string `json:"volume,required"`      // Total trading amount within 24 hours in base currency
	VolumeQuote string `json:"volumeQuote,required"` // Total trading amount within 24 hours in quote currency
	Timestamp   string `json:"timestamp,required"`   // Last update or refresh ticker timestamp
	Symbol      string `json:"symbol,required"`
}

// SubscribeTicker subscribes to the specified market ticker notifications.
func (c *WSClient) SubscribeTicker(symbol string) (<-chan WSNotificationTickerResponse, error) {
	err := c.subscriptionOp("subscribeTicker", symbol)
	if err != nil {
		return nil, errors.Annotate(err, "Hitbtc SubscribeTicker")
	}

	if c.updates.notifications.TickerFeed[symbol] == nil {
		c.updates.notifications.TickerFeed[symbol] = make(chan WSNotificationTickerResponse)
	}

	return c.updates.notifications.TickerFeed[symbol], nil
}

// UnsubscribeTicker subscribes to the specified market ticker notifications.
//
// This closes also the connected channel of updates.
func (c *WSClient) UnsubscribeTicker(symbol string) error {
	err := c.subscriptionOp("unsubscribeTicker", symbol)
	if err != nil {
		return errors.Annotate(err, "Hitbtc UnsubscribeTicker")
	}

	close(c.updates.notifications.TickerFeed[symbol])
	delete(c.updates.notifications.TickerFeed, symbol)

	return nil
}

// wsSubscriptionResponse is the response for a subscribe/unsubscribe requests.
type wsSubscriptionResponse bool

// WSSubscriptionRequest is request type on websocket subscription.
type WSSubscriptionRequest struct {
	Symbol string `json:"symbol,required"`
}

func (c *WSClient) subscriptionOp(op string, symbol string) error {
	if c.conn == nil {
		return errors.New("Connection is unitialized")
	}

	var request = WSSubscriptionRequest{Symbol: symbol}
	var success wsSubscriptionResponse

	err := c.conn.Call(context.Background(), op, request, &success)
	if err != nil {
		return err
	}

	if !success {
		return errors.New("Subscribe not successful")
	}

	return nil
}
