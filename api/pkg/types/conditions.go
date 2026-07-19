package types

import (
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

const maxConditionDepth = 16

var conditionFields = map[string]struct{}{
	"question": {}, "equals": {}, "not_equals": {}, "contains": {},
	"greater_than": {}, "greater_than_or_equal": {}, "less_than": {},
	"less_than_or_equal": {}, "is_answered": {}, "is_not_answered": {},
	"all": {}, "any": {}, "not": {},
}

// UnmarshalYAML rejects misspelled or unsupported condition operators instead
// of silently ignoring them, as yaml's default struct decoder would.
func (c *QuestionCondition) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("condition must be a mapping")
	}
	for i := 0; i < len(value.Content); i += 2 {
		field := value.Content[i].Value
		if _, ok := conditionFields[field]; !ok {
			return fmt.Errorf("unsupported condition operator: %s", field)
		}
	}
	type rawQuestionCondition QuestionCondition
	var decoded rawQuestionCondition
	if err := value.Decode(&decoded); err != nil {
		return err
	}
	*c = QuestionCondition(decoded)
	return nil
}

// Validate checks both the shape of a condition and its supported operators.
func (c *QuestionCondition) Validate(depth int) error {
	if c == nil {
		return fmt.Errorf("condition is required")
	}
	if depth > maxConditionDepth {
		return fmt.Errorf("condition nesting exceeds maximum depth of %d", maxConditionDepth)
	}
	count := 0
	leaf := []struct {
		name  string
		value any
	}{
		{"equals", c.Equals}, {"not_equals", c.NotEquals}, {"contains", c.Contains},
		{"greater_than", c.GreaterThan}, {"greater_than_or_equal", c.GreaterThanOrEqual},
		{"less_than", c.LessThan}, {"less_than_or_equal", c.LessThanOrEqual},
	}
	for _, op := range leaf {
		if op.value != nil {
			count++
			if (op.name == "equals" || op.name == "not_equals") && !isScalarOrArray(op.value) {
				return fmt.Errorf("condition.%s must be a scalar or array", op.name)
			}
			if (op.name == "contains" || stringsNumericOperator(op.name)) && !isScalar(op.value) {
				return fmt.Errorf("condition.%s must be a scalar", op.name)
			}
			if stringsNumericOperator(op.name) && !isNumber(op.value) {
				return fmt.Errorf("condition.%s must be numeric", op.name)
			}
		}
	}
	if c.IsAnswered != nil {
		count++
	}
	if c.IsNotAnswered != nil {
		count++
	}
	if len(c.All) > 0 {
		count++
		for i := range c.All {
			if err := c.All[i].Validate(depth + 1); err != nil {
				return fmt.Errorf("condition.all[%d]: %w", i, err)
			}
		}
	} else if c.All != nil {
		return fmt.Errorf("condition.all must not be empty")
	}
	if len(c.Any) > 0 {
		count++
		for i := range c.Any {
			if err := c.Any[i].Validate(depth + 1); err != nil {
				return fmt.Errorf("condition.any[%d]: %w", i, err)
			}
		}
	} else if c.Any != nil {
		return fmt.Errorf("condition.any must not be empty")
	}
	if c.Not != nil {
		count++
		if err := c.Not.Validate(depth + 1); err != nil {
			return fmt.Errorf("condition.not: %w", err)
		}
	}
	if count != 1 {
		return fmt.Errorf("condition must contain exactly one operator")
	}
	isLogical := c.All != nil || c.Any != nil || c.Not != nil
	if isLogical && c.Question != "" {
		return fmt.Errorf("logical condition cannot specify question")
	}
	if !isLogical && c.Question == "" {
		return fmt.Errorf("condition.question is required for comparison operators")
	}
	return nil
}

func isScalar(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice && rv.Kind() != reflect.Map
}

func isScalarOrArray(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() != reflect.Map
}
func isNumber(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	}
	return false
}
func stringsNumericOperator(op string) bool {
	return op == "greater_than" || op == "greater_than_or_equal" || op == "less_than" || op == "less_than_or_equal"
}
func number(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case jsonNumber:
		f, err := strconv.ParseFloat(string(n), 64)
		return f, err == nil
	}
	return 0, false
}

// jsonNumber avoids importing encoding/json solely for a type switch while
// accepting values produced by YAML/JSON decoding through float64.
type jsonNumber string

func valuesEqual(left, right any) bool {
	if a, ok := number(left); ok {
		if b, ok := number(right); ok {
			return a == b
		}
	}
	leftValue, rightValue := reflect.ValueOf(left), reflect.ValueOf(right)
	if leftValue.IsValid() && rightValue.IsValid() &&
		(leftValue.Kind() == reflect.Array || leftValue.Kind() == reflect.Slice) &&
		(rightValue.Kind() == reflect.Array || rightValue.Kind() == reflect.Slice) {
		if leftValue.Len() != rightValue.Len() {
			return false
		}
		for i := 0; i < leftValue.Len(); i++ {
			if !valuesEqual(leftValue.Index(i).Interface(), rightValue.Index(i).Interface()) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(left, right)
}

func (c *QuestionCondition) Evaluate(answers map[string]any) (bool, error) {
	if err := c.Validate(0); err != nil {
		return false, err
	}
	return c.evaluate(answers)
}
func (c *QuestionCondition) evaluate(answers map[string]any) (bool, error) {
	if len(c.All) > 0 {
		for i := range c.All {
			ok, err := c.All[i].evaluate(answers)
			if err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	}
	if len(c.Any) > 0 {
		for i := range c.Any {
			ok, err := c.Any[i].evaluate(answers)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	if c.Not != nil {
		ok, err := c.Not.evaluate(answers)
		return !ok, err
	}
	value, answered := answers[c.Question]
	if c.IsAnswered != nil {
		return answered == *c.IsAnswered, nil
	}
	if c.IsNotAnswered != nil {
		return answered != *c.IsNotAnswered, nil
	}
	if c.Equals != nil {
		return valuesEqual(value, c.Equals), nil
	}
	if c.NotEquals != nil {
		return !valuesEqual(value, c.NotEquals), nil
	}
	if c.Contains != nil {
		rv := reflect.ValueOf(value)
		if rv.IsValid() && (rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice) {
			for i := 0; i < rv.Len(); i++ {
				if valuesEqual(rv.Index(i).Interface(), c.Contains) {
					return true, nil
				}
			}
			return false, nil
		}
		return false, nil
	}
	for _, op := range []struct {
		name      string
		threshold any
		test      func(float64, float64) bool
	}{
		{"greater_than", c.GreaterThan, func(a, b float64) bool { return a > b }}, {"greater_than_or_equal", c.GreaterThanOrEqual, func(a, b float64) bool { return a >= b }},
		{"less_than", c.LessThan, func(a, b float64) bool { return a < b }}, {"less_than_or_equal", c.LessThanOrEqual, func(a, b float64) bool { return a <= b }},
	} {
		if op.threshold != nil {
			a, ok := number(value)
			if !answered {
				return false, nil
			}
			if !ok {
				return false, nil
			}
			b, _ := number(op.threshold)
			return op.test(a, b), nil
		}
	}
	return false, nil
}
