<img src="https://github.com/plutov/formulosity/blob/main/ui/public/logo_wide.png" height="100px">

## Formulosity - self-hosted Surveys as Code platform.

This approach offers a number of advantages, including:

- **Version control**: Survey files can be stored in a Github repository, which makes it easy to track changes and collaborate with others.
- **Reproducibility**: Survey files can be easily shared and reproduced, making it easy to reuse surveys or create variations of existing surveys.
- **Automation**: Survey files can be used to automate the creation and deployment of surveys, which can save time and effort.

**Formulosity** uses human-readable declarative language [YAML](https://en.wikipedia.org/wiki/YAML).

## Features

- [x] API first
- [x] Survey UI: for end users (respondents)
- [x] Console UI: manage surveys
- [x] YAML survey configuration
- [x] Basic question types
- [x] Default theme
- [x] Custom themes support
- [x] Personalized options: intro, outro, etc.
- [x] Cookie/IP duplicate response protection
- [x] Admin user authentication
- [x] Continue where you left off
- [x] Advanced validation rules
- [x] Export responses in UI or via API
- [ ] Advanced question types
- [ ] Pipe answers into the following questions

## Survey Structure

Each directory in `SURVEYS_DIR` is a survey. You can configure the source of your surveys by setting different `SURVEYS_DIR` env var.

```bash
surveys/
├── survey1/
│   ├── metadata.yaml
│   ├── questions.yaml
│   ├── security.yaml
│   ├── variables.yaml
│   └── ...
└── survey2/
    ├── metadata.yaml
    ├── questions.yaml
    └── ...
```

To get started, check out the `./api/surveys` folder with multiple examples.

## Survey Files

### metadata.yaml

This file is required! The file consists of a YAML object with specific properties describing the survey.

- **title**: This is the main title displayed to users at the beginning of the survey.
- **theme**: This specifies the visual theme applied to the survey. Currently supported themes are: default.
- **intro**: This text appears as an introduction before the first question.
- **outro**: This text appears as a conclusion after the last question.

```yaml
title: Survey Title
theme: default # or custom
intro: |
  This is the introduction to the survey.
  It can be multiple lines long.
outro: |
  Thank you for taking the survey.
  Your feedback is important to us.
```

### questions.yaml

This file is required! The file consists of a list of questions, each defined as a YAML object with specific properties.

- **type**: This specifies the type of question being asked.
- **id**: This provides a unique identifier for the question. It is useful for referencing specific questions in branching logic or data analysis. IDs must be unique across all questions in the survey.
- **label**: This is the text displayed to the user as the question itself.
- **description**: This provides additional information about the question for the user, such as clarification or instructions.
- **options**: This list defines the available answer choices for question types like `single-choice`, `multiple-choice` and `ranking`.
- **optionsFromVariable**: This property references a variable defined in a separate variables.yaml file. The variable should contain a list of options to be used for the question. This allows for reusability and centralized management of option lists.
- **validation**: This property is used to define validation rules for specific question types.

```yaml
questions:
  - type: single-choice
    id: question1 # optional ID, must be unique across all questions
    label: What is the capital of Germany?
    description: You can select multiple options
    optionsFromVariable: german-city-options # defined in variables.yaml
    options:
      - Berlin
      - Munich
      - Paris
      - London
      - Hamburg
      - Cologne
    validation:
      min: 1
      max: 3
```

### security.yaml

This file is optional. The file consists of a YAML object with specific properties for survey security settings.

- **duplicateProtection**: This property defines how the platform handles duplicate responses from the same user.

```yaml
duplicateProtection: cookie # cookie | ip
```

### variables.yaml

This file is optional. The file consists of a list of variables, each defined as a YAML object with specific properties.

- **id**: This unique identifier references the variable within questions. IDs must be unique across all variables defined in the file.
- **type**: This specifies the type of data stored in the variable. Currently supported types are: list.

```yaml
variables:
  - id: german-city-options # must be unique
    type: list
    options:
      - Berlin
      - Munich
      - Hamburg
      - Cologne
```

## Question Types

### Short Text

Prompts users for a brief written answer.

```yaml
- type: short-text
  label: What is the capital of Germany?
  # set min/max characters
  validation:
    min: 10
    max: 100
```

### Long Text

Prompts users for a detailed written answer.

```yaml
- type: long-text
  label: What is the capital of Germany?
  # set min/max characters
  validation:
    min: 10
    max: 100
```

### Single Choice

Presents a question with only one correct answer from a list of options.

```yaml
- type: single-choice
  label: What is the capital of Germany?
  options:
    - Berlin
    - Munich
    - Paris
    - London
    - Hamburg
    - Cologne
```

### Multiple Choice

Presents a question where users can select multiple answers (with limitations). You can customize the minimum and maximum allowed selections in the validation section.

```yaml
- type: multiple-choice
  label: Which of the following are cities in Germany?
  description: You can select multiple options
  validation:
    min: 1
    max: 3
  options:
    - Berlin
    - Munich
    - Paris
    - London
    - Hamburg
    - Cologne
```

### Date

Asks users to enter a specific date.

```yaml
- type: date
  label: When was the Berlin Wall built?
```

### Rating

Presents a scale for users to rate something on a predefined range.

```yaml
- type: rating
  label: How much do you like Berlin?
  min: 1
  max: 5
```

### Ranking

Asks users to rank options based on a given criteria.

```yaml
- type: ranking
  label: Rank the following cities by population
  optionsFromVariable: german-city-options
```

### Yes/No

Presents a question where users can only answer "yes" or "no".

```yaml
- type: yes-no
  label: Is Berlin the capital of Germany?
```

### Email

Prompts user to enter their email

```yaml
- type: email
  label: Please enter your email.
```

### File

Prompts user to upload their file based on a given formats and maximum upload size.

```yaml
- type: file
  label: Upload a Berlin Image
  validation:
    formats:
      - .jpg
      - .png
    max_size_bytes: 5*1024*1024 # 5 MB
```

## Responses

Responses can be shown in the UI and exported as a JSON. Alternatively you can use REST API to get survey resposnes:

```bash
curl -XGET \
http://localhost:9900/app/surveys/{SURVEY_ID}/sessions?limit=100&offset=0&sort_by=created_at&order=desc
```

Where `{SURVEY_ID}` id the UUID of a given survey.

## Screenshots

<p align="center" width="100%">
	<img src="https://github.com/plutov/formulosity/blob/main/screenshots/app.png" hspace="10" height="200px">
	<img src="https://github.com/plutov/formulosity/blob/main/screenshots/survey.png" hspace="10" height="200px">
</p>

## Installation & Deployment

You can build and run both API and UI with Docker Compose:

```
docker-compose up -d --build
```

And you should be able to access the UI on [localhost:3000](http://localhost:3000) (default basic auth: `user:pass`).

You can deploy individual services to any cloud provider or self host them.

- Go backend
- React Router UI
- Postgres database

### Environment Variables

API:

- `DATABASE_URL` - Postgres connection string
- `SURVEYS_DIR` - Directory with surveys, e.g. `/root/surveys`. It's suggested to use mounted volume for this directory.
- `UPLOADS_DIR` - Directory for uploading files from the survey forms.

UI:

- `CONSOLE_API_ADDR` - Public address of the Go backend. Need to be accessible from the browser.
- `CONSOLE_API_ADDR_INTERNAL` - Internal address of the Go backend, e.g. `http://api:8080` (could be the same as `CONSOLE_API_ADDR`).
- `IRON_SESSION_SECRET` - Secret for session encryption
- `HTTP_BASIC_AUTH` - Format: `user:pass` for basic auth (optional)

### Run UI with npm

It's also possible to run UI using `npm`:

```
npm install
npm run dev
```

## Tech Stack

- Backend: Go, Postgres
- UI: Next.js, Tailwind CSS

## Create new Postgres migration

Make sure to install [go-migrate](https://github.com/golang-migrate/migrate) first.

```
cd api
migrate create -dir migrations -ext sql -seq name
```

## Run Go tests

```
cd api
make test
```

## Contributing Guidelines

Pull requests, bug reports, and all other forms of contribution are welcomed and highly encouraged!
