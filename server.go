package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

var limit = 3

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
				var votes_tmp, contre_votes_tmp []*models.Record

				records, _ := app.Dao().FindRecordsByExpr("participants")
				total_voters, _ := app.Dao().FindRecordsByExpr("votes")

				indice := make([]int, len(records))

				for i := 0; i < len(records); i++ {

					// esorina aloha le description fa mavesatra
					records[i].Set("description", nil)

					apis.EnrichRecord(c, app.Dao(), records[i])

					data := records[i].Expand()

					// votes collection
					if votes, ok := data["votes(participant)"].([]*models.Record); ok {
						indice[i] = len(votes)
						// get only max 3 votes in votes_tmp variable
						if indice[i] > limit {
							votes_tmp = votes[:limit]
						} else {
							votes_tmp = votes
						}
					} else {
						indice[i] = 0
						votes_tmp = []*models.Record{}
					}

					// contre_votes collection
					if contre_votes, ok := data["contre_votes(participant)"].([]*models.Record); ok {
						// get only max 3 votes in contre_votes variable
						if len(contre_votes) > limit {
							contre_votes_tmp = contre_votes[:limit]
						} else {
							contre_votes_tmp = contre_votes
						}
					} else {
						contre_votes_tmp = []*models.Record{}
					}

					// set all data in expand
					records[i].SetExpand(
						map[string]interface{}{
							"contre_votes_count": len(contre_votes_tmp),
							"voters_count":       indice[i],
							"voters_pourcent":    fmt.Sprintf("%.2f ", float64(indice[i])/float64(len(total_voters))*100) + "%",
							"votes":              votes_tmp,
							"contre_votes":       contre_votes_tmp,
						},
					)
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
