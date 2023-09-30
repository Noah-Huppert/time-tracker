-- Create "csv_imports" table
CREATE TABLE "public"."csv_imports" (
  "id" bigserial NOT NULL,
  "file_name" text NOT NULL,
  "file_contents" text NOT NULL,
  "duplicate_time_entry_ids" integer[] NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "time_entries" table
CREATE TABLE "public"."time_entries" (
  "id" bigserial NOT NULL,
  "start_time" timestamptz NOT NULL,
  "end_time" timestamptz NOT NULL,
  "comment" text NOT NULL,
  "csv_import_id" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_time_entries_csv_import" FOREIGN KEY ("csv_import_id") REFERENCES "public"."csv_imports" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "time_entry_identity_unique" to table: "time_entries"
CREATE UNIQUE INDEX "time_entry_identity_unique" ON "public"."time_entries" ("start_time", "end_time", "comment");
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
-- Create "invoices" table
CREATE TABLE "public"."invoices" (
  "id" bigserial NOT NULL,
  "invoice_settings_id" bigint NOT NULL,
  "start_date" timestamptz NOT NULL,
  "end_date" timestamptz NOT NULL,
  "sent_to_client" timestamptz NULL,
  "paid_by_client" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_invoices_invoice_settings" FOREIGN KEY ("invoice_settings_id") REFERENCES "public"."invoice_settings" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "invoice_time_entries" table
CREATE TABLE "public"."invoice_time_entries" (
  "id" bigserial NOT NULL,
  "invoice_id" bigint NOT NULL,
  "time_entry_id" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_invoice_time_entries_time_entry" FOREIGN KEY ("time_entry_id") REFERENCES "public"."time_entries" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_invoices_invoice_time_entries" FOREIGN KEY ("invoice_id") REFERENCES "public"."invoices" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
