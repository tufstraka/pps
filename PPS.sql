CREATE TABLE "payments" (
  "id" integer PRIMARY KEY,
  "user_id" integer,
  "amount" decimal(10,2) NOT NULL,
  "currency" varchar(3) NOT NULL DEFAULT 'USD',
  "method" varchar(50) NOT NULL,
  "status" varchar(20) DEFAULT 'PENDING',
  "created_at" timestamp DEFAULT (now()),
  "updated_at" timestamp DEFAULT (now())
);

CREATE TABLE "users" (
  "id" integer PRIMARY KEY,
  "username" varchar(50) UNIQUE NOT NULL,
  "password_hash" varchar(255) UNIQUE NOT NULL,
  "location" varchar,
  "phone" varchar(15) NOT NULL,
  "email" varchar(20) UNIQUE NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "payment_logs" (
  "id" serial PRIMARY KEY,
  "payment_id" integer,
  "status" varchar(20) NOT NULL,
  "message" text,
  "logged_at" timestamp DEFAULT (now())
);

ALTER TABLE "payments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "payment_logs" ADD FOREIGN KEY ("payment_id") REFERENCES "payments" ("id");
