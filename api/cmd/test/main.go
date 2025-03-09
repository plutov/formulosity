package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/plutov/formulosity/api/pkg/db"
)

func main() {
	dsn := "postgres://user:pass@localhost:5432/formulosity"

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("cannot connect to postgres: %v", err)
	}
	defer conn.Close(ctx)

	q := db.New(conn)

	tx, _ := conn.Begin(ctx)
	defer tx.Rollback(ctx)
	txq := q.WithTx(tx)

	created, err := txq.CreateSurvey(ctx, db.CreateSurveyParams{
		Name:    fmt.Sprintf("%d", time.Now().Unix()),
		UrlSlug: fmt.Sprintf("%d", time.Now().Unix()),
	})
	if err != nil {
		log.Fatalf("cannot create survey: %v", err)
	}

	fmt.Printf("survey created: %v", created)

	survey, err := txq.GetSurvey(ctx, created.Uuid)
	if err != nil {
		log.Fatalf("cannot get survey: %v", err)
	}

	fmt.Printf("got survey: %v", survey)

	surveys, err := txq.GetSurveys(ctx)
	if err != nil {
		log.Fatalf("cannot get surveys: %v", err)
	}

	fmt.Printf("got surveys: %d", len(surveys))

	tx.Commit(ctx)
}
