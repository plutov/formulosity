package types

import (
	"bytes"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
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
	Title string `json:"title" yaml:"title"`
	Intro string `json:"intro" yaml:"intro"`
	Outro string `json:"outro" yaml:"outro"`
	Theme string `json:"theme" yaml:"theme"`

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

	return nil
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
	gob.NewEncoder(&b).Encode(*s)

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
