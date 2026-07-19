# Implementation plan: answer piping and conditional question visibility

## Final decisions

- Use structured YAML predicates, not free-form expressions.
- Supported boolean composition: `all`, `any`, `not`.
- Supported leaf operators: `equals`, `not_equals`, `contains`, `greater_than`, `greater_than_or_equal`, `less_than`, `less_than_or_equal`, `is_answered`, `is_not_answered`.
- Conditions reference questions by explicit `id`.
- References must point to an earlier question.
- Unknown, forward, and self-references are parser errors.
- All visible questions remain required.
- Hidden questions do not block completion, cannot receive API answers, have stored answers cleared, and are excluded from session responses, exports, and webhooks.
- The client computes navigation and visibility for responsiveness.
- Go independently evaluates visibility for enforcement and persistence.
- Piping is supported in survey `intro`, survey `outro`, question `label`, question `description`, and choice/ranking option labels.
- Interpolation syntax: `{{ question_id }}`.
- Missing values render as empty strings.
- Arrays render as comma-separated values.
- No interpolation defaults or expressions in v1.
- README documentation and examples are part of the PR.

## Phase 1: Extend the survey configuration model

### Add condition types

Likely location: `api/pkg/types/questions.go`.

Add a recursive condition model and add `VisibleIf` to `Question`:

```go
type QuestionCondition struct {
    Question           string
    Equals             any
    NotEquals          any
    Contains           any
    GreaterThan        any
    GreaterThanOrEqual any
    LessThan           any
    LessThanOrEqual    any
    IsAnswered         *bool
    IsNotAnswered      *bool

    All []QuestionCondition
    Any []QuestionCondition
    Not *QuestionCondition
}
```

The exact representation may require a custom YAML unmarshaler because condition values can be strings, booleans, numbers, or arrays.

```go
VisibleIf *QuestionCondition `json:"visible_if,omitempty" yaml:"visible_if,omitempty"`
```

### Validate condition structure

Validate that:

- Each condition node is either one valid leaf operator or one logical node.
- `all` and `any` are non-empty.
- `not` has exactly one child.
- Operators receive compatible values.
- Nesting depth is bounded.
- Unsupported operators are rejected.

### Enforce dependency order

`Validate()` currently runs before `Normalize()`, and questions without IDs receive generated IDs in `Normalize()`. For v1, referenced questions must have explicit IDs.

Validation should:

1. Build a map of question IDs to positions.
2. Validate every referenced ID exists.
3. Ensure the referenced question appears earlier in YAML.
4. Reject self-references and forward references.
5. Validate interpolation references in labels, descriptions, options, intro, and outro.

## Phase 2: Implement Go evaluation

### Normalize answers

Likely locations: `api/pkg/types/answers.go`, `api/pkg/types/questions.go`.

Convert answers to a map keyed by question ID:

```go
map[string]any
```

Use these normalized types:

- Text, email, date, and file: string.
- Rating: number.
- Yes/no: boolean.
- Single choice: string.
- Multiple choice/ranking: `[]string`.
- Missing answer: null/missing.

### Implement condition evaluation

Add a function such as:

```go
func (c *QuestionCondition) Evaluate(answers map[string]any) (bool, error)
```

Semantics:

- `equals` and `not_equals` compare scalar values or arrays.
- `contains` checks exact membership for arrays.
- Numeric operators work only with numeric answers.
- `is_answered` and `is_not_answered` check whether an answer exists.
- `all` requires every child to be true.
- `any` requires at least one child to be true.
- `not` negates its child.

### Add visibility helpers

Likely location: `api/pkg/types/survey.go`.

Add helpers such as:

```go
func (s *SurveyConfig) IsQuestionVisible(
    question Question,
    answers map[string]any,
) (bool, error)
```

and:

```go
func (s *SurveyConfig) VisibleQuestions(
    answers map[string]any,
) ([]Question, error)
```

Questions without `VisibleIf` are visible.

## Phase 3: Implement interpolation

### Add a Go interpolator

Likely location: `api/pkg/types/survey.go`.

Implement:

```go
func Interpolate(text string, answers map[string]any) (string, error)
```

Behavior:

- Match both `{{ question_id }}` and `{{question_id}}`.
- Trim whitespace around IDs.
- Missing answer becomes an empty string.
- Arrays join with `, `.
- Scalars use normal string conversion.
- Unknown question IDs are parser errors.
- No arbitrary expressions.

Keep stored `SurveyConfig` templates unchanged. Resolve interpolations on answer-aware copies when returning session-specific data.

For public survey retrieval before a session exists, unresolved values render as empty strings.

## Phase 4: Enforce visibility and clear hidden answers

### Update answer submission

Likely locations: `api/pkg/surveys/answers.go`, `api/pkg/controllers/survey_sessions.go`.

Before accepting an answer:

1. Build the current answer map.
2. Evaluate the submitted question's visibility.
3. Reject the request if the question is hidden.
4. Validate and persist the answer.
5. Recalculate visibility for all questions.
6. Delete answers for questions that are now hidden.
7. Recalculate completion.

### Add answer deletion to storage

Current storage supports upsert but not individual answer deletion. Add:

```go
DeleteSurveyQuestionAnswer(sessionUUID string, questionUUID string) error
```

Update:

- `api/pkg/storage/interface.go`
- `api/pkg/db/queries/surveys.sql`
- Generated SQLC code via `cd api && make sqlc`.
- Generated mocks via `make gen` if needed.

The operation should be idempotent.

### Clear hidden answers transitively

For an ordered dependency chain such as `A -> B -> C`, changing `A` must clear `B` and `C` if they are no longer visible.

Evaluate questions in order, determine the visible set, delete answers whose questions are not visible, and rebuild the answer map if necessary.

### Update completion logic

