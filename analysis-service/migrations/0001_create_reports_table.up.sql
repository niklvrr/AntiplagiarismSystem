CREATE TABLE reports
(
    task_id UUID PRIMARY KEY NOT NULL,
    is_plagiarism BOOLEAN DEFAULT FALSE,
    plagiarism_percentage float DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);