CREATE SEQUENCE ai_tasks_seq;

CREATE TABLE ism.ai_tasks
(
    task_id          INTEGER   DEFAULT nextval('ai_tasks_seq') NOT NULL,
    task_date        timestamp default now()                   NOT NULL,
    ai_request_id    TEXT                                      NOT NULL,
    target_table     TEXT                                      NOT NULL,
    ai_request_dates DATE[]                                    NOT NULL,
    ai_status        TEXT,
    ai_start_date    timestamp,
    ai_finish_date   timestamp,
    ai_meta          JSONB,
    CONSTRAINT ai_tasks_pk PRIMARY KEY (task_id)
);