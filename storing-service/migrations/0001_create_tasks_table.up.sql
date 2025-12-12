CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    filename TEXT NOT NULL ,
    uploaded_by UUID NOT NULL ,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);