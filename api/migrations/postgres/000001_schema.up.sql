CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE common_statuses AS ENUM ('active', 'inactive');

CREATE TYPE survey_parse_statuses AS ENUM ('success', 'error', 'deleted');

CREATE TYPE survey_delivery_statuses AS ENUM ('launched', 'stopped');

CREATE TABLE surveys (
  id serial NOT NULL PRIMARY KEY,
  uuid uuid NOT NULL DEFAULT uuid_generate_v4 () UNIQUE,
  created_at timestamp without time zone default (now () at time zone 'utc'),
  parse_status survey_parse_statuses,
  delivery_status survey_delivery_statuses,
  error_log TEXT,
  name varchar(32) NOT NULL UNIQUE,
  url_slug varchar(1024) NOT NULL UNIQUE,
  config JSONB
);

CREATE TYPE surveys_sessions_status AS ENUM ('in_progress', 'completed');

CREATE TABLE surveys_sessions (
  id serial NOT NULL PRIMARY KEY,
  uuid uuid NOT NULL DEFAULT uuid_generate_v4 () UNIQUE,
  created_at timestamp without time zone default (now () at time zone 'utc'),
  completed_at timestamp without time zone,
  status surveys_sessions_status,
  survey_id integer NOT NULL,
  ip_addr varchar(512) NULL,
  CONSTRAINT fk_surveys_sessions1 FOREIGN KEY (survey_id) REFERENCES surveys (id) ON DELETE CASCADE
);

CREATE TABLE surveys_questions (
  id serial NOT NULL PRIMARY KEY,
  uuid uuid NOT NULL DEFAULT uuid_generate_v4 () UNIQUE,
  survey_id integer NOT NULL,
  question_id varchar(256) NOT NULL,
  CONSTRAINT fk_surveys_questions1 FOREIGN KEY (survey_id) REFERENCES surveys (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX surveys_questions_id ON surveys_questions (survey_id, question_id);

CREATE TABLE surveys_answers (
  id serial NOT NULL PRIMARY KEY,
  uuid uuid NOT NULL DEFAULT uuid_generate_v4 () UNIQUE,
  created_at timestamp without time zone default (now () at time zone 'utc'),
  session_id integer NOT NULL,
  question_id integer NOT NULL,
  answer JSONB,
  CONSTRAINT fk_surveys_answers1 FOREIGN KEY (session_id) REFERENCES surveys_sessions (id) ON DELETE CASCADE,
  CONSTRAINT fk_surveys_answers2 FOREIGN KEY (question_id) REFERENCES surveys_questions (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX surveys_answers_unique ON surveys_answers (session_id, question_id);

CREATE TABLE surveys_webhook_responses (
  id serial NOT NULL PRIMARY KEY,
  created_at timestamp without time zone default (now () at time zone 'utc'),
  session_id integer NOT NULL,
  response_status integer NOT NULL,
  response TEXT,
  CONSTRAINT fk_surveys_webhooks1 FOREIGN KEY (session_id) REFERENCES surveys_sessions (id) ON DELETE CASCADE
);
