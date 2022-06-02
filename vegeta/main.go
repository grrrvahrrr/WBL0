package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	rate := vegeta.Rate{Freq: 100, Per: time.Second}
	duration := 4 * time.Second

	target := vegeta.Target{
		Method: "POST",
		URL:    "http://localhost:3333/order",
	}

	req, err := target.Request()
	if err != nil {
		log.Fatalln(err)
	}

	form := url.Values{}
	form.Add("orderid", "b563feb7b2b84b6test")

	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(strings.NewReader(form.Encode()))

	targeter := vegeta.NewStaticTargeter(target)
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}
