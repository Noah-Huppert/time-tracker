-- Create "invoice_settings" table
CREATE TABLE "public"."invoice_settings" (
  "id" bigserial NOT NULL,
  "slot" text NOT NULL,
  "hourly_rate" numeric NOT NULL,
  "recipient" text NOT NULL,
  "sender" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "invoice_settings_slot_key" to table: "invoice_settings"
CREATE UNIQUE INDEX "invoice_settings_slot_key" ON "public"."invoice_settings" ("slot");
-- Create "time_entries" table
CREATE TABLE "public"."time_entries" (
  "id" bigserial NOT NULL,
  "start_time" timestamptz NOT NULL,
  "end_time" timestamptz NOT NULL,
  "comment" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "time_entry_identity_unique" to table: "time_entries"
CREATE UNIQUE INDEX "time_entry_identity_unique" ON "public"."time_entries" ("start_time", "end_time", "comment");
