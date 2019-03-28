//
// export QUOTE_DSN="postgres://postgres:123@localhost/postgres?sslmode=disable"
// export QUOTE_PORT=80
//

package main

import (
	"encoding/json"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/go-ini/ini"
	"github.com/influxdata/influxdb1-client"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Point struct {
	Currency  string `json:"id"      form:"id"       `
	Value     string `json:"quote"   form:"quote"   binding:"required"`
	Timestamp string `json:"quote"   form:"quote"   binding:"required"`
}
type ApiSingle struct {
	Base  string `json:"base"`
	Date  string `json:"date"`
	Rates Rates  `json:"rates"`
}
type ApiMultiple struct {
	Base      string  `json:"base"`
	StartDate string  `json:"date"`
	EndDate   string  `json:"date"`
	Hist      History `json:"rates"`
}
type Rates map[string]float32
type History map[string]Rates

const api_url = "https://api.exchangeratesapi.io"

func main() {

	var influxdb *client.Client

	var opts struct {
		DateFrom string `short:"f" long:"from" description:"Date to start from" required:"true"`
		DateTO   string `short:"t" long:"to"   description:"Date to end at"     required:"true"`

		Args struct {
			Action string
		} `positional-args:"yes" required:"yes"`
	}

	_, err := flags.Parse(&opts)
	checkErr(err)

	cfg, err := ini.Load("config.ini")

	u, err := url.Parse(fmt.Sprintf("http://%s:%s", cfg.Section("influxdb").Key("host").String(), cfg.Section("influxdb").Key("port").String()))
	checkErr(err)

	influxdb, err = client.NewClient(client.Config{URL: *u})
	checkErr(err)

	if _, _, err := influxdb.Ping(); err != nil {
		checkErr(err)
	}

	influxdb.SetAuth(cfg.Section("influxdb").Key("user").String(), cfg.Section("influxdb").Key("pass").String())

	// create the db instance here
	influxdb.Query(client.Query{
		Command:  fmt.Sprintf("create database %s", cfg.Section("influxdb").Key("name").String()),
		Database: cfg.Section("influxdb").Key("name").String(),
	})

	// set infinite retension policy
	//influxdb.Query(client.Query{
	//	Command:  fmt.Sprintf("CREATE RETENTION POLICY rates ON %s DURATION INF REPLICATION 1", cfg.Section("influxdb").Key("name").String()),
	//	Database: cfg.Section("influxdb").Key("name").String(),
	//})

	if opts.Args.Action == "history" {
		var multiple ApiMultiple
		var points []client.Point
		response, err := http.Get(api_url + "/history?start_at=" + opts.DateFrom + "&end_at=" + opts.DateTO)
		checkErr(err)
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		json.Unmarshal(body, &multiple)
		for date, rates := range multiple.Hist {
			//spew.Dump(date)
			//spew.Dump(rates)
			time, err := time.Parse(time.RFC3339, date+"T12:00:00+01:00")
			checkErr(err)
			for currency_name, value := range rates {
				//fmt.Println(currency_name, value, date)
				points = append(points, client.Point{
					Measurement: "rate",
					Tags: map[string]string{
						"currency": currency_name,
						"base":     multiple.Base,
					},
					Fields: map[string]interface{}{
						"value": value,
					},
					Time: time,
				})
			}
		}
		bps := client.BatchPoints{
			Points:          points,
			Database:        cfg.Section("influxdb").Key("name").String(),
			RetentionPolicy: "autogen",
		}

		if _, err := influxdb.Write(bps); err != nil {
			log.Println("Insert data error:")
			checkErr(err)
		}
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
