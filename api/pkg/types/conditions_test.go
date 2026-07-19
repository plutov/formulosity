package types

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestQuestionConditionEvaluate(t *testing.T) {
	isAnswered := true
	condition := QuestionCondition{All: []QuestionCondition{
		{Question: "employment", Equals: "yes"},
		{Question: "interests", Contains: "Go"},
		{Question: "name", IsAnswered: &isAnswered},
	}}
	answers := map[string]any{"employment": "yes", "interests": []string{"Go", "Rust"}, "name": "Ada"}
	ok, err := condition.Evaluate(answers)
	if err != nil || !ok {
		t.Fatalf("expected condition true, got %v, %v", ok, err)
	}

	answers["interests"] = []string{"Rust"}
	ok, err = condition.Evaluate(answers)
	if err != nil || ok {
		t.Fatalf("expected condition false, got %v, %v", ok, err)
	}
}

func TestSurveyConfigRejectsForwardConditionAndUnknownInterpolation(t *testing.T) {
	config := SurveyConfig{Title: "test", Questions: &Questions{Questions: []Question{
		{ID: "first", Type: QuestionType_ShortText, Label: "First"},
		{ID: "second", Type: QuestionType_ShortText, Label: "Second", VisibleIf: &QuestionCondition{Question: "third", Equals: "yes"}},
		{ID: "third", Type: QuestionType_ShortText, Label: "Third"},
	}}}
	if err := config.Validate(); err == nil {
		t.Fatal("expected forward reference error")
	}

	config.Questions.Questions[1].VisibleIf = nil
	config.Questions.Questions[0].Label = "Hello {{ missing }}"
	if err := config.Validate(); err == nil {
		t.Fatal("expected unknown interpolation error")
	}
}

func TestInterpolate(t *testing.T) {
	result, err := Interpolate("Hi {{ name }}: {{ choices }} / {{ missing }}", map[string]any{"name": "Ada", "choices": []string{"Go", "Rust"}})
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hi Ada: Go, Rust / " {
		t.Fatalf("unexpected interpolation: %q", result)
	}
}

func TestQuestionConditionValidation(t *testing.T) {
	for _, condition := range []QuestionCondition{
		{All: []QuestionCondition{}},
		{Question: "q", Equals: "yes", NotEquals: "no"},
		{Question: "q", GreaterThan: "1"},
		{Question: "q", Contains: []string{"not", "scalar"}},
	} {
		if err := condition.Validate(0); err == nil {
			t.Fatalf("expected invalid condition: %#v", condition)
		}
	}
}

func TestQuestionConditionYAMLRejectsUnsupportedOperator(t *testing.T) {
	var condition QuestionCondition
	if err := yaml.Unmarshal([]byte("question: q\nstarts_with: a\n"), &condition); err == nil {
		t.Fatal("expected unsupported operator error")
	}
}

func TestSurveyConfigInterpolatedOmitsHiddenQuestions(t *testing.T) {
	config := SurveyConfig{Intro: "Hi {{ name }}", Questions: &Questions{Questions: []Question{
		{ID: "name", Label: "Name"},
		{ID: "secret", Label: "Hi {{ name }}", Options: []string{"{{ name }}"}, VisibleIf: &QuestionCondition{Question: "name", Equals: "Ada"}},
	}}}
	rendered, err := config.Interpolated(map[string]any{"name": "Grace"})
	if err != nil {
		t.Fatal(err)
	}
	if rendered.Intro != "Hi Grace" || len(rendered.Questions.Questions) != 1 {
		t.Fatalf("unexpected interpolated config: %#v", rendered)
	}
}

func TestQuestionConditionArrayEqualityAndNumericComparison(t *testing.T) {
	condition := QuestionCondition{Question: "choices", Equals: []any{"Go", "Rust"}}
	ok, err := condition.Evaluate(map[string]any{"choices": []string{"Go", "Rust"}})
	if err != nil || !ok {
		t.Fatalf("expected arrays to compare equal, got %v, %v", ok, err)
	}

	condition = QuestionCondition{Question: "score", GreaterThanOrEqual: 3}
	ok, err = condition.Evaluate(map[string]any{"score": int64(3)})
	if err != nil || !ok {
		t.Fatalf("expected numeric comparison true, got %v, %v", ok, err)
	}
}
