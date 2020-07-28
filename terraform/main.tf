provider "aws" {
  region    = var.aws_region
  version   = "~>2.0"
}


# Uncomment when site is ready to be deployed
# in the meantime test creation of s3, cloudfront, route53 resources
module "pipeline" {
  source = "./pipeline"
  pipeline_name = var.pipeline_name
  github_username = var.github_username
  github_repo = var.github_repo
  github_token = var.github_token
}

resource "aws_s3_bucket" "personal_bucket_logs" {
  bucket = "${var.website_name}-logs"
  acl    = "log-delivery-write"
  region = var.aws_region

  tags = {
    Name        = "Website logs bucket"
    Environment = "Prod"
    owner       = "cherrera"
  }
}

resource "aws_s3_bucket_public_access_block" "block_public_access_personal_bucket_logs_bucket" {
  bucket = aws_s3_bucket.personal_bucket_logs.id

  block_public_acls   = true
  block_public_policy = true
  ignore_public_acls = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket" "website_bucket" {
  bucket = var.website_name
  acl    = "public-read"
  region = var.aws_region
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "OnlyCloudfrontReadAccess",
      "Principal": {
        "AWS": "${aws_cloudfront_origin_access_identity.personal_site_origin_access_identity.iam_arn}"
      },
      "Effect": "Allow",
      "Action": [
        "s3:GetObject"
      ],
      "Resource": "arn:aws:s3:::www.carlosaherrera.com/*"
    }
  ]
}
EOF

  logging {
    target_bucket = aws_s3_bucket.personal_bucket_logs.id
    target_prefix = "www.carlosaherrera.com/"
  }

  website {
    index_document = "index.html"
    error_document = "error.html"
  }

  tags = {
    Name        = "Personal website bucket"
    Environment = "PROD"
    owner = "cherrera"
  }
}

resource "aws_cloudfront_origin_access_identity" "personal_site_origin_access_identity" {
  comment = "The OAI for the website"
}

locals {
  s3_origin_id = "s3_cahp_website_bucket"
}

resource "aws_cloudfront_distribution" "s3_website_distribution" {
  origin {
    domain_name = aws_s3_bucket.website_bucket.bucket_regional_domain_name
    origin_id   = local.s3_origin_id

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.personal_site_origin_access_identity.cloudfront_access_identity_path
    }
  }

  enabled             = true
  is_ipv6_enabled     = true
  comment             = "Cloudfront distribution for personal website"
  default_root_object = "index.html"

  logging_config {
    include_cookies = false
    bucket          = aws_s3_bucket.personal_bucket_logs.bucket_domain_name
    prefix          = "cf_website_logs/"
  }

  aliases = [var.website_name, substr(var.website_name,4,19), var.second_website_name, substr(var.second_website_name,4,22)]

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = local.s3_origin_id

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 300
    max_ttl                = 1200
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  price_class = "PriceClass_All"

  tags = {
    Environment = "production"
  }

  viewer_certificate {
    acm_certificate_arn = var.certificate_arn
    ssl_support_method = "sni-only"
  }

  custom_error_response {
    error_code = 404
  }
}

resource "aws_route53_record" "www" {
  zone_id = var.zone_id
  name    = var.website_name
  type    = "A"
  alias {
    name = aws_cloudfront_distribution.s3_website_distribution.domain_name
    zone_id = aws_cloudfront_distribution.s3_website_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "second_www" {
  zone_id = var.second_domain_zone_id
  name    = var.second_website_name
  type    = "A"
  alias {
    name = aws_cloudfront_distribution.s3_website_distribution.domain_name
    zone_id = aws_cloudfront_distribution.s3_website_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "second_root" {
  zone_id = var.second_domain_zone_id
  name    = substr(var.second_website_name,4,22)
  type    = "A"
  alias {
    name = aws_cloudfront_distribution.s3_website_distribution.domain_name
    zone_id = aws_cloudfront_distribution.s3_website_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "root_www" {
  zone_id = var.zone_id
  name    = substr(var.website_name,4,19)
  type    = "A"
  alias {
    name = aws_cloudfront_distribution.s3_website_distribution.domain_name
    zone_id = aws_cloudfront_distribution.s3_website_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}
