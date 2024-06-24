CREATE TABLE "payments" (
  "id" serial PRIMARY KEY,
  "user_id" integer,
  "amount" decimal(10,2) NOT NULL,
  "currency" varchar(3) NOT NULL DEFAULT 'USD',
  "method" varchar(50) NOT NULL,
  "status" varchar(20) DEFAULT 'PENDING',
  "created_at" timestamp DEFAULT (now()),
  "updated_at" timestamp DEFAULT (now())
);

ALTER TABLE "payments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

