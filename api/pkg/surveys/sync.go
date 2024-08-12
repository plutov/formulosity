package surveys

import (
	"fmt"
	"os"

	"github.com/plutov/formulosity/api/pkg/log"
	"github.com/plutov/formulosity/api/pkg/parser"
	"github.com/plutov/formulosity/api/pkg/services"
)

func SyncSurveys(svc services.Services) error {
	logCtx := log.With("func", "SyncSurveys")
	logCtx.Info("started surveys sync")

	parser := parser.NewParser()
	syncResult, err := parser.ReadSurveys(os.Getenv("SURVEYS_DIR"))
	if err != nil {
		logCtx.WithError(err).Error("unable to read surveys dir")
		return fmt.Errorf("unable to read surveys dir %w", err)
	}

	logCtx.With("surveys_count", len(syncResult.Surveys)).With("errors", len(syncResult.Errors)).Info("synced")
	logCtx.Info("persisting sync result")

	err = PersistSurveysSyncResult(svc, syncResult)
	if err != nil {
		logCtx.WithError(err).Error("unable to persist sync result")
		return fmt.Errorf("unable to persist sync result %w", err)
	}

	logCtx.Info("sync result persisted")

	return nil
}
