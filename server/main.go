//
// export QUOTE_DSN="postgres://postgres:123@localhost/postgres?sslmode=disable"
// export QUOTE_PORT=80
//

package server

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	_ "github.com/lib/pq"
	"net/http"
	"os"
)

type Quote struct {
	Id    int    `json:"id"      form:"id"       `
	Quote string `json:"quote"   form:"quote"   binding:"required"`
}

var db *sql.DB

func setupRouter() *gin.Engine {

	r := gin.Default()

	r.GET("/quotes", func(c *gin.Context) {
		var quote Quote
		var quotes []Quote
		var sqlQuery string
		var rows *sql.Rows

		var err error

		sqlQuery = `SELECT id, quote FROM quotes`

		rows, err = db.Query(sqlQuery)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			panic(err)
			return
		}

		for rows.Next() {
			err = rows.Scan(&quote.Id, &quote.Quote)
			checkErr(err)
			quotes = append(quotes, quote)
		}

		if len(quotes) > 0 {
			c.JSON(http.StatusOK, quotes)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "No quotes"})
		}
	})

	r.POST("/quote", func(c *gin.Context) {
		var quote Quote

		if err := c.ShouldBindWith(&quote, binding.Form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Query(`INSERT INTO quotes (quote) VALUES ('` + quote.Quote + `')`)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Data(http.StatusOK, "text/html", []byte(`
<html><body>Added:<br><p><i>`+quote.Quote+`</i></p><br>
<form action="quote" method="POST">Enter new quote<br>
<input type="text" style="width:400px" name="quote"> <input type="submit">
</form><br><a href="/">Home</a><br><a href="quotes">all quotes</a></body></html>`))
	})

	r.GET("/", func(c *gin.Context) {
		var quote Quote
		sqlQuery := `SELECT id, quote FROM quotes ORDER BY RANDOM() LIMIT 1`
		row := db.QueryRow(sqlQuery)
		err := row.Scan(&quote.Id, &quote.Quote)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			panic(err)
			return
		}
		c.Data(http.StatusOK, "text/html", []byte(`
<html><body><p><i>`+quote.Quote+`</i></p><br>
<form action="quote" method="POST">Enter new quote<br>
<input type="text" style="width:400px" name="quote"> <input type="submit">
</form><br><a href="quotes">all quotes</a></body></html>`))
	})

	return r
}

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("QUOTE_DSN"))
	_, err = db.Query(`SELECT 1 FROM quotes`)
	if err != nil && err.Error() == `pq: relation "quotes" does not exist` {
		_, err = db.Query(`
			CREATE TABLE quotes
			(
			    id SERIAL PRIMARY KEY,
			    quote    VARCHAR(256)
			);
            INSERT INTO quotes (quote) VALUES ('"It''s not fully shipped until it''s fast."');
            INSERT INTO quotes (quote) VALUES ('"Practicality beats purity."');
            INSERT INTO quotes (quote) VALUES ('"Avoid administrative distraction."');
            INSERT INTO quotes (quote) VALUES ('"Mind your words, they are important."');
            INSERT INTO quotes (quote) VALUES ('"Non-blocking is better than blocking."');
            INSERT INTO quotes (quote) VALUES ('"Design for failure."');
            INSERT INTO quotes (quote) VALUES ('"Half measures are as bad as nothing at all."');
            INSERT INTO quotes (quote) VALUES ('"Favor focus over features."');
            INSERT INTO quotes (quote) VALUES ('"Approachable is better than simple."');
            INSERT INTO quotes (quote) VALUES ('"Encourage flow."');
            INSERT INTO quotes (quote) VALUES ('"Anything added dilutes everything else."');
            INSERT INTO quotes (quote) VALUES ('"Speak like a human."');
            INSERT INTO quotes (quote) VALUES ('"Responsive is better than fast."');
            INSERT INTO quotes (quote) VALUES ('"Keep it logically awesome."');
		`)
	}
	checkErr(err)

	defer func() {
		err = db.Close()
		checkErr(err)
	}()

	r := setupRouter()
	err = r.Run(":" + os.Getenv("QUOTE_PORT"))
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
