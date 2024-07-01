# Payment Polling Service

This project implements a payment polling service integrating with Payd APIs for handling card and mobile payments. 


## Getting Started

### Prerequisites

- [Gitpod](https://gitpod.io/)
- PostgreSQL
- Golang
- Payd API credentials

### Open in Gitpod

You can quickly start working on this project by opening it in Gitpod:

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/tufstraka/pps)

It will automatically start a postgreSQL database in the background on port 5432

go to the terminal and run the following commands to set up the environment then skip to Step 2 under local setup

```sh
unset PGHOSTADDR
```

```sh
go mod tidy
```


### Local Setup

#### Step 1: Set Up PostgreSQL and RabbitMQ

Ensure you have PostgreSQL and RabbitMQ installed and running on your local machine.

#### Step 2: Connect to PostgreSQL:

Open a terminal or command prompt and connect to PostgreSQL using the psql command-line tool. 

```sh
psql
```

#### Create a New User:
Once connected to the PostgreSQL shell (psql), run the following SQL command to create a new user. Replace `<new_username>` and `<new_password>` with your desired username and password:

```sh
CREATE USER <new_username> WITH PASSWORD '<new_password>';
```

#### Create test database
Run the following command to create a database that will be used when migrating

```sh
CREATE DATABASE <new_database>
```

### Grant Permissions (Optional):
If you want to grant specific permissions to the new user, you can do so using the GRANT command. For example, to grant all privileges on a specific database:

```sh
GRANT ALL PRIVILEGES ON DATABASE <database> TO <new_username>;
```

Replace `<database>` with the name of your database.

#### Exit psql:
Once you've created the user and granted necessary permissions, you can exit psql by typing:

```sh
\q
```

#### Step 3: Create a .env File

Create a .env file at the root of each service with the following credentials:

```sh
DATABASE_URL=postgres://<username>:<password>@<host>:<port>/<database>?sslmode=disable
PAYD_USERNAME=<your_payd_username>
PAYD_PASSWORD=<your_payd_password>
JWT_SECRET_KEY=<your_jwt_secret_key>
```

Replace the variables with your actual database and Payd API credentials. You can get the Payd username and password from your Payd dashboard.

#### Step 4: Migrate the Database

Run the following command to migrate the database. Ensure all the necessary environment variables (`PGUSER`, `PGPASSWORD`, `PGHOST`, `PGPORT`, `PGDATABASE`, and `PGSSLMODE`) are defined based on your setup:

```sh
migrate -path=db/migrations -database "postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE?sslmode=$PGSSLMODE" -verbose up
```


### Running the Services

To run each service, cd into the directory and run the following command

```sh
go run .
```

### Running the tests

To run the tests, after running the service, run the following command

```sh
go test
```
### API Documentation
To access the API documentation, go to the following endpoint for the payment and authentication services

`<api_url>/swagger/`

### Database Schema
![Database Schema](./PPS.png)






