package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

func sortByIndiceCountRecord(records []*models.Record, indice []int) []*models.Record {
	for i := 0; i < len(records); i++ {
		for j := i + 1; j < len(records); j++ {
			if indice[i] < indice[j] {
				indice[i], indice[j] = indice[j], indice[i]
				records[i], records[j] = records[j], records[i]
			}
		}
	}
	return records
}

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET(
			"/api/custom/participants",
			func(c echo.Context) error {
				// declare a int variable to store the number of voters
				var voters_count int

				records, _ := app.Dao().FindRecordsByExpr("participants")
				total_voters, _ := app.Dao().FindRecordsByExpr("votes")

				indice := make([]int, len(records))

				for i := 0; i < len(records); i++ {

					p_id := records[i].GetId()
					votes, err := app.Dao().FindRecordsByExpr("votes", dbx.HashExp{"participant": p_id})

					if err != nil {
						log.Println(err)
						voters_count = 0

					} else {
						voters_count = len(votes)
					}

					// esorina aloha le description fa mavesatra
					records[i].Set("description", nil)
					records[i].SetExpand(
						map[string]interface{}{
							"voters_count":    voters_count,
							"voters_pourcent": fmt.Sprintf("%.2f ", float64(voters_count)/float64(len(total_voters))*100) + "%"},
					)

					indice[i] = voters_count
				}

				records = sortByIndiceCountRecord(records, indice)

				return c.JSON(http.StatusOK, records)
			},
			apis.ActivityLogger(app),
		)
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}
