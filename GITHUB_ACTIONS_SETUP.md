# 🚀 GitHub Actions CI/CD Setup Guide

This guide will help you set up automated deployment for your Hugo site using GitHub Actions with secure OIDC authentication.

## 🏗️ Architecture Overview

```
GitHub Push → GitHub Actions → AWS OIDC → S3 + CloudFront → Live Website
```

- **GitHub Actions**: Builds Hugo site automatically on push
- **AWS OIDC**: Secure authentication without long-term keys
- **S3**: Hosts your static files
- **CloudFront**: Global CDN with automatic cache invalidation

## 📋 Prerequisites

1. ✅ GitHub repository for your site
2. ✅ AWS account with Pulumi infrastructure deployed
3. ✅ Domain configured in Route53

## 🚀 Quick Setup

### Step 1: Deploy Infrastructure with GitHub OIDC

```bash
# Deploy Pulumi infrastructure (includes GitHub OIDC setup)
cd infrastructure
pulumi up --stack prod
```

This creates:
- GitHub OIDC Identity Provider
- IAM Role for GitHub Actions
- S3 and CloudFront permissions

### Step 2: Configure GitHub Repository Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions

Add these **Repository Secrets**:

```bash
AWS_ROLE_ARN=arn:aws:iam::ACCOUNT-ID:role/GitHubActionsRole-carlosaherrera
S3_BUCKET_NAME=carlosaherrera.com
CLOUDFRONT_DISTRIBUTION_ID=E1234567890ABC
```

**To get these values:**

```bash
# Get the Role ARN
pulumi stack output githubActionsRoleArn --stack prod

# Get the S3 bucket name  
pulumi stack output bucketName --stack prod

# Get CloudFront Distribution ID
pulumi stack output distributionId --stack prod
```

### Step 3: Push to Repository

```bash
# Add the workflow file to your repository
git add .github/workflows/deploy.yml
git commit -m "Add GitHub Actions deployment workflow"
git push origin main
```

## 🔧 Workflow Features

### ✅ **Automated Deployment**
- Triggers on push to `main`/`master` branch
- Builds Hugo site with optimizations
- Deploys to S3 with proper caching headers
- Invalidates CloudFront cache

### ✅ **Security**
- Uses OIDC (no long-term AWS keys)
- Least-privilege IAM permissions
- Secure credential handling

### ✅ **Optimization**
- Hugo minification and GC
- Proper S3 content types
- Cache headers for performance
- CloudFront invalidation

### ✅ **Developer Experience**  
- Build summaries in GitHub
- PR preview builds (optional)
- Detailed logging and validation
- Fast deployment (~2-3 minutes)

## 🔍 Monitoring & Debugging

### Check Deployment Status
- Go to your GitHub repository → Actions tab
- View workflow runs and logs
- Check deployment summaries

### AWS CloudWatch Logs
- Monitor S3 access logs
- CloudFront distribution metrics
- IAM role usage

### Troubleshooting Common Issues

#### ❌ **"AssumeRoleWithWebIdentity" errors**
```bash
# Check the role ARN is correct
pulumi stack output githubActionsRoleArn --stack prod

# Verify GitHub repository settings match Pulumi config
cat infrastructure/Pulumi.prod.yaml
```

#### ❌ **S3 permission errors**
```bash
# Check bucket name is correct
pulumi stack output bucketName --stack prod

# Verify IAM policy includes S3 permissions
aws iam get-role-policy --role-name GitHubActionsRole-carlosaherrera --policy-name GitHubActionsPolicy-carlosaherrera
```

#### ❌ **CloudFront invalidation errors**
```bash
# Check distribution ID
pulumi stack output distributionId --stack prod

# Verify CloudFront permissions in IAM policy
```

## 🎯 Advanced Configuration

### Custom Hugo Version
Edit `.github/workflows/deploy.yml`:
```yaml
env:
  HUGO_VERSION: 0.148.2  # Change to your preferred version
```

### Different Branch Deployment
```yaml
on:
  push:
    branches: [main, develop]  # Add more branches
```

### PR Preview Environments
The workflow includes optional PR previews. To enable:

1. Uncomment the `preview` job
2. Set up additional S3 buckets for staging
3. Configure subdomain routing

### Custom Build Commands
Add your own build steps:
```yaml
- name: 🔧 Custom build steps
  run: |
    npm install
    npm run build:css
    hugo --minify --gc
```

## 📊 Performance Metrics

Expected deployment times:
- ⚡ Hugo build: ~30 seconds
- 📤 S3 sync: ~1-2 minutes  
- 🔄 CloudFront invalidation: ~1 minute
- 🎯 **Total**: ~2-3 minutes

## 🔒 Security Best Practices

✅ **Implemented:**
- OIDC instead of access keys
- Least-privilege IAM policies
- Repository-scoped permissions
- Secure secret handling

✅ **Additional recommendations:**
- Enable branch protection rules
- Require PR reviews for main branch
- Use dependabot for dependency updates
- Monitor AWS CloudTrail logs

## 🆘 Support

If you encounter issues:

1. Check GitHub Actions logs
2. Verify AWS permissions
3. Validate Pulumi outputs match GitHub secrets
4. Review CloudWatch logs
5. Test manual deployment with `./deploy.sh`

---

## 🎉 Success!

Once set up, your workflow will:

1. ✅ **Automatically build** your Hugo site on every push
2. ✅ **Deploy to production** within 2-3 minutes  
3. ✅ **Invalidate CDN cache** for immediate updates
4. ✅ **Provide deployment summaries** in GitHub

Your website at `https://carlosaherrera.com` will always reflect your latest code! 🚀