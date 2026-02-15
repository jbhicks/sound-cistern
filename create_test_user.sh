#!/bin/bash
# Create a test user in PocketBase

curl -X POST "http://localhost:8090/api/collections/users/records" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123",
    "passwordConfirm": "testpassword123",
    "first_name": "Test",
    "last_name": "User",
    "role": "user"
  }'