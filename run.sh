#!/bin/bash

trap clean INT

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

function clean() {
    docker-compose -p form3paymentsapi down
    docker-compose -p form3paymentsapi kill
    docker-compose -p form3paymentsapi rm -f
}


function api() {
    docker-compose -p form3paymentsapi up --build  
    if [ $? -ne 0 ] ; then
    echo "Failed to start containers."
    clean
    exit 1
    fi
}

function tests() {
    docker-compose -p form3paymentsapi up --build -d
    if [ $? -ne 0 ] ; then
    echo "Failed to start containers."
    clean
    exit 1
    fi
    TEST_RESULT=`docker wait form3paymentsapi_test_1`
    docker logs form3paymentsapi_test_1
    if [ "$TEST_RESULT" -ne 0 ]; then
        printf "\n${RED}Tests have failed.${NC}\n\n"
    else
        printf "${GREEN}Tests have passed.${NC}\n\n"
    fi
}

case $1 in
        "api" )
           api;;
        "tests" )
           tests;;
        *)
            echo "Usage:";
            echo "./run.sh tests";
            echo "./run.sh api";
            echo
            exit 1;;
   esac

clean
