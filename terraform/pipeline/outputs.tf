output "artifact-bucket" {
    description = "S3 bucket for the CodeBuild artifacts"
    value = aws_s3_bucket.build_artifact_bucket.bucket_domain_name
}

# output "badge_url" {
#     description = "CodeBuild Badge"
#     value = aws_codebuild_project.build_personalweb_project.badge_url
# }