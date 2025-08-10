#!/bin/bash

# Hugo Site Deployment Script for Carlos A. Herrera
# This script builds the Hugo site and deploys it to AWS S3 via Pulumi

set -e

echo "🚀 Starting deployment for carlosaherrera.com"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if required commands exist
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

echo -e "${BLUE}📋 Checking prerequisites...${NC}"

if ! command_exists hugo; then
    echo -e "${RED}❌ Hugo is not installed. Please install Hugo first.${NC}"
    exit 1
fi

if ! command_exists pulumi; then
    echo -e "${RED}❌ Pulumi is not installed. Please install Pulumi first.${NC}"
    exit 1
fi

if ! command_exists aws; then
    echo -e "${YELLOW}⚠️  AWS CLI not found. Make sure your AWS credentials are configured.${NC}"
fi

# Build Hugo site
echo -e "${BLUE}🏗️  Building Hugo site...${NC}"
hugo --minify --gc

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Hugo build completed successfully${NC}"
else
    echo -e "${RED}❌ Hugo build failed${NC}"
    exit 1
fi

# Change to infrastructure directory
cd infrastructure

# Install Pulumi dependencies
echo -e "${BLUE}📦 Installing Pulumi dependencies...${NC}"
go mod tidy

# Deploy infrastructure
echo -e "${BLUE}☁️  Deploying infrastructure with Pulumi...${NC}"
pulumi up --stack prod --yes

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Infrastructure deployment completed${NC}"
else
    echo -e "${RED}❌ Infrastructure deployment failed${NC}"
    exit 1
fi

# Get bucket name from Pulumi outputs
BUCKET_NAME=$(pulumi stack output bucketName --stack prod)
DISTRIBUTION_ID=$(pulumi stack output distributionId --stack prod)

# Go back to root directory for Hugo files
cd ..

# Sync Hugo public directory to S3
echo -e "${BLUE}📤 Uploading site files to S3...${NC}"
aws s3 sync public/ s3://$BUCKET_NAME --delete --acl public-read

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Files uploaded to S3 successfully${NC}"
else
    echo -e "${RED}❌ S3 upload failed${NC}"
    exit 1
fi

# Invalidate CloudFront cache
echo -e "${BLUE}🔄 Invalidating CloudFront cache...${NC}"
aws cloudfront create-invalidation --distribution-id $DISTRIBUTION_ID --paths "/*"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ CloudFront cache invalidated${NC}"
else
    echo -e "${YELLOW}⚠️  CloudFront invalidation may have failed${NC}"
fi

# Display deployment info
echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
echo -e "${BLUE}📍 Website URL: https://carlosaherrera.com${NC}"
echo -e "${BLUE}📍 WWW URL: https://www.carlosaherrera.com${NC}"
echo -e "${BLUE}🔗 CloudFront Distribution: $DISTRIBUTION_ID${NC}"

echo -e "${YELLOW}💡 Note: It may take a few minutes for changes to propagate globally.${NC}"