package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

const limit = 3

func chooseRandomlyInResult(records []*models.Record, limit int) []*models.Record {
	selectedVotes := make([]*models.Record, 0)

	// misafidy 3 votes/contre_votes atao patiny par hasard
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < limit; i++ {
		randIndex := r.Intn(len(records))
		selectedVotes = append(selectedVotes, records[randIndex])
		// esorina izay efa vo safidy mba tsy afaka voafidy indray
		records = append(records[:randIndex], records[randIndex+1:]...)
	}

	return selectedVotes
}

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
				var contre_votes_count, total_voters int

				records, _ := app.Dao().FindRecordsByExpr("participants")
				app.DB().Select("count(*)").From("votes").Row(&total_voters)

				indice := make([]int, len(records))
				for i := 0; i < len(records); i++ {
					// esorina aloha le description fa mavesatra
					records[i].Set("description", nil)
					apis.EnrichRecord(c, app.Dao(), records[i])

					data := records[i].Expand()

					if votes, ok := data["votes(participant)"].([]*models.Record); ok {
						indice[i] = len(votes)
						// get only max 3 votes in votes_tmp variable
						if indice[i] > limit {
							votes_tmp = chooseRandomlyInResult(votes, limit)
						} else {
							votes_tmp = votes
						}
					} else {
						indice[i] = 0
						votes_tmp = []*models.Record{}
					}

					if contre_votes, ok := data["contre_votes(participant)"].([]*models.Record); ok {
						// get only max 3 votes in contre_votes variable
						contre_votes_count = len(contre_votes)
						if contre_votes_count > limit {
							contre_votes_tmp = chooseRandomlyInResult(contre_votes, limit)
						} else {
							contre_votes_tmp = contre_votes
						}
					} else {
						contre_votes_tmp = []*models.Record{}
						contre_votes_count = 0
					}

					// set all data in expand
					records[i].SetExpand(
						map[string]interface{}{
							"contre_votes_count":   contre_votes_count,
							"contre_votes_preview": contre_votes_tmp,
							"voters_count":         indice[i],
							"participant_pourcent": fmt.Sprintf("%.2f", float64(indice[i])/float64(total_voters)*100) + " %",
							"votes_preview":        votes_tmp,
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
