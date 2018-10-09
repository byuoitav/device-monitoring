#!/usr/bin/env bash

echo "branch " ${CIRCLE_BRANCH:=$(git rev-parse --abbrev-ref HEAD)}

if [ "master" == "$CIRCLE_BRANCH" ]; then
    echo "setting db vars to DEV"
    export DB_ADDRESS=$DB_ADDRESS_DEV
    export DB_USERNAME=$DB_USERNAME_DEV
    export DB_PASSWORD=$DB_PASSWORD_DEV
elif [ "testing" == "$CIRCLE_BRANCH" ]; then
    echo "setting db vars to STG"
    export DB_ADDRESS=$DB_ADDRESS_STG
    export DB_USERNAME=$DB_USERNAME_STG
    export DB_PASSWORD=$DB_PASSWORD_STG
elif [ "stage" == "$CIRCLE_BRANCH" ]; then
    echo "setting db vars to STG"
    export DB_ADDRESS=$DB_ADDRESS_STG
    export DB_USERNAME=$DB_USERNAME_STG
    export DB_PASSWORD=$DB_PASSWORD_STG
elif [ "production" == "$CIRCLE_BRANCH" ]; then
    echo "setting db vars to PRD"
    export DB_ADDRESS=$DB_ADDRESS
    export DB_USERNAME=$DB_USERNAME
    export DB_PASSWORD=$DB_PASSWORD
fi

[ -z "$DB_ADDRESS" ] && echo "DB_ADDRESS not set" && exit 1
[ -z "$DB_USERNAME" ] && echo "DB_USERNAME not set" && exit 1
[ -z "$DB_PASSWORD" ] && echo "DB_PASSWORD not set" && exit 1

exit 0