Replace the unconditional loop in `api/pkg/surveys/answers.go:isSessionCompleted` with logic that:

1. Builds the answer map including the newly submitted answer.
2. Iterates questions in order.
3. Skips hidden questions.
4. Requires answers for every visible question.
5. Completes only when all visible questions are answered.

## Phase 5: Update session and webhook representations

### Filter session answers

Likely location: `api/pkg/surveys/sessions.go`.

Filter `QuestionAnswers` against the current visible question set when returning a session. This protects against old sessions or partial cleanup failures.

### Interpolate session-specific content

When returning a respondent's session or survey, resolve intro/outro, labels, descriptions, and options against the current answers without mutating persisted configuration.

### Filter webhook payloads

Before calling webhooks, construct a sanitized session payload that:

- Includes only visible questions.
- Excludes hidden/stale answers.
- Resolves interpolated text for that session.

## Phase 6: Update the Vue client

### Add TypeScript condition types

Likely location: `ui/src/lib/types.ts`.

Add a TypeScript representation matching the API/YAML condition structure and add `visible_if` to `SurveyQuestion`.

### Add a client evaluator

Likely location: `ui/src/lib/questions.ts`.

Implement:

- `evaluateCondition`
- `isQuestionVisible`
- `getVisibleQuestions`
- `interpolateText`

The TypeScript evaluator must match Go semantics exactly.

After every answer:

- Recompute visible questions.
- Move away from the current question if it became hidden.
- Clear hidden answers from local session state.
- Recompute previous/next navigation using only visible questions.
- Replace local state with the API response after submission.

### Update rendering

Likely files:

- `ui/src/components/app/survey/SurveyIntro.vue`
- `ui/src/components/app/survey/SurveyQuestions.vue`
- `ui/src/components/app/survey/SurveyFooter.vue`

Apply interpolation to intro, outro, labels, descriptions, and option labels.

Keep option identity stable. Initially interpolate existing string options and submit the rendered string. A future separate option-value/option-label schema is out of scope unless required during implementation.

## Phase 7: Tests

### Go parser and validation tests

Add tests for:

- Simple and nested conditions.
- Unknown question references.
- Forward and self references.
- Invalid operator structures.
- Empty `all`/`any`.
- Invalid interpolation references.
- Interpolation in options.
- Explicit IDs required for references.

### Go evaluator tests

Cover:

- Scalar, boolean, and numeric comparisons.
- Multiple-choice exact membership with `contains`.
- Exact array equality.
- `is_answered` and `is_not_answered`.
- Nested logical operators.
- Missing answers.
- Invalid type/operator combinations.

### Go completion and persistence tests

Cover:

- All visible questions answered means completed.
- Hidden unanswered questions do not block completion.
- Newly visible questions make a session incomplete.
- Hidden answers are deleted.
- Descendant answers are deleted transitively.
- Direct submission to hidden questions is rejected.

### Vue tests

If introducing UI test infrastructure is straightforward, test:

- Condition evaluation.
- Navigation skipping hidden questions.
- Clearing dependent local answers.
- Interpolated labels updating after answer changes.
- Array formatting.
- Missing values rendering as empty.

Otherwise keep evaluator functions pure and testable and perform manual UI validation.

## Phase 8: Documentation

Update `README.md` with a dedicated section covering:

1. `visible_if`.
2. All leaf operators.
3. `all`, `any`, and `not`.
4. Top-to-bottom dependency rules.
5. Explicit IDs for referenced questions.
6. Multiple-choice `contains` semantics.
7. Hidden-answer deletion.
8. Completion behavior.
9. Interpolation syntax.
10. Supported interpolation locations.
11. Missing values becoming empty strings.
12. Array formatting.
13. Combined piping and branching examples.
14. Current limitations.

Example:

```yaml
title: Employment survey
intro: |
  Welcome {{ name }}.

outro: |
  Thanks {{ name }}. Your interests were: {{ interests }}.

questions:
  - id: name
    type: short-text
    label: What is your name?

  - id: employment
    type: single-choice
    label: Are you employed?
    options:
      - yes
      - no

  - id: employer
    type: short-text
    label: Who employs you, {{ name }}?
    visible_if:
      question: employment
      equals: yes

  - id: interests
    type: multiple-choice
    label: Which tools do you use?
    options:
      - Go
      - Rust
      - TypeScript

  - id: go_experience
    type: long-text
    label: Tell us about your Go experience.
    visible_if:
      question: interests
      contains: Go
```

## Recommended implementation order

1. Add Go condition/interpolation data structures.
2. Add parser validation and dependency checks.
3. Add Go evaluator and unit tests.
4. Add answer deletion storage method and SQLC query.
5. Update Go submission, completion, and session filtering.
6. Add TypeScript types and evaluator.
7. Update Vue navigation and answer clearing.
8. Add interpolation rendering.
9. Add webhook/export filtering.
10. Update README and sample survey.
11. Run the full validation suite.

## Validation commands

```bash
cd api
gofmt -w .
go test ./...
go test -race ./...
make lint
```

```bash
cd ui
npm install
npm run type-check
npm run build
npm run lint
```

Manual checks:

- Surveys without conditions behave exactly as before.
- Hidden questions cannot be submitted directly.
- Changing an earlier answer removes dependent answers.
- Later questions appear only when predicates are true.
- Interpolated text updates after answer changes.
- Webhooks and response exports omit hidden answers.

## Non-goals

- No arbitrary expression language.
- No forward references.
- No optional-question support.
- No branching graph/editor UI.
- No dynamic question creation.
- No persisted answer history/audit trail.
- No interpolation defaults.
- No new answer types.
- No new option `id`/`label` schema unless required by implementation.
- No database migration beyond the SQLC delete-answer query, assuming existing answer rows can be deleted directly.
