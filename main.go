package main

import (
	"metrics-crawler/crawlers"
	"os"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	prometheusMiddleware "github.com/iris-contrib/middleware/prometheus"
	"github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	totalMinutesUsed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "github_actions_total_minutes_used",
		Help: "Total GitHub Actions minutes used this period",
	})
)

var (
	totalPaidMinutesUsed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "github_actions_total_paid_minutes_used",
		Help: "Total paid GitHub Actions minutes used this period",
	})
)
var (
	totalInclusiveMinutes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "github_actions_total_inclusive_minutes",
		Help: "Total inclusive GitHub Actions minutes this period",
	})
)

func getData(dataCh chan<- *crawlers.Billing_response) {
	data, _ := crawlers.GetGithubActionsUsage()
	dataCh <- data
}

func main() {

	//Prometheus
	dataCh := make(chan *crawlers.Billing_response)
	done := make(chan bool)

	//data processing goroutine
	go func() {
		for data := range dataCh {
			//process data
			totalInclusiveMinutes.Set(float64(data.IncludedMinutes))
			totalMinutesUsed.Set(float64(data.TotalMinutes))
			totalPaidMinutesUsed.Set(float64(data.TotalPaidMinutes))
		}
		done <- true
	}()

	polling_interval := os.Getenv("POLLING_INTERVAL_MINUTES")
	interval, _ := strconv.Atoi(polling_interval)

	s := gocron.NewScheduler(time.UTC)
	_, _ = s.Every(time.Duration(interval)*time.Minute).Do(getData, dataCh)

	s.StartAsync()

	app := iris.New()
	m := prometheusMiddleware.New("githubActionsImporter", 300, 1200, 5000)

	app.Use(m.ServeHTTP)

	prometheus.MustRegister(totalInclusiveMinutes)
	prometheus.MustRegister(totalMinutesUsed)
	prometheus.MustRegister(totalPaidMinutesUsed)

	app.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
		m.ServeHTTP(ctx)
		ctx.Writef("Not Found")
	})
	app.Get("/metrics", iris.FromStd(promhttp.Handler()))

	app.Run(iris.Addr(":8080"))

}
