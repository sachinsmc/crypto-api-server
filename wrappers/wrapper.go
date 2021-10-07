package wrappers

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/crypto-api-server/inmemorycache"
	"github.com/crypto-api-server/wsclient"
)

var SymbolsFeeCurrency = make(map[string]string, 0)
var CurrencyFullName = make(map[string]string, 0)

type Wrappers struct {
	api         *wsclient.HitBtc
	ws          *wsclient.WSClient
	websocketOn bool
	summaries   *inmemorycache.CurrencyCache
	AllSymbols  []string
}

// NewHitBtcV2Wrapper creates a generic wrapper of the HitBtc API v2.0.
func NewHitBtcV2Wrapper(publicKey string, secretKey string) *Wrappers {
	ws, _ := wsclient.NewWSClient()
	return &Wrappers{
		api:         wsclient.New(publicKey, secretKey),
		ws:          ws,
		websocketOn: false,
		summaries:   inmemorycache.NewCurrencyCache(),
	}
}

// GetTicker gets the updated ticker for a market.
func (wrapper *Wrappers) GetTicker(symbol string) (*wsclient.Ticker, error) {
	hitbtcTicker, err := wrapper.api.GetTicker(symbol)
	if err != nil {
		return nil, err
	}

	return &wsclient.Ticker{
		Last:        hitbtcTicker.Last,
		Ask:         hitbtcTicker.Ask,
		Bid:         hitbtcTicker.Bid,
		Open:        hitbtcTicker.Open,
		Low:         hitbtcTicker.Low,
		High:        hitbtcTicker.High,
		Volume:      hitbtcTicker.Volume,
		VolumeQuote: hitbtcTicker.VolumeQuote,
		Symbol:      hitbtcTicker.Symbol,
		Timestamp:   hitbtcTicker.Timestamp,
	}, nil
}

// GetMarketSummary gets the current market summary.
func (wrapper *Wrappers) GetMarketSummary(symbol string) (*wsclient.Ticker, error) {
	ret, exists := wrapper.summaries.Get(symbol)
	if !exists {
		hitbtcTicker, err := wrapper.GetTicker(symbol)
		if err != nil {
			return nil, err
		}
		feeCurrency := ""
		_, ok := SymbolsFeeCurrency[hitbtcTicker.Symbol]
		if ok {
			feeCurrency = SymbolsFeeCurrency[hitbtcTicker.Symbol]
		}
		fullName := ""
		_, ok = CurrencyFullName[feeCurrency]
		if ok {
			fullName = CurrencyFullName[feeCurrency]
		}
		ret = &wsclient.Ticker{
			Last:        hitbtcTicker.Last,
			Ask:         hitbtcTicker.Ask,
			Bid:         hitbtcTicker.Bid,
			Open:        hitbtcTicker.Open,
			Low:         hitbtcTicker.Low,
			High:        hitbtcTicker.High,
			Volume:      hitbtcTicker.Volume,
			VolumeQuote: hitbtcTicker.VolumeQuote,
			Symbol:      hitbtcTicker.Symbol,
			Timestamp:   hitbtcTicker.Timestamp,
			FeeCurrency: feeCurrency,
			FullName:    fullName,
			ID:          hitbtcTicker.Symbol,
		}
		if wrapper.Contains(supportedSymbols, hitbtcTicker.Symbol) {
			wrapper.summaries.Set(symbol, ret)
		}
		return ret, nil
	}

	return ret, nil
}

// subscribeFeeds subscribes to the Market Summary Feed service.
func (wrapper *Wrappers) subscribeFeeds(symbol string, closeChan chan bool, c chan os.Signal) error {
	handleTicker := func(wrapper *Wrappers, currencyChannel <-chan wsclient.WSNotificationTickerResponse, m string) {
		for {
			select {
			case <-closeChan:

				wrapper.Close(symbol)
				return
			default:
				hitbtcSummary, stillOpen := <-currencyChannel
				if !stillOpen {
					return
				}
				last, _ := strconv.ParseFloat(hitbtcSummary.Last, 64)
				ask, _ := strconv.ParseFloat(hitbtcSummary.Ask, 64)
				bid, _ := strconv.ParseFloat(hitbtcSummary.Bid, 64)
				open, _ := strconv.ParseFloat(hitbtcSummary.Open, 64)
				low, _ := strconv.ParseFloat(hitbtcSummary.Low, 64)
				high, _ := strconv.ParseFloat(hitbtcSummary.High, 64)
				volume, _ := strconv.ParseFloat(hitbtcSummary.Volume, 64)
				volumeQuota, _ := strconv.ParseFloat(hitbtcSummary.VolumeQuote, 64)
				feeCurrency := ""
				_, ok := SymbolsFeeCurrency[hitbtcSummary.Symbol]
				if ok {
					feeCurrency = SymbolsFeeCurrency[hitbtcSummary.Symbol]
				}
				fullName := ""
				_, ok = CurrencyFullName[feeCurrency]
				if ok {
					fullName = CurrencyFullName[feeCurrency]
				}
				sum := &wsclient.Ticker{
					Last:        last,
					Ask:         ask,
					Bid:         bid,
					Open:        open,
					Low:         low,
					High:        high,
					Volume:      volume,
					VolumeQuote: volumeQuota,
					Symbol:      hitbtcSummary.Symbol,
					FeeCurrency: feeCurrency,
					FullName:    fullName,
					ID:          hitbtcSummary.Symbol,
				}
				if wrapper.Contains(supportedSymbols, hitbtcSummary.Symbol) {
					wrapper.summaries.Set(symbol, sum)
				}

			}
		}
	}
	summaryChannel, err := wrapper.ws.SubscribeTicker(symbol)
	if err != nil {
		return err
	}

	go handleTicker(wrapper, summaryChannel, symbol)
	return nil
}

// FeedConnect connects to the feed of the exchange.
func (wrapper *Wrappers) FeedConnect() error {
	wrapper.websocketOn = true
	closeChan := make(chan bool)
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		os.Exit(0)
	}()
	for _, m := range supportedSymbols {
		err := wrapper.subscribeFeeds(m, closeChan, ch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (wrapper *Wrappers) Close(m string) {
	wrapper.ws.UnsubscribeTicker(m)
}

func (wrapper *Wrappers) CacheAllSymbols() error {
	symbolsrecords, err := wrapper.api.GetSymbols()
	if err != nil {
		return err
	}
	var symbols []string
	for _, sym := range symbolsrecords {
		SymbolsFeeCurrency[sym.Id] = sym.FeeCurrency
		symbols = append(symbols, sym.Id)
	}
	wrapper.AllSymbols = symbols
	return nil
}

func (wrapper *Wrappers) GetCurrenciesFromCache() ([]*wsclient.Ticker, error) {
	allRecords, err := wrapper.summaries.GetAll()
	if err != nil {
		return nil, err
	}
	return allRecords, nil
}

func (wrapper *Wrappers) CacheFullName() error {
	currencyRecords, err := wrapper.api.GetCurrencies()
	if err != nil {
		return err
	}
	for _, currency := range currencyRecords {
		CurrencyFullName[currency.Id] = currency.FullName
	}
	return nil
}

// Contains checks if a string is present in a slice
func (wrapper *Wrappers) Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
