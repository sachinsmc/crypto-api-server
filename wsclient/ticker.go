package wsclient

import (
	"encoding/json"
	"time"
)

type Tickers []Ticker

//Ticker represents a Ticker from hitbtc API.
type Ticker struct {
	ID          string    `json:"id"`
	FullName    string    `json:"fullname"`
	Ask         float64   `json:"ask,string"`
	Bid         float64   `json:"bid,string"`
	Last        float64   `json:"last,string"`
	Open        float64   `json:"open,string"`
	Low         float64   `json:"low,string,omitempty"`
	High        float64   `json:"high,string,omitempty"`
	Volume      float64   `json:"volume,string"`
	VolumeQuote float64   `json:"volumeQuote,string"`
	Timestamp   time.Time `json:"timestamp"`
	Symbol      string    `json:"symbol"`
	FeeCurrency string    `json:"feecurrency"`
}

func (t *Ticker) UnmarshalJSON(data []byte) error {
	var err error
	type Alias Ticker
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err = json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.Timestamp, err = time.Parse("2006-01-02T15:04:05.999Z", aux.Timestamp)
	if err != nil {
		return err
	}
	return nil
}
