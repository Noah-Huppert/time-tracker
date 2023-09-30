-- Create "invoice_settings" table
CREATE TABLE "public"."invoice_settings" (
  "id" bigserial NOT NULL,
  "hourly_rate" numeric NOT NULL,
  "recipient" text NOT NULL,
  "sender" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "time_entries" table
CREATE TABLE "public"."time_entries" (
  "id" bigserial NOT NULL,
  "start_time" timestamptz NOT NULL,
  "end_time" timestamptz NOT NULL,
  "comment" text NOT NULL,
  PRIMARY KEY ("id")
);
