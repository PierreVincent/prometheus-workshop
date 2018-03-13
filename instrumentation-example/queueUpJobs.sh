#!/bin/bash

for i in {1..100}; do
    curl -X POST http://localhost:8888/jobs
done
