variable "aws_region"{
    type = string
    default = "us-east-1"
}

variable "pipeline_name" {
    type = string
    default = "cahp-personalweb-pipeline"
}

variable "zone_id" {
  type = string
}

variable "second_domain_zone_id" {
  type = string
}

variable "github_username" {
  type    = string
  default = "Carlos4ndresh"
}

variable "github_token" {
  type = string
}

variable "github_repo" {
  type = string
}

variable "certificate_arn" {
  type = string
}

variable "website_name" {
  type = string
  default = "www.carlosaherrera.com"
}

variable "second_website_name" {
  type = string
  default = "www.carlos4ndresh.com"
}