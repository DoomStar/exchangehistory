//
// export QUOTE_DSN="postgres://postgres:123@localhost/postgres?sslmode=disable"
// export QUOTE_PORT=80
//

package main

import (
	"encoding/json"
	"fmt"
	"os"

	//"github.com/davecgh/go-spew/spew"
	"github.com/go-ini/ini"
	"github.com/influxdata/influxdb1-client/v2"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

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

func influxDBClient(cfg *ini.File) client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", cfg.Section("influxdb").Key("host").String(), cfg.Section("influxdb").Key("port").String()),
		Username: cfg.Section("influxdb").Key("user").String(),
		Password: cfg.Section("influxdb").Key("pass").String(),
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	return c
}

func main() {

	var opts struct {
		DateFrom string `short:"f" long:"from" description:"Date to start from"`
		DateTO   string `short:"t" long:"to"   description:"Date to end at"`

		Args struct {
			Action string
		} `positional-args:"yes"`
	}

	_, err := flags.Parse(&opts)
	checkErr(err)

	if opts.Args.Action != "history" && opts.Args.Action != "update" {
		fmt.Println("Usage: You should provide argument 'history' or 'update'")
	}

	cfg, err := ini.Load("config/config.ini")

	influxdb := influxDBClient(cfg)
	checkErr(err)

	if _, _, err := influxdb.Ping(5); err != nil {
		checkErr(err)
	}
	// create the db instance here
	influxdb.Query(client.Query{
		Command:  fmt.Sprintf("create database %s", cfg.Section("influxdb").Key("name").String()),
		Database: cfg.Section("influxdb").Key("name").String(),
	})

	if opts.Args.Action == "history" {
		if len(opts.DateFrom) < 1 || len(opts.DateTO) < 1 {
			fmt.Println("Usage: history --from 2018-01-01 --to 2019-01-03")
			os.Exit(1)
		}
		var multiple ApiMultiple
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  cfg.Section("influxdb").Key("name").String(),
			Precision: "h",
		})
		checkErr(err)

		response, err := http.Get(api_url + "/history?start_at=" + opts.DateFrom + "&end_at=" + opts.DateTO)
		checkErr(err)
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		checkErr(err)

		err = json.Unmarshal(body, &multiple)
		checkErr(err)

		for date, rates := range multiple.Hist {
			timestamp, err := time.Parse(time.RFC3339, date+"T12:00:00+01:00")
			checkErr(err)
			for currency_name, value := range rates {
				tags := map[string]string{
					"currency": currency_name,
					"base":     multiple.Base,
				}
				values := map[string]interface{}{
					"value": value,
				}
				point, err := client.NewPoint("rate", tags, values, timestamp)
				checkErr(err)
				bp.AddPoint(point)
			}
		}

		if err := influxdb.Write(bp); err != nil {
			log.Println("Insert data error:")
			checkErr(err)
		}
	}

	if opts.Args.Action == "update" {
		var single ApiSingle
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  cfg.Section("influxdb").Key("name").String(),
			Precision: "h",
		})
		checkErr(err)

		response, err := http.Get(api_url + "/latest")
		checkErr(err)
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		checkErr(err)

		err = json.Unmarshal(body, &single)
		checkErr(err)

		timestamp, err := time.Parse(time.RFC3339, single.Date+"T12:00:00+01:00")
		checkErr(err)
		for currency_name, value := range single.Rates {
			tags := map[string]string{
				"currency": currency_name,
				"base":     single.Base,
			}
			values := map[string]interface{}{
				"value": value,
			}
			point, err := client.NewPoint("rate", tags, values, timestamp)
			checkErr(err)
			bp.AddPoint(point)
		}

		if err := influxdb.Write(bp); err != nil {
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
