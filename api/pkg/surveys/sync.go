package surveys

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/plutov/formulosity/api/pkg/parser"
	"github.com/plutov/formulosity/api/pkg/services"
)

func SyncSurveysOnChange(svc services.Services) {
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()

	dir := os.Getenv("SURVEYS_DIR")

	if err := watcher.Add(dir); err != nil {
		svc.Logger.Error("unable to add watcher", "err", err)
	}

	done := make(chan bool)
	go func() {
		for {
			event := <-watcher.Events
			svc.Logger.With("event", event).Info("file change event received")
			if err := SyncSurveys(svc); err != nil {
				svc.Logger.Error("unable to sync surveys on file change", "err", err)
			}
		}
	}()

	<-done
}

func SyncSurveys(svc services.Services) error {
	logCtx := svc.Logger.With("func", "SyncSurveys")
	logCtx.Info("started surveys sync")

	parser := parser.NewParser(svc)
	syncResult, err := parser.ReadSurveys(os.Getenv("SURVEYS_DIR"))
	if err != nil {
		logCtx.Error("unable to read surveys dir", "err", err)
		return fmt.Errorf("unable to read surveys dir %w", err)
	}

	logCtx.With("surveys_count", len(syncResult.Surveys)).With("errors", len(syncResult.Errors)).Info("synced")
	logCtx.Info("persisting sync result")

	err = PersistSurveysSyncResult(svc, syncResult)
	if err != nil {
		logCtx.Error("unable to persist sync result", "err", err)
		return fmt.Errorf("unable to persist sync result %w", err)
	}

	logCtx.Info("sync result persisted")

	return nil
}
