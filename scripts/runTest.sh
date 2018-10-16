#!/usr/bin/env bash
set -oux pipefail

DATA_PATH=$1
SUBMISSION_ID=$2
DATASET_ID=$3
TIME_LIMIT=$4
SOURCE_PATH="submissions/$SUBMISSION_ID"

cp "$DATA_PATH/$DATASET_ID/in" "$SOURCE_PATH"
cp "$DATA_PATH/$DATASET_ID/out" "$SOURCE_PATH"

if [ "$5" == "java" ]; then
    filename=$6
    filename=${filename%.java}
    pushd $SOURCE_PATH
    timeout $TIME_LIMIT java $filename < in > myout
    popd
else
    docker run --rm -v $(pwd)/$SOURCE_PATH:/judge -w /judge debian /bin/sh -c "timeout $TIME_LIMIT ./$SUBMISSION_ID < in > myout"
fi
EXIT_CODE=$?

if [ "$EXIT_CODE" != "0" ]; then
    echo $EXIT_CODE
    exit 1
fi

if ! (diff "$SOURCE_PATH/out" "$SOURCE_PATH/myout" > /dev/null); then
    echo "WA"
    exit 1
fi
