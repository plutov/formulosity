package types

import (
	"strings"
	"testing"
)
func ptrString(s string) *string {
	return &s
}

func TestFileAnswerValidate(t *testing.T) {
	cases := []struct {
		question  Question
		answer    FileAnswer
		expectErr bool
		errMsg    string
	}{
		{
			question:  Question{Type: QuestionType_File, Validation: &QuestionValidation{}},
			answer:    FileAnswer{FileSize: 2 * 1024 * 1024},
			expectErr: true,
			errMsg:    "maxSizeBytes is required",
		},
		{
			question:  Question{Type: QuestionType_File, Validation: &QuestionValidation{MaxSizeBytes: ptrString("1 * 1024 *1024")}},
			answer:    FileAnswer{FileSize: 2 * 1024 * 1024},
			expectErr: true,
			errMsg:    "file size exceeds",
		},
		{
			question:  Question{Type: QuestionType_File, Validation: &QuestionValidation{MaxSizeBytes: ptrString("1*1024*1024")}},
			answer:    FileAnswer{FileSize: 2 * 1024 * 1024},
			expectErr: true,
			errMsg:    "file size exceeds",
		},
		{
			question:  Question{Type: QuestionType_File, Validation: &QuestionValidation{MaxSizeBytes: ptrString("1*1024*1024"), Formats: &[]string{"jpg", "png"}}},
			answer:    FileAnswer{FileFormat: "exe"},
			expectErr: true,
			errMsg:    "file format is invalid",
		},
		{
			question:  Question{Type: QuestionType_File, Validation: &QuestionValidation{Formats: &[]string{"jpg", "png"}, MaxSizeBytes: ptrString("2*1024*1024")}},
			answer:    FileAnswer{FileFormat: "jpg", FileSize: 1 * 1024 * 1024},
			expectErr: false,
		},
	}

	for _, c := range cases {
		err := c.answer.Validate(c.question)
		if c.expectErr {
			if err == nil || !strings.Contains(err.Error(), c.errMsg) {
				t.Errorf("expected error containing %s, got %v", c.errMsg, err)
			}
		} else {
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		}
	}
}
