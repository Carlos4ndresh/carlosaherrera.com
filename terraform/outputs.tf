output "website_bucket" {
    value = aws_s3_bucket.website_bucket.bucket_domain_name
} 

# output "badge_url" {
#     value = module.pipeline.badge_url
# }