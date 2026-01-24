#!/bin/bash

# This script initializes a restic backup repository.
# It expects the following environment variables to be set:
# - RESTIC_REPOSITORY: The URL of the restic repository (e.g., s3:s3.scaleway.com/my-bucket/restic)
# - AWS_ACCESS_KEY_ID: The access key for the S3-compatible storage
# - AWS_SECRET_ACCESS_KEY: The secret key for the S3-compatible storage
# - RESTIC_PASSWORD: The password for the restic repository

# Check if environment variables are set
if [ -z "$RESTIC_REPOSITORY" ]; then
  echo "Error: RESTIC_REPOSITORY environment variable is not set."
  exit 1
fi

if [ -z "$AWS_ACCESS_KEY_ID" ]; then
  echo "Error: AWS_ACCESS_KEY_ID environment variable is not set."
  exit 1
fi

if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
  echo "Error: AWS_SECRET_ACCESS_KEY environment variable is not set."
  exit 1
fi

if [ -z "$RESTIC_PASSWORD" ]; then
  echo "Error: RESTIC_PASSWORD environment variable is not set."
  exit 1
fi

echo "Initializing restic repository at $RESTIC_REPOSITORY..."

# Run restic init
restic init

if [ $? -eq 0 ]; then
  echo "Restic repository initialized successfully."
else
  echo "Error: Restic repository initialization failed."
fi

echo ""
echo "To use this script, set the required environment variables first. For example:"
echo "export RESTIC_REPOSITORY=\"s3:s3.scaleway.com/your-bucket-name/restic\""
echo "export AWS_ACCESS_KEY_ID=\"YOUR_AWS_ACCESS_KEY_ID\""
echo "export AWS_SECRET_ACCESS_KEY=\"YOUR_AWS_SECRET_ACCESS_KEY\""
echo "export RESTIC_PASSWORD=\"YOUR_RESTIC_PASSWORD\""
echo "./init_restic.sh"
