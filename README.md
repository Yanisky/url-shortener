# URL Shortener Service

A URL shortener service similar to bit.ly

Table of contents
=================
   * [Developer Setup](#developer-setup)
      * [Prerequisites](#prerequisites)
      * [Installation](#installation)
   * [Usage](#usage)
      * [API](#api)
        * [Create URLs](#create-urls)
        * [Get Usage stats](#get-usage-stats)
      * [Redirect](#redirect)
   * [Testing](#testing)
   	  * [Unit tests](#unit-tests)
   	  * [Integration tests](#integration-tests)
   * [Remarks](#remarks)

# Developer Setup

## Prerequisites

The following describes how to set up an instance of the app on your
computer for development with Docker and docker-compose.

1. Install [Docker](https://docs.docker.com/engine/installation/). You need 17.04.0 or higher.

2. Install [docker-compose](https://docs.docker.com/compose/install/). You need 1.13 or higher.

## Installation

1. Clone the repository.

2. In the root folder of the repository rename `.env.example` to `.env`.

4. Defaults in .example file are enough to get you going.

5. In the root folder of the repository, run:
    
        $ docker-compose build

   It will build the containers required for running the app: goserver
   
6. Run the app:
        
        $ docker-compose up -d
        
    It will run containers in detached mode (in the background).

11. Your app should be running in localhost:80 (in case docker is running in a different host, replace `localhost` by docker's host).


# Usage

In order to use the redirect feature you need to create some URLs first through the api.

## API

If your docker host is not running under localhost you should replace all instances of localhost by your docker's host IP or name.

### Create URLs

	
```
http POST: "localhost/api/v1/urls"
payload: {"url":"https://www.example.com"}
```
    
Curl example:

```
$ curl --header "Content-Type: application/json" --request POST --data '{"url":"https://www.example.com"}' http://localhost/api/v1/urls 
```

Example response:
```json
{
    "data": {
        "hash": "wedgpzL",
        "url": "https://www.example.com",
        "created_at": "2020-04-03T20:48:48.302946Z"
    }
}
```
The "hash" value (wedgpzL) is what you will use to get a redirect or to get stats.

### Get Usage Stats
```
http GET: "localhost/api/v1/urls/{hash}/views"
```
    
Curl example, replace {hash} with a valid hash:

```
$ curl localhost/api/v1/urls/{hash}/views
```

Example response:
```json
{
    "data": {
        "url": {
            "hash": "wedgpzL",
            "url": "https://www.example.com",
            "created_at": "2020-04-03T20:48:48Z"
        },
        "views": {
            "past_day_count": 0,
            "past_week_count": 0,
            "count": 0
        }
    }
}
```


## Redirect

To be redirect you will need a valid hash.

```
http GET: "localhost/{hash}"
```

Using our example hash if you browse to `localhost/wedgpzL` you will get redirect to `https://www.example.com`



# Testing

## Unit tests
To run unit tests:

    $ go test -v ./pkg/... 

## Integration tests
To run integration tests for Redis:

    $ go test -v ./pkg/redis/... -redis -test_redis_url=redis://localhost:6379

To run integration tests for PostgreSQL:

    $ go test -v ./pkg/postgres/... -postgres -test_postgres_url=postgres://postgresdangerouspoweruser:Password123!@localhost:5432/urlshortener


# Remarks

Right now the app is running with Redis for caching and PostgreSQL for storage. There's a solution where only PostgreSQL is used under `./cmd/url-shortener/postgres/main.go`. 

With the interfaces under `./pkg/repository.go` it's possible to use the cache, store, and analytics separately, in this case PostgreSQL is used for both storage and analytics and Redis for cache but it's possible to implement an Elasticsearch repository that matches the analytics interface and pass it to the service (TODO).

Right now PostgreSQL is a single point of failure but it's possible to improve on that and implement something like [this](https://code.flickr.net/2010/02/08/ticket-servers-distributed-unique-primary-keys-on-the-cheap/). I do like IDs as they are easy to encode with Base62 to get small url "hashes".

