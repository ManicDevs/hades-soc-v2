#!/bin/bash
# Simple token getter for HADES API testing
TOKEN=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}' http://localhost:8080/api/v1/auth/login | jq -r '.data.token')
echo "$TOKEN"
