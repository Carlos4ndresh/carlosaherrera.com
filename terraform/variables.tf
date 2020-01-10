variable "aws_region"{
    type = "string"
    default = "us-east-2"
}

variable "pipeline_name" {
    type = "string"
    default = "cahp-personalweb-pipeline"
}

variable "github_username" {
  type    = "string"
  default = "Carlos4ndresh"
}

variable "github_token" {
  type = "string"
}

variable "github_repo" {
  type = "string"
}