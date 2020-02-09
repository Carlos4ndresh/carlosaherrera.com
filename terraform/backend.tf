terraform {
    backend "s3" {
        bucket = "terraform-cahpsite-state-bucket"
        key    = "web_tfstate"
        region = "us-east-2"

    }
    required_version = "~>0.12"
}