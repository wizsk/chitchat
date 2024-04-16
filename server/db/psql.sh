#!/bin/sh

docker run --name postgres_container -e POSTGRES_PASSWORD=pass -d -p 5432:5432 postgres
