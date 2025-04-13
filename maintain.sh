#!/usr/bin/bash
# This script is used to maintain the repository by updating the query

PWD=$(pwd)

cd query-generator && go run . && cd ../
protoc --go_out=. --go_opt=paths=source_relative     --go-grpc_out=. --go-grpc_opt=paths=source_relative     models/taskmanager.proto
swag i --pd --parseInternal
echo "Done"