package parser

import (
	"testing"

	"github.com/plutov/formulosity/api/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestReadSurveys(t *testing.T) {
	p := NewParser()

	result, err := p.ReadSurveys("./../../surveys")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	surveys := result.Surveys

	resultCopy, err := p.ReadSurveys("./../../surveys")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	surveysCopy := resultCopy.Surveys

	assert.Len(t, surveys, 4)
	assert.Equal(t, "simple", surveys[3].Name)

	surveyConfig := surveys[3].Config
	surveyConfigCopy := surveysCopy[3].Config

	assert.Len(t, surveyConfig.Hash, 64)
	assert.Len(t, surveyConfig.Questions.Questions, 9)
	assert.Len(t, surveyConfig.Variables.Variables, 1)
	assert.Equal(t, "Survey Title", surveyConfig.Title)
	assert.Equal(t, types.Theme_Default, surveyConfig.Theme)
	assert.Equal(t, surveyConfig.Hash, surveyConfigCopy.Hash)

	_, err = p.ReadSurveys("../../../notfound/")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}