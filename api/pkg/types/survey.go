package types

import (
	"bytes"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

const (
	Theme_Default  = "default"
	Theme_Custom   = "custom"
	DateTimeFormat = "2006-01-02 15:04:05"
)

var SupportedThemes = map[string]bool{
	Theme_Default: true,
	Theme_Custom:  true,
}

type SurveyParseError struct {
	Name      string `json:"name" yaml:"name"`
	Err       error  `json:"-" yaml:"-"`
	ErrString string `json:"error" yaml:"error"`
}

type SurveyParseStatus string

const (
	SurveyParseStatus_Success = "success"
	SurveyParseStatus_Error   = "error"
	SurveyParseStatus_Deleted = "deleted"
)

type SurveyDeliveryStatus string

const (
	SurveyDeliveryStatus_Launched = "launched"
	SurveyDeliveryStatus_Stopped  = "stopped"
)

type Survey struct {
	ID             int64                `json:"-"`
	UUID           string               `json:"uuid"`
	CreatedAt      time.Time            `json:"created_at"`
	ParseStatus    SurveyParseStatus    `json:"parse_status"`
	DeliveryStatus SurveyDeliveryStatus `json:"delivery_status"`
	ErrorLog       string               `json:"error_log"`
	Name           string               `json:"name"`
	URLSlug        string               `json:"url_slug"`
	URL            string               `json:"url"`

	Config *SurveyConfig `json:"config"`
	Stats  SurveyStats   `json:"stats"`
}

type SurveyStats struct {
	SessionsCountInProgess int `json:"sessions_count_in_progress"`
	SessionsCountCompleted int `json:"sessions_count_completed"`
	CompletionRate         int `json:"completion_rate"`
}

type SurveyConfig struct {
	Title   string         `json:"title" yaml:"title"`
	Intro   string         `json:"intro" yaml:"intro"`
	Outro   string         `json:"outro" yaml:"outro"`
	Theme   string         `json:"theme" yaml:"theme"`
	Webhook *WebhookConfig `json:"webhook" yaml:"webhook"`

	Hash      string     `json:"hash" yaml:"-"`
	Questions *Questions `json:"questions" yaml:"-"`
	Variables *Variables `json:"variables" yaml:"-"`
	Security  *Security  `json:"security" yaml:"-"`
}

type SurveysSyncResult struct {
	Surveys []*Survey          `json:"surveys"`
	Errors  []SurveyParseError `json:"errors"`
}

func (a SurveyConfig) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *SurveyConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

var interpolationPattern = regexp.MustCompile(`{{\s*([A-Za-z0-9_-]+)\s*}}`)

// Interpolate resolves question references in text. Missing answers are
// intentionally rendered as empty strings.
func Interpolate(text string, answers map[string]any) (string, error) {
	return interpolationPattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := interpolationPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return ""
		}
		return formatInterpolatedValue(answers[parts[1]])
	}), nil
}

func formatInterpolatedValue(value any) string {
	if value == nil {
		return ""
	}
	if values, ok := value.([]string); ok {
		return strings.Join(values, ", ")
	}
	if values, ok := value.([]any); ok {
		out := make([]string, len(values))
		for i, v := range values {
			out[i] = formatInterpolatedValue(v)
		}
		return strings.Join(out, ", ")
	}
	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return fmt.Sprint(value)
}

