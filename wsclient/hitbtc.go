// Package hitbtc is an implementation of the HitBTC API in Golang.
package wsclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	API_BASE = "https://api.hitbtc.com/api/2" // HitBtc API endpoint
)

// New returns an instantiated HitBTC struct
func New(apiKey, apiSecret string) *HitBtc {
	client := NewClient(apiKey, apiSecret)
	return &HitBtc{client}
}

// NewWithCustomTimeout returns an instantiated HitBTC struct with custom timeout
func NewWithCustomTimeout(apiKey, apiSecret string, timeout time.Duration) *HitBtc {
	client := NewClientWithCustomTimeout(apiKey, apiSecret, timeout)
	return &HitBtc{client}
}

// handleErr gets JSON response from livecoin API en deal with error
func handleErr(r interface{}) error {
	switch v := r.(type) {
	case map[string]interface{}:
		error := r.(map[string]interface{})["error"]
		if error != nil {
			switch v := error.(type) {
			case map[string]interface{}:
				errorMessage := error.(map[string]interface{})["message"]
				return errors.New(errorMessage.(string))
			default:
				return fmt.Errorf("i don't know about type %T ", v)
			}
		}
	case []interface{}:
		return nil
	default:
		return fmt.Errorf("i don't know about type %T ", v)
	}

	return nil
}

// HitBtc represent a HitBTC client
type HitBtc struct {
	client *client
}

// SetDebug sets enable/disable http request/response dump
func (b *HitBtc) SetDebug(enable bool) {
	b.client.debug = enable
}

// GetCurrencies is used to get all supported currencies at HitBtc along with other meta data.
func (b *HitBtc) GetCurrencies() (currencies []Currency, err error) {
	r, err := b.client.do("GET", "public/currency", nil, false)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &currencies)
	return
}

// GetSymbols is used to get the open and available trading markets at HitBtc along with other meta data.
func (b *HitBtc) GetSymbols() (symbols []Symbol, err error) {
	r, err := b.client.do("GET", "public/symbol", nil, false)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &symbols)
	return
}

// GetTicker is used to get the current ticker values for a market.
func (b *HitBtc) GetTicker(market string) (ticker Ticker, err error) {
	r, err := b.client.do("GET", "public/ticker/"+strings.ToUpper(market), nil, false)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &ticker)
	return
}

// GetAllTicker is used to get the current ticker values for all markets.
func (b *HitBtc) GetAllTicker() (tickers Tickers, err error) {
	r, err := b.client.do("GET", "public/ticker", nil, false)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &tickers)
	return
}
