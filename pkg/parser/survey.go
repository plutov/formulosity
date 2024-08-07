package parser

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/plutov/formulosity/pkg/log"
	"github.com/plutov/formulosity/pkg/types"
	yaml "gopkg.in/yaml.v3"
)

type surveyFile struct {
	Name     surveyFileType
	Required bool
	Type     string
}

type surveyFileType string

const (
	surveyFileType_Metadata  surveyFileType = "metadata.yaml"
	surveyFileType_Questions surveyFileType = "questions.yaml"
	surveyFileType_Security  surveyFileType = "security.yaml"
	surveyFileType_Variables surveyFileType = "variables.yaml"
	surveyFileType_Theme     surveyFileType = "theme.css"
)

var surveyFiles = []surveyFile{
	{
		Name:     surveyFileType_Metadata,
		Required: true,
		Type:     "yaml",
	},
	{
		Name:     surveyFileType_Questions,
		Required: true,
		Type:     "yaml",
	},
	{
		Name:     surveyFileType_Security,
		Required: true,
		Type:     "yaml",
	},
	{
		Name:     surveyFileType_Variables,
		Required: false,
		Type:     "yaml",
	},
	{
		Name:     surveyFileType_Theme,
		Required: false,
		Type:     "css",
	},
}

func (p *Parser) ReadSurvey(path string) (*types.SurveyConfig, error) {
	logCtx := log.With("path", path)
	logCtx.Info("reading survey folder")

	if path == "" {
		logCtx.Error("path is empty")
		return nil, errors.New("invalid survey folder")
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}

	items, err := os.ReadDir(path)
	if err != nil {
		logCtx.WithError(err).Error("unable to read survey dir")
		return nil, errors.New("unable to read survey folder")
	}

	emptyDir := true
	for _, surveyFile := range surveyFiles {
		for _, item := range items {
			if item.IsDir() {
				continue
			}

			if strings.ToLower(item.Name()) == string(surveyFile.Name) {
				emptyDir = false
				break
			}
		}
	}

	// skip empty dirs (no survey files)
	if emptyDir {
		logCtx.Warn("empty survey folder")
		return nil, nil
	}

	surveyConfig := &types.SurveyConfig{}
	for _, surveyFile := range surveyFiles {
		found := false
		for _, item := range items {
			if item.IsDir() {
				continue
			}

			if strings.ToLower(item.Name()) == string(surveyFile.Name) {
				found = true
				break
			}
		}

		if !found && surveyFile.Required {
			return nil, fmt.Errorf("required file '%s' not found", surveyFile.Name)
		}

		if found {
			file, err := os.ReadFile(path + string(surveyFile.Name))
			if err != nil {
				logCtx.WithError(err).With("file", surveyFile.Name).Error("unable to read survey file")
				return nil, fmt.Errorf("unable to read survey file '%s'", surveyFile.Name)
			}

			var fileParseErr error
			switch surveyFile.Name {
			case surveyFileType_Metadata:
				fileParseErr = yaml.Unmarshal(file, &surveyConfig)
			case surveyFileType_Questions:
				fileParseErr = yaml.Unmarshal(file, &surveyConfig.Questions)
			case surveyFileType_Security:
				fileParseErr = yaml.Unmarshal(file, &surveyConfig.Security)
			case surveyFileType_Variables:
				fileParseErr = yaml.Unmarshal(file, &surveyConfig.Variables)
			}
			if fileParseErr != nil {
				return nil, fmt.Errorf("unable to parse file '%s': %w", surveyFile.Name, fileParseErr)
			}
		}
	}

	surveyConfig.GenerateHash()
	if err := surveyConfig.Validate(); err != nil {
		return nil, err
	}

	surveyConfig.Normalize()

	return surveyConfig, nil
}
