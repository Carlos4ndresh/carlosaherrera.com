# Carlos A. Herrera Website Infrastructure

This directory contains the Pulumi infrastructure-as-code for deploying the modern Hugo website for Carlos A. Herrera.

## Architecture

The infrastructure creates a modern, secure, and performant static website deployment using:

- **S3 Bucket**: Hosts the static Hugo-generated files
- **CloudFront Distribution**: Global CDN with HTTP/2 and HTTP/3 support
- **ACM Certificate**: SSL/TLS certificate for HTTPS (both apex and www)
- **Route53 Records**: DNS records pointing to CloudFront
- **Origin Access Control (OAC)**: Secure access from CloudFront to S3

## Key Features

✅ **Modern Security**: Uses Origin Access Control (OAC) instead of deprecated OAI  
✅ **HTTP/3 Support**: Latest protocol for best performance  
✅ **Global CDN**: CloudFront edge locations worldwide  
✅ **SSL/TLS**: Automatic HTTPS with ACM certificates  
✅ **Both Domains**: Supports both `carlosaherrera.com` and `www.carlosaherrera.com`  
✅ **S3 Security**: Private bucket with secure CloudFront access only  
✅ **Proper Error Pages**: Custom 404 handling  

## Prerequisites

1. **Pulumi CLI** installed
2. **AWS CLI** configured with appropriate permissions
3. **Go** runtime (for Pulumi Go provider)
4. **Route53 Hosted Zone** for `carlosaherrera.com` already exists

## Required AWS Permissions

Your AWS credentials need permissions for:
- S3 (bucket creation, policy management)
- CloudFront (distribution management, OAC)
- ACM (certificate request, validation)
- Route53 (record management)
- IAM (policy creation for bucket access)

## Deployment

### Initial Setup

```bash
# Navigate to infrastructure directory
cd infrastructure

# Install dependencies
go mod tidy

# Login to Pulumi (if not already done)
pulumi login

# Create/select the production stack
pulumi stack init prod
# OR
pulumi stack select prod
```

### Deploy Infrastructure

```bash
# Preview changes
pulumi preview --stack prod

# Deploy infrastructure
pulumi up --stack prod
```

### Deploy Website Content

Use the automated deployment script from the root directory:

```bash
# From project root
./deploy.sh
```

Or manually:

```bash
# Build Hugo site
hugo --minify --gc

# Get bucket name from Pulumi
BUCKET_NAME=$(pulumi stack output bucketName --stack prod)

# Upload to S3
aws s3 sync public/ s3://$BUCKET_NAME --delete

# Invalidate CloudFront cache
DISTRIBUTION_ID=$(pulumi stack output distributionId --stack prod)
aws cloudfront create-invalidation --distribution-id $DISTRIBUTION_ID --paths "/*"
```

## Configuration

The infrastructure is configured via `Pulumi.prod.yaml`:

```yaml
config:
  aws:region: ca-central-1
  carlosaherrera.com:domain: "carlosaherrera.com"
```

## Outputs

After deployment, Pulumi exports these values:

- `bucketName`: S3 bucket name for uploading content
- `distributionId`: CloudFront distribution ID for cache invalidation
- `distributionDomainName`: CloudFront domain name
- `certificateArn`: ACM certificate ARN
- `websiteUrl`: Main website URL (https://carlosaherrera.com)
- `wwwWebsiteUrl`: WWW website URL (https://www.carlosaherrera.com)

## Monitoring

Monitor your deployment:

- **CloudWatch**: AWS service metrics and logs
- **CloudFront Reports**: Traffic and performance analytics
- **ACM Certificate**: Expiration and renewal status

## Troubleshooting

### Certificate Validation
- Ensure Route53 hosted zone exists for your domain
- DNS validation records are created automatically
- Certificate must be in `us-east-1` for CloudFront

### S3 Access Issues
- Bucket policy allows only CloudFront OAC access
- All public access is blocked for security
- Files must be uploaded to the correct bucket

### CloudFront Distribution
- Changes can take 15-20 minutes to propagate
- Use cache invalidation for immediate updates
- HTTP/2 and HTTP/3 enabled by default

## Security Best Practices

✅ **Private S3 Bucket**: No direct public access  
✅ **OAC Authentication**: Modern CloudFront security  
✅ **HTTPS Only**: Redirects HTTP to HTTPS  
✅ **TLS 1.2+**: Modern encryption standards  
✅ **No Hardcoded Secrets**: Uses AWS IAM roles  

## Cost Optimization

- **PriceClass_100**: Uses only US, Canada, and Europe edge locations
- **Compression**: Enabled for smaller transfers  
- **Caching**: 1-hour default TTL reduces origin requests
- **S3 Versioning**: Enabled for content management

## Updates

To update the infrastructure:

1. Modify the Pulumi code
2. Run `pulumi preview` to see changes
3. Run `pulumi up` to apply changes
4. Redeploy content if needed with `./deploy.sh`