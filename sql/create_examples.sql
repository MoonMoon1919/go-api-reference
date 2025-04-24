DROP TABLE IF EXISTS schemas.users CASCADE;
DROP TABLE IF EXISTS schemas.examples CASCADE;
DROP TABLE IF EXISTS schemas.auditlog;
DROP SCHEMA IF EXISTS schemas;

CREATE SCHEMA schemas;

CREATE TABLE schemas.users (
    id uuid PRIMARY KEY NOT NULL
);

CREATE TABLE schemas.examples (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    message VARCHAR(128) NOT NULL,
    uid uuid NOT NULL REFERENCES schemas.users (id) ON DELETE CASCADE
);

CREATE TABLE schemas.auditlog (
    eventname VARCHAR(48) NOT NULL,
    uid uuid NOT NULL,
    entityid uuid NOT NULL,
    timestamp numeric NOT NULL,
    event jsonb,
    PRIMARY KEY (eventname, entityid, uid, timestamp)
)