func (s *SurveyConfig) Validate() error {
	if s.Title == "" {
		return fmt.Errorf("metadata.title is required")
	}
	if s.Theme == "" {
		s.Theme = Theme_Default
	}
	if _, ok := SupportedThemes[s.Theme]; !ok {
		return fmt.Errorf("theme is invalid: %s", s.Theme)
	}

	if s.Questions == nil {
		return fmt.Errorf("questions is required")
	}
	if s.Variables != nil {
		if err := s.Variables.Validate(); err != nil {
			return err
		}
	}
	if err := s.SetOptionsFromVariables(); err != nil {
		return err
	}
	if err := s.Questions.Validate(); err != nil {
		return err
	}
	if err := s.validateReferences(); err != nil {
		return err
	}

	if s.Webhook != nil {
		if err := s.Webhook.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (s *SurveyConfig) validateReferences() error {
	ids := make(map[string]int)
	for i, q := range s.Questions.Questions {
		if q.ID != "" {
			ids[q.ID] = i
		}
	}
	validateText := func(text string) error {
		for _, match := range interpolationPattern.FindAllStringSubmatch(text, -1) {
			if _, ok := ids[match[1]]; !ok {
				return fmt.Errorf("interpolation references unknown question: %s", match[1])
			}
		}
		return nil
	}
	if err := validateText(s.Intro); err != nil {
		return err
	}
	if err := validateText(s.Outro); err != nil {
		return err
	}
	for i, q := range s.Questions.Questions {
		if err := validateText(q.Label); err != nil {
			return fmt.Errorf("questions[%d].label: %w", i, err)
		}
		if err := validateText(q.Description); err != nil {
			return fmt.Errorf("questions[%d].description: %w", i, err)
		}
		for j, option := range q.Options {
			if err := validateText(option); err != nil {
				return fmt.Errorf("questions[%d].options[%d]: %w", i, j, err)
			}
		}
		if q.VisibleIf != nil {
			if err := q.VisibleIf.Validate(0); err != nil {
				return fmt.Errorf("questions[%d].visible_if: %w", i, err)
			}
			if err := validateConditionReferences(q.VisibleIf, ids, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateConditionReferences(c *QuestionCondition, ids map[string]int, current int) error {
	if c.Question != "" {
		index, ok := ids[c.Question]
		if !ok {
			return fmt.Errorf("condition references unknown question: %s", c.Question)
		}
		if index >= current {
			return fmt.Errorf("condition question %s must reference an earlier question", c.Question)
		}
	}
	for i := range c.All {
		if err := validateConditionReferences(&c.All[i], ids, current); err != nil {
			return err
		}
	}
	for i := range c.Any {
		if err := validateConditionReferences(&c.Any[i], ids, current); err != nil {
			return err
		}
	}
	if c.Not != nil {
		return validateConditionReferences(c.Not, ids, current)
	}
	return nil
}

func (s *SurveyConfig) IsQuestionVisible(question Question, answers map[string]any) (bool, error) {
	if question.VisibleIf == nil {
		return true, nil
	}
	return question.VisibleIf.Evaluate(answers)
}

func (s *SurveyConfig) VisibleQuestions(answers map[string]any) ([]Question, error) {
	visible := make([]Question, 0, len(s.Questions.Questions))
	for _, q := range s.Questions.Questions {
		ok, err := s.IsQuestionVisible(q, answers)
		if err != nil {
			return nil, err
		}
		if ok {
			visible = append(visible, q)
		}
	}
	return visible, nil
}

// Interpolated returns an answer-aware presentation copy of the configuration.
// Persisted templates are never changed. Hidden questions are omitted, so this
// copy is safe for a respondent-specific representation such as a webhook.
func (s *SurveyConfig) Interpolated(answers map[string]any) (*SurveyConfig, error) {
	copyConfig := *s
	copyConfig.Intro, _ = Interpolate(s.Intro, answers)
	copyConfig.Outro, _ = Interpolate(s.Outro, answers)
	visible, err := s.VisibleQuestions(answers)
	if err != nil {
		return nil, err
	}
	copyConfig.Questions = &Questions{Questions: make([]Question, len(visible))}
	for i, question := range visible {
		question.Label, _ = Interpolate(question.Label, answers)
		question.Description, _ = Interpolate(question.Description, answers)
		question.Options = append([]string(nil), question.Options...)
		for optionIndex := range question.Options {
			question.Options[optionIndex], _ = Interpolate(question.Options[optionIndex], answers)
		}
		copyConfig.Questions.Questions[i] = question
	}
	return &copyConfig, nil
}

func (s *SurveyConfig) SetOptionsFromVariables() error {
	for i, q := range s.Questions.Questions {
		if q.OptionsFromVariable != nil && *q.OptionsFromVariable != "" {
			if s.Variables == nil {
				return fmt.Errorf("variable with id %s is not found", *q.OptionsFromVariable)
			}

			var variable *Variable
			for _, v := range s.Variables.Variables {
				if v.ID == *q.OptionsFromVariable && v.Type == VariableType_List {
					foundVariable := v
					variable = &foundVariable
					break
				}
			}

			if variable == nil {
				return fmt.Errorf("variable with id %s and type list is not found", *q.OptionsFromVariable)
			}

			q.Options = variable.Options

			s.Questions.Questions[i] = q
		}
	}

	return nil
}

func (s *SurveyConfig) Normalize() {
	p := bluemonday.StripTagsPolicy()
	s.Intro = p.Sanitize(s.Intro)
	s.Outro = p.Sanitize(s.Outro)

	uniqueIDs := make(map[string]bool)
	if s.Questions != nil {
		for i, q := range s.Questions.Questions {
			q.Description = p.Sanitize(q.Description)

			if q.ID == "" {
				q.ID = q.GenerateHash()
			}

			originalID := q.ID
			for i := 2; i < 1000; i++ {
				if _, ok := uniqueIDs[q.ID]; !ok {
					uniqueIDs[q.ID] = true
					break
				}

				q.ID = fmt.Sprintf("%s-%d", originalID, i)
			}

			s.Questions.Questions[i] = q
		}
	}
}

func (s *SurveyConfig) GenerateHash() {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(*s); err != nil {
		return
	}

	h := sha256.New()
	h.Write(b.Bytes())
	bs := h.Sum(nil)

	s.Hash = fmt.Sprintf("%x", bs)
}

func (s *SurveyConfig) FindQuestionByUUID(questionUUID string) (*Question, error) {
	for _, q := range s.Questions.Questions {
		if q.UUID == questionUUID {
			question := q
			return &question, nil
		}
	}

	return nil, errors.New("question not found")
}
