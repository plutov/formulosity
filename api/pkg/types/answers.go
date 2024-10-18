package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const DATE_FORMAT = "2006-01-02"

type Answer interface {
	Validate(q Question) error
	Value() (driver.Value, error)
}

type SingleOptionAnswer struct {
	AnswerValue string `json:"value"`
}

func (a SingleOptionAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *SingleOptionAnswer) Validate(q Question) error {
	if len(a.AnswerValue) == 0 {
		return fmt.Errorf("invalid option selected")
	}

	optionFound := false
	for _, option := range q.Options {
		if option == a.AnswerValue {
			optionFound = true
			break
		}
	}

	if !optionFound {
		return fmt.Errorf("invalid option selected")
	}

	return nil
}

type MultiOptionsAnswer struct {
	AnswerValue []string `json:"value"`
}

func (a MultiOptionsAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *MultiOptionsAnswer) Validate(q Question) error {
	uniqueOptions := make(map[string]bool)
	for _, option := range a.AnswerValue {
		if len(option) == 0 {
			return fmt.Errorf("invalid option selected")
		}

		optionFound := false
		for _, validOption := range q.Options {
			if validOption == option {
				optionFound = true
				break
			}
		}
		if !optionFound {
			return fmt.Errorf("invalid option selected")
		}
		if _, ok := uniqueOptions[option]; ok {
			return fmt.Errorf("duplicate option selected")
		}

		uniqueOptions[option] = true
	}

	if q.Validation != nil && q.Validation.Min != nil && len(a.AnswerValue) < int(*q.Validation.Min) {
		return fmt.Errorf("select at least %d options", *q.Validation.Min)
	}
	if q.Validation != nil && q.Validation.Max != nil && len(a.AnswerValue) > int(*q.Validation.Max) {
		return fmt.Errorf("select at most %d options", *q.Validation.Max)
	}

	return nil
}

type TextAnswer struct {
	AnswerValue string `json:"value"`
}

func (a TextAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *TextAnswer) Validate(q Question) error {
	if q.Validation != nil && q.Validation.Min != nil && len(a.AnswerValue) < int(*q.Validation.Min) {
		return fmt.Errorf("please write at least %d characters", *q.Validation.Min)
	}
	if q.Validation != nil && q.Validation.Max != nil && len(a.AnswerValue) > int(*q.Validation.Max) {
		return fmt.Errorf("please write at most %d characters", *q.Validation.Max)
	}

	return nil
}

type DateAnswer struct {
	AnswerValue string `json:"value"`
}

func (a DateAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *DateAnswer) Validate(q Question) error {
	if _, err := time.Parse(DATE_FORMAT, a.AnswerValue); err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}

	return nil
}

type NumberAnswer struct {
	AnswerValue int64 `json:"value"`
}

func (a NumberAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *NumberAnswer) Validate(q Question) error {
	if q.Type == QuestionType_Rating {
		if q.Min != nil && a.AnswerValue < int64(*q.Min) {
			return fmt.Errorf("minimum: %d", *q.Min)
		}
		if q.Max != nil && a.AnswerValue > int64(*q.Max) {
			return fmt.Errorf("maximum: %d", *q.Max)
		}
	}

	return nil
}

type BoolAnswer struct {
	AnswerValue bool `json:"value"`
}

func (a BoolAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *BoolAnswer) Validate(q Question) error {
	return nil
}

type EmailAnswer struct {
	AnswerValue string `json:"value"`
}

func (a EmailAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *EmailAnswer) Validate(q Question) error {
	if err := validation.Validate(a.AnswerValue, validation.Required, is.Email); err != nil {
		return err
	}
	return nil
}

type FileAnswer struct {
	AnswerValue string `json:"value"`
	FileSize    int64
	FileFormat  string
}

func (a FileAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *FileAnswer) Validate(q Question) error {
	if q.Type != QuestionType_File && q.Validation == nil {
		return nil
	}

	if q.Validation.MaxSizeBytes != nil {
		bytes, err := GetStringMultiplication(*q.Validation.MaxSizeBytes)
		if err != nil {
			return fmt.Errorf("invalid MaxSizeBytes format: %v", err)
		}
		if a.FileSize > bytes {
			return fmt.Errorf("file size exceeds the maximum size of %s", formatBytes(bytes))
		}
	} else {
		return fmt.Errorf("questions[].validation.maxSizeBytes is required when questions[].type is file")
	}

	if q.Validation.Formats != nil {
		formatValid := false
		for _, allowedFormat := range *q.Validation.Formats {
			if a.FileFormat == allowedFormat {
				formatValid = true
				break
			}
		}
		if !formatValid {
			return fmt.Errorf("file format is invalid: %s. Allowed formats: %v", a.FileFormat, *q.Validation.Formats)
		}
	}
	return nil
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
