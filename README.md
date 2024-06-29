# Payment Polling Service

This project implements a payment polling service integrating with Payd APIs for handling card and mobile payments. 

## Database Schema
![Database Schema](./PPS.png)

## Getting Started

### Prerequisites

- [Gitpod](https://gitpod.io/)
- PostgreSQL
- Golang
- Payd API credentials

### Open in Gitpod

You can quickly start working on this project by opening it in Gitpod:

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/tufstraka/pps)

It will automatically start a postgreSQL database in the background on port 5432, all you have to do is set up a username and password.

go to the terminal and run the following commands to set up the environment then install the necessary dependencies then skip to Step 2 under local setuo

```sh
unset PGHOSTADDR
```

```sh
go mod tidy
```


### Local Setup

#### Step 1: Set Up PostgreSQL

Ensure you have PostgreSQL installed and running on your local machine. Set up a username and password for accessing PostgreSQL.

#### Step 2: Migrate the Database

Run the following command to migrate the database. Ensure all the necessary environment variables (`PGUSER`, `PGPASSWORD`, `PGHOST`, `PGPORT`, `PGDATABASE`, and `PGSSLMODE`) are defined based on your setup:

```sh
migrate -path=db/migrations -database "postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE?sslmode=$PGSSLMODE" -verbose up
```

### Step 3: Create a .env File

Create a .env file at the root of each service with the following credentials:

```sh
DATABASE_URL=postgres://<username>:<password>@<host>:<port>/<database>?sslmode=disable
PAYD_USERNAME=<your_payd_username>
PAYD_PASSWORD=<your_payd_password>
```

Replace the variables with your actual database and Payd API credentials. You can get the Payd username and password from your Payd dashboard.

## Running the Services

To run each service, cd into the directory and run the following command

```sh
go run .
```

## Running the tests

To run the tests, after running the service, run the following command

```sh
go test
```







