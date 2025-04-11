#!/usr/bin/bash
# This script is used to maintain the repository by updating the query

PWD=$(pwd)
cd query-generator && go run . && cd ../
swag i --pd --parseInternal
echo "Done"