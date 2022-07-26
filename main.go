package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
)

const (
	endpoint = `https://www.poweralert.co.za/PowerAlertAPI/api/PowerAlertForecast/PowerAlertForecasts?callback=deleteMe`
	prefix   = `/**/ typeof deleteMe === 'function' && deleteMe(`
	suffix   = `);`
)

type APITime struct {
	time.Time
}

const apiTimeFormat = "2006-01-02T15:04:05"

func (t *APITime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t.Time = time.Time{}
		return nil
	}
	var err error
	t.Time, err = time.Parse(apiTimeFormat, s)
	return err
}

type Record struct {
	Timestamp           APITime
	ColorId             int
	DirectionId         int
	Color               string
	Direction           string
	DeclaredAvailabilty float64
	LoadForecast        float64
	MaxAvailability     float64
}

func main() {
	resp, err := http.Get(endpoint)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	text := strings.TrimPrefix(string(body), prefix)
	text = strings.TrimSuffix(text, suffix)

	records := make([]Record, 0, 1)

	err = json.Unmarshal([]byte(text), &records)
	if err != nil {
		panic(err)
	}

	if err := resp.Body.Close(); err != nil {
		panic(err)
	}

	declared := make([]float64, len(records))
	forecast := make([]float64, len(records))
	max := make([]float64, len(records))

	for i, record := range records {
		declared[i] = record.DeclaredAvailabilty
		forecast[i] = record.LoadForecast
		max[i] = record.MaxAvailability
	}
	spacer := asciigraph.White.String() + "--"
	graph := asciigraph.PlotMany(
		[][]float64{max, forecast, declared},
		asciigraph.Height(30),
		asciigraph.Width(190),
		asciigraph.SeriesColors(
			asciigraph.Red,
			asciigraph.Blue,
			asciigraph.Green,
		),
		asciigraph.Caption(asciigraph.Red.String()+"MaxAvailable"+spacer+asciigraph.Blue.String()+"Forecast"+spacer+asciigraph.Green.String()+"Declared"))

	fmt.Println(graph)

	fmt.Print("\t")
	for _, record := range records {
		fmt.Print(record.Timestamp.Format("02T15   "))
	}
	fmt.Println()
}
