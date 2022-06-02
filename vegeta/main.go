package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	rate := vegeta.Rate{Freq: 100, Per: time.Second}
	duration := 4 * time.Second

	form := url.Values{}
	form.Add("orderid", "b563feb7b2b84b6test")

	target := vegeta.Target{
		Method: "POST",
		URL:    "http://localhost:3333/order",
		Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		Body:   []byte(form.Encode()),
	}

	targeter := vegeta.NewStaticTargeter(target)
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}
