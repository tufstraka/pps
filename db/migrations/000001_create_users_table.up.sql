CREATE TABLE "users" (
  "id" serial PRIMARY KEY,
  "username" varchar(50) UNIQUE NOT NULL,
  "password_hash" varchar(255) UNIQUE NOT NULL,
  "location" varchar,
  "phone" varchar(15) NOT NULL,
  "email" varchar(55) UNIQUE NOT NULL,
  "created_at" timestamp DEFAULT (now())
);
