package parser

import (
	"fmt"
	"os"
	"sort"

	"github.com/plutov/formulosity/api/pkg/services"
	"github.com/plutov/formulosity/api/pkg/types"
)

type Parser struct {
	svc services.Services
}

func NewParser(svc services.Services) *Parser {
	return &Parser{
		svc: svc,
	}
}

func (p *Parser) ReadSurveys(path string) (*types.SurveysSyncResult, error) {
	surveys := []*types.Survey{}
	errors := []types.SurveyParseError{}

	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}

	items, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read surveys directory: %w", err)
	}

	for _, item := range items {
		if item.IsDir() {
			surveyConfig, err := p.ReadSurvey(path + item.Name())
			if err != nil {
				errors = append(errors, types.SurveyParseError{
					Name:      item.Name(),
					Err:       err,
					ErrString: err.Error(),
				})
				continue
			}

			// surveyConfig can be empty for empty folders
			if surveyConfig != nil {
				iterConfig := *surveyConfig
				survey := &types.Survey{
					Name:   item.Name(),
					Config: &iterConfig,
				}

				surveys = append(surveys, survey)
			}
		}
	}

	// sort surveys and errors by name alphabetically
	sort.SliceStable(surveys, func(i, j int) bool {
		return surveys[i].Name < surveys[j].Name
	})
	sort.SliceStable(errors, func(i, j int) bool {
		return errors[i].Name < errors[j].Name
	})

	return &types.SurveysSyncResult{
		Surveys: surveys,
		Errors:  errors,
	}, nil
}
