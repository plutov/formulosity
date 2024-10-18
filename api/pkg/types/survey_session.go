package types

import (
	"fmt"
	"time"
)

type SurveySessionStatus string

const (
	SurveySessionStatus_InProgress = "in_progress"
	SurveySessionStatus_Completed  = "completed"
)

type QuestionAnswer struct {
	QuestionID   string `json:"question_id"`
	QuestionUUID string `json:"question_uuid"`
	AnswerBytes  []byte `json:"answer_bytes"`
	Answer       Answer `json:"answer"`
}

type SurveySession struct {
	ID              int64               `json:"-"`
	UUID            string              `json:"uuid"`
	CreatedAt       time.Time           `json:"created_at"`
	CompletedAt     *time.Time          `json:"completed_at"`
	Status          SurveySessionStatus `json:"status"`
	SurveyUUID      string              `json:"survey_uuid"`
	IPAddr          string              `json:"ip_addr"`
	QuestionAnswers []QuestionAnswer    `json:"question_answers"`
	WebhookData     WebhookData         `json:"webhookData"`
}

type WebhookData struct {
	StatusCode int16  `json:"statusCode"`
	Response   string `json:"response"`
}

type SurveySessionsFilter struct {
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	SortBy string `query:"sort_by"`
	Order  string `query:"order"`
}

var supportedSortBy = map[string]bool{
	"uuid":         true,
	"created_at":   true,
	"completed_at": true,
	"status":       true,
}

var supportedOrder = map[string]bool{
	"asc":  true,
	"desc": true,
}

func (v *SurveySessionsFilter) Validate() error {
	if v.Limit == 0 {
		v.Limit = 100
	}

	if v.Offset < 0 {
		v.Offset = 0
	}
	if v.SortBy == "" {
		v.SortBy = "created_at"
	}
	if _, ok := supportedSortBy[v.SortBy]; !ok {
		return fmt.Errorf("sort_by is invalid: %s", v.SortBy)
	}
	if v.Order == "" {
		v.Order = "desc"
	}
	if _, ok := supportedOrder[v.Order]; !ok {
		return fmt.Errorf("order is invalid: %s", v.Order)
	}

	return nil
}

func (v *SurveySessionsFilter) ToString() string {
	return fmt.Sprintf("limit=%d_offset=%d_sort_by=%s_order=%s", v.Limit, v.Offset, v.SortBy, v.Order)
}
