CREATE TABLE surveys_webhook_responses (
  id serial NOT NULL PRIMARY KEY,
  created_at timestamp without time zone default (now () at time zone 'utc'),
  session_id integer NOT NULL,
  response_status integer NOT NULL,
  response TEXT,
  CONSTRAINT fk_surveys_webhooks1 FOREIGN KEY (session_id) REFERENCES surveys_sessions (id) ON DELETE CASCADE
);

