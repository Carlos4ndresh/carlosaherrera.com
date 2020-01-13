output "artifact-bucket" {
    description = "S3 bucket for the CodeBuild artifacts"
    value = aws_s3_bucket.build_artifact_bucket.bucket_domain_name
}