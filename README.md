# Payments REST API Exercise (Form3)
[![CircleCI](https://circleci.com/gh/liamg/form3-test-payments-api.svg?style=svg)](https://circleci.com/gh/liamg/form3-test-payments-api)

intro + link to design.pdf

## Requirements

- docker/docker-compose
- Nothing running on port 8080

## Running Tests

The tests run on a [CircleCI build](https://circleci.com/gh/liamg/form3-test-payments-api), but can also be run locally using docker-compose:

```
./run.sh tests
```

## Running the API Service

The API service can be run locally using docker-compose:

```
./run.sh api
```

You can then access the API at [http://localhost:8080](http://localhost:8080).

