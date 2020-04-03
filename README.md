# URL Shortener Service

# Developer Setup

## Prerequisites

The following describes how to set up an instance of the app on your
computer for development with Docker and docker-compose.

1. Install [Docker](https://docs.docker.com/engine/installation/). You need 17.04.0 or higher.

2. Install [docker-compose](https://docs.docker.com/compose/install/). You need 1.13 or higher.

## Quickstart

1. Clone the repository.

2. In the root folder of the repository rename `.env.example` to `.env`.

4. Defaults in .example file are be enough to get you going.

5. In the root folder of the repository, run:
    
        $ docker-compose build

   It will build the containers required for running the app: goserver
   
6. Run the app:
        
        $ docker-compose up -d
        
    It will run containers in detached mode (in the background).

11. Your app should be running in localhost:80 (in case docker is running in a different host, replace `localhost` by docker's host).

12. To add urls to the service make a POST request to 'localhost/api/v1/urls' with a json payload of {"url":"www.google.com"}. You will get a hash back. e.g. {"data":{"hash":"GW8OM2O","url":"http://www.google.com","created_at":"2020-04-03T19:05:43.870439+01:00"}}

13. With the hash that you get back, browse to localhost/{hash} and you'll be redirected. e.g localhost/GW8OM2O

# Testing

## Unit tests
To run unit tests:

    $ go test -v ./pkg/... 

## Integration tests
To run integration tests for Redis:

    $ go test -v ./pkg/redis/... -redis -test_redis_url=redis://localhost:6379

To run integration tests for PostgreSQL:

    $ go test -v ./pkg/postgres/... -postgres -test_postgres_url=postgres://postgresdangerouspoweruser:Password123!@localhost:5432/urlshortener
