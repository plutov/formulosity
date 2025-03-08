package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plutov/formulosity/api/pkg/db"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/formulosity"

	poolConf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("unable to parse dsn: %v", err)
	}

	connPool, err := pgxpool.NewWithConfig(context.Background(), poolConf)
	if err != nil {
		log.Fatalf("unable to create db pool: %v", err)
	}

	queries := db.New(connPool)
	survey, err := queries.CreateSurvey(context.Background(), db.CreateSurveyParams{
		Name:    string(time.Now().Unix()),
		UrlSlug: string(time.Now().Unix()),
	})
	if err != nil {
		log.Fatalf("unable to create survey: %v", err)
	}

	fmt.Printf("survey created, id: %s", survey.Uuid)

	surveys, err := queries.GetSurveys(context.Background())
	if err != nil {
		log.Fatalf("unable to get surveys: %v", err)
	}

	fmt.Printf("surveys: %v", surveys)

}
