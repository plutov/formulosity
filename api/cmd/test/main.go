package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plutov/formulosity/api/pkg/db"
)

func main() {
	dsn := "postgres://test:test@localhost:5432/surveys"

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
		Name: "test",
	})
	if err != nil {
		log.Fatalf("unable to create db pool: %v", err)
	}

	fmt.Printf("survey created, id: %s", survey.Uuid)
}
