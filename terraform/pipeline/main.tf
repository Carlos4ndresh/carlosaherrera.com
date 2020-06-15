# CodePipeline resources
resource "aws_s3_bucket" "build_artifact_bucket" {
  bucket = "${var.pipeline_name}-artifact-bucket"
  acl    = "private"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "block_public_access_artifact_bucket" {
  bucket = aws_s3_bucket.build_artifact_bucket.id

  block_public_acls   = true
  block_public_policy = true
  ignore_public_acls = true
  restrict_public_buckets = true
}


data "aws_iam_policy_document" "codepipeline_web_assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["codepipeline.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "codepipeline_role" {
  name               = "${var.pipeline_name}-codepipeline-role"
  assume_role_policy = data.aws_iam_policy_document.codepipeline_web_assume_policy.json
}

resource "aws_iam_role_policy" "attach_codepipelineweb_policy" {
    name = "${var.pipeline_name}-codepipeline-policy"
    role = aws_iam_role.codepipeline_role.id

    policy = <<EOF
{
      "Statement": [
        {
            "Action": [
                "s3:GetObject",
                "s3:GetObjectVersion",
                "s3:GetBucketVersioning",
                "s3:PutObject",
                "s3:ListBucket",
                "s3:DeleteObject"
            ],
            "Resource": "*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "cloudwatch:*",
                "sns:*",
                "sqs:*",
                "iam:PassRole"
            ],
            "Resource": "*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "codebuild:BatchGetBuilds",
                "codebuild:StartBuild"
            ],
            "Resource": "*",
            "Effect": "Allow"
        }
    ],
    "Version": "2012-10-17"
    }
    EOF
}

resource "aws_iam_role" "codebuild_assume_role" {
  name = "${var.pipeline_name}-codebuild-role"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Service": "codebuild.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
      }
    ]
  }
  EOF
}

resource "aws_iam_role_policy" "codebuild_policy" {
  name = "${var.pipeline_name}-codebuild-policy"
  role = aws_iam_role.codebuild_assume_role.id

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning",
        "s3:ListBucket",
        "s3:DeleteObject"
        ],
        "Resource": "*",
        "Effect": "Allow"
      },
      {
        "Effect": "Allow",
        "Resource": [
            "${aws_codebuild_project.build_personalweb_project.id}"
        ],
        "Action": [
          "codebuild:*"
        ]
      },
      {
        "Effect": "Allow",
        "Resource": [
          "*"
        ],
        "Action": [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
      }
    ]
  }
  POLICY
}

resource "aws_codebuild_project" "build_personalweb_project" {
  name = "${var.pipeline_name}-build"
  description = "The CodeBuild project for ${var.pipeline_name}"
  service_role = aws_iam_role.codebuild_assume_role.arn
  build_timeout = "60"
  # badge_enabled = true

  artifacts {
    type = "CODEPIPELINE"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image = "aws/codebuild/standard:4.0"
    type = "LINUX_CONTAINER"
  }

  source {
    type = "CODEPIPELINE"
    buildspec = "wsite/buildspec.yml"
  }

}


resource "aws_codepipeline" "codepipeline_personalweb" {
  name = "${var.pipeline_name}-codepipeline"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.build_artifact_bucket.bucket
    type = "S3"
  }

  stage {
    name = "Source"

    action {
      name = "Source"
      category = "Source"
      owner = "ThirdParty"
      provider = "GitHub"
      version = "1"
      output_artifacts = ["code"]

      configuration = {
        Owner = var.github_username
        OAuthToken = var.github_token
        Repo                 = var.github_repo
        Branch               = "master"
        PollForSourceChanges = "true"
      }
    }
  }

  stage {
    name = "DeployToS3"

    action {
      name             = "DeployToS3"
      category         = "Test"
      owner            = "AWS"
      provider         = "CodeBuild"
      input_artifacts  = ["code"]
      output_artifacts = ["deployed"]
      version          = "1"

      configuration = {
        ProjectName = aws_codebuild_project.build_personalweb_project.name
      }
    }
  }

}