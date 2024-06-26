CREATE TABLE "payment_logs" (
  "id" serial PRIMARY KEY,
  "payment_id" integer,
  "status" varchar(20) NOT NULL,
  "message" text,
  "logged_at" timestamp DEFAULT (now())
);

ALTER TABLE "payment_logs" ADD FOREIGN KEY ("payment_id") REFERENCES "payments" ("id");