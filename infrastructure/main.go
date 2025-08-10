package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "")
		domain := conf.Require("domain")
		wwwDomain := fmt.Sprintf("www.%s", domain)

		// Look up the existing Route53 zone
		zone, err := route53.LookupZone(ctx, &route53.LookupZoneArgs{
			Name: &domain,
		})
		if err != nil {
			return fmt.Errorf("error looking up Route53 zone for domain %s: %v", domain, err)
		}

		// Create S3 bucket for website hosting
		siteBucket, err := s3.NewBucket(ctx, "site-bucket", &s3.BucketArgs{
			Bucket: pulumi.String(domain),
		})
		if err != nil {
			return err
		}

		// Configure bucket for static website hosting
		_, err = s3.NewBucketWebsiteConfigurationV2(ctx, "site-bucket-website", &s3.BucketWebsiteConfigurationV2Args{
			Bucket: siteBucket.ID(),
			IndexDocument: &s3.BucketWebsiteConfigurationV2IndexDocumentArgs{
				Suffix: pulumi.String("index.html"),
			},
			ErrorDocument: &s3.BucketWebsiteConfigurationV2ErrorDocumentArgs{
				Key: pulumi.String("404.html"),
			},
		})
		if err != nil {
			return err
		}

		// Enable versioning
		_, err = s3.NewBucketVersioningV2(ctx, "site-bucket-versioning", &s3.BucketVersioningV2Args{
			Bucket: siteBucket.ID(),
			VersioningConfiguration: &s3.BucketVersioningV2VersioningConfigurationArgs{
				Status: pulumi.String("Enabled"),
			},
		})
		if err != nil {
			return err
		}

		// Create US East 1 provider for ACM (CloudFront requires certificates in us-east-1)
		usEast1Provider, err := aws.NewProvider(ctx, "aws-us-east-1", &aws.ProviderArgs{
			Region: pulumi.String("us-east-1"),
		})
		if err != nil {
			return err
		}

		// Request SSL certificate for both domain and www subdomain
		certificate, err := acm.NewCertificate(ctx, "ssl-cert", &acm.CertificateArgs{
			DomainName:       pulumi.String(domain),
			ValidationMethod: pulumi.String("DNS"),
			SubjectAlternativeNames: pulumi.StringArray{
				pulumi.String(wwwDomain),
			},
		}, pulumi.Provider(usEast1Provider))
		if err != nil {
			return err
		}

		// Create validation records for certificate
		certificate.DomainValidationOptions.ApplyT(func(options []acm.CertificateDomainValidationOption) error {
			for i, option := range options {
				recordName := fmt.Sprintf("validation-record-%d", i)
				_, err := route53.NewRecord(ctx, recordName, &route53.RecordArgs{
					Name:   pulumi.String(*option.ResourceRecordName),
					Type:   pulumi.String(*option.ResourceRecordType),
					ZoneId: pulumi.String(zone.ZoneId),
					Records: pulumi.StringArray{
						pulumi.String(*option.ResourceRecordValue),
					},
					Ttl: pulumi.Int(300),
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

		// Wait for certificate validation
		certValidation, err := acm.NewCertificateValidation(ctx, "cert-validation", &acm.CertificateValidationArgs{
			CertificateArn: certificate.Arn,
		}, pulumi.Provider(usEast1Provider))
		if err != nil {
			return err
		}

		// Create CloudFront Origin Access Control (OAC)
		oac, err := cloudfront.NewOriginAccessControl(ctx, "site-oac", &cloudfront.OriginAccessControlArgs{
			Name:                          pulumi.String(fmt.Sprintf("%s-oac", domain)),
			Description:                   pulumi.String("Origin Access Control for carlosaherrera.com"),
			OriginAccessControlOriginType: pulumi.String("s3"),
			SigningBehavior:               pulumi.String("always"),
			SigningProtocol:               pulumi.String("sigv4"),
		})
		if err != nil {
			return err
		}

		// Create CloudFront distribution
		distribution, err := cloudfront.NewDistribution(ctx, "site-distribution", &cloudfront.DistributionArgs{
			Enabled: pulumi.Bool(true),
			Comment: pulumi.String("Carlos A. Herrera personal website"),

			Origins: cloudfront.DistributionOriginArray{
				&cloudfront.DistributionOriginArgs{
					DomainName:            siteBucket.BucketDomainName,
					OriginId:              pulumi.String("s3-origin"),
					OriginAccessControlId: oac.ID(),
				},
			},

			DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
				TargetOriginId:       pulumi.String("s3-origin"),
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				Compress:             pulumi.Bool(true),

				AllowedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
					pulumi.String("OPTIONS"),
				},
				CachedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
				},

				ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
					QueryString: pulumi.Bool(false),
					Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
						Forward: pulumi.String("none"),
					},
				},

				MinTtl:     pulumi.Int(0),
				DefaultTtl: pulumi.Int(3600),  // 1 hour
				MaxTtl:     pulumi.Int(86400), // 1 day
			},

			// Custom error pages for better UX
			CustomErrorResponses: cloudfront.DistributionCustomErrorResponseArray{
				&cloudfront.DistributionCustomErrorResponseArgs{
					ErrorCode:          pulumi.Int(404),
					ResponseCode:       pulumi.Int(404),
					ResponsePagePath:   pulumi.String("/404.html"),
					ErrorCachingMinTtl: pulumi.Int(300),
				},
				&cloudfront.DistributionCustomErrorResponseArgs{
					ErrorCode:          pulumi.Int(403),
					ResponseCode:       pulumi.Int(404),
					ResponsePagePath:   pulumi.String("/404.html"),
					ErrorCachingMinTtl: pulumi.Int(300),
				},
			},

			Aliases: pulumi.StringArray{
				pulumi.String(domain),
				pulumi.String(wwwDomain),
			},

			ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
				AcmCertificateArn:      certValidation.CertificateArn,
				SslSupportMethod:       pulumi.String("sni-only"),
				MinimumProtocolVersion: pulumi.String("TLSv1.2_2021"),
			},

			PriceClass:        pulumi.String("PriceClass_100"),
			HttpVersion:       pulumi.String("http2and3"),
			IsIpv6Enabled:     pulumi.Bool(true),
			DefaultRootObject: pulumi.String("index.html"),

			Restrictions: &cloudfront.DistributionRestrictionsArgs{
				GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
					RestrictionType: pulumi.String("none"),
				},
			},

			Tags: pulumi.StringMap{
				"Name":        pulumi.String("Carlos Herrera Website"),
				"Environment": pulumi.String("production"),
				"Project":     pulumi.String("personal-website"),
			},
		}, pulumi.DependsOn([]pulumi.Resource{certValidation}))
		if err != nil {
			return err
		}

		// Create bucket policy to allow CloudFront access via OAC
		bucketPolicyJSON := pulumi.All(siteBucket.Arn, distribution.Arn).ApplyT(func(args []interface{}) string {
			bucketArn := args[0].(string)
			distributionArn := args[1].(string)
			
			return fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Sid": "AllowCloudFrontServicePrincipal",
						"Effect": "Allow",
						"Principal": {
							"Service": "cloudfront.amazonaws.com"
						},
						"Action": "s3:GetObject",
						"Resource": "%s/*",
						"Condition": {
							"StringEquals": {
								"AWS:SourceArn": "%s"
							}
						}
					}
				]
			}`, bucketArn, distributionArn)
		}).(pulumi.StringOutput)

		_, err = s3.NewBucketPolicy(ctx, "site-bucket-policy", &s3.BucketPolicyArgs{
			Bucket: siteBucket.ID(),
			Policy: bucketPolicyJSON,
		})
		if err != nil {
			return err
		}

		// Block all public access to the bucket (CloudFront will access via OAC)
		_, err = s3.NewBucketPublicAccessBlock(ctx, "site-bucket-pab", &s3.BucketPublicAccessBlockArgs{
			Bucket:                siteBucket.ID(),
			BlockPublicAcls:       pulumi.Bool(true),
			BlockPublicPolicy:     pulumi.Bool(true),
			IgnorePublicAcls:      pulumi.Bool(true),
			RestrictPublicBuckets: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Create Route53 records for the domain
		_, err = route53.NewRecord(ctx, "root-domain-record", &route53.RecordArgs{
			Name:   pulumi.String(domain),
			Type:   pulumi.String("A"),
			ZoneId: pulumi.String(zone.ZoneId),
			Aliases: route53.RecordAliasArray{
				&route53.RecordAliasArgs{
					Name:                 distribution.DomainName,
					ZoneId:               distribution.HostedZoneId,
					EvaluateTargetHealth: pulumi.Bool(false),
				},
			},
		})
		if err != nil {
			return err
		}

		_, err = route53.NewRecord(ctx, "www-domain-record", &route53.RecordArgs{
			Name:   pulumi.String(wwwDomain),
			Type:   pulumi.String("A"),
			ZoneId: pulumi.String(zone.ZoneId),
			Aliases: route53.RecordAliasArray{
				&route53.RecordAliasArgs{
					Name:                 distribution.DomainName,
					ZoneId:               distribution.HostedZoneId,
					EvaluateTargetHealth: pulumi.Bool(false),
				},
			},
		})
		if err != nil {
			return err
		}

		// Create GitHub OIDC Provider and IAM Role for Actions
		err = createGitHubActionsRole(ctx, siteBucket, distribution)
		if err != nil {
			return err
		}

		// Export important values
		ctx.Export("bucketName", siteBucket.ID())
		ctx.Export("bucketDomainName", siteBucket.BucketDomainName)
		ctx.Export("distributionId", distribution.ID())
		ctx.Export("distributionDomainName", distribution.DomainName)
		ctx.Export("certificateArn", certificate.Arn)
		ctx.Export("websiteUrl", pulumi.Sprintf("https://%s", domain))
		ctx.Export("wwwWebsiteUrl", pulumi.Sprintf("https://%s", wwwDomain))

		return nil
	})
}

// createGitHubActionsRole creates OIDC provider and IAM role for GitHub Actions
func createGitHubActionsRole(ctx *pulumi.Context, siteBucket *s3.Bucket, distribution *cloudfront.Distribution) error {
	conf := config.New(ctx, "")
	githubRepo := conf.Get("githubRepo")
	githubOwner := conf.Get("githubOwner")

	// Default to carlosaherrera.com if not specified
	if githubRepo == "" {
		githubRepo = "carlosaherrera.com"
	}
	if githubOwner == "" {
		githubOwner = "Carlos4ndresh" // Update this to your GitHub username
	}

	// Create GitHub OIDC Identity Provider
	githubOidc, err := iam.NewOpenIdConnectProvider(ctx, "github-oidc", &iam.OpenIdConnectProviderArgs{
		ClientIdLists: pulumi.StringArray{
			pulumi.String("sts.amazonaws.com"),
		},
		ThumbprintLists: pulumi.StringArray{
			pulumi.String("6938fd4d98bab03faadb97b34396831e3780aea1"),
			pulumi.String("1c58a3a8518e8759bf075b76b750d4f2df264fcd"),
		},
		Url: pulumi.String("https://token.actions.githubusercontent.com"),
		Tags: pulumi.StringMap{
			"Name":    pulumi.String("GitHub Actions OIDC"),
			"Project": pulumi.String("carlosaherrera.com"),
		},
	})
	if err != nil {
		return err
	}

	// Create IAM role for GitHub Actions
	githubActionsRole, err := iam.NewRole(ctx, "github-actions-role", &iam.RoleArgs{
		Name: pulumi.String("GitHubActionsRole-carlosaherrera"),
		AssumeRolePolicy: pulumi.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"Federated": "%s"
					},
					"Action": "sts:AssumeRoleWithWebIdentity",
					"Condition": {
						"StringEquals": {
							"token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
						},
						"StringLike": {
							"token.actions.githubusercontent.com:sub": "repo:%s/%s:*"
						}
					}
				}
			]
		}`, githubOidc.Arn, githubOwner, githubRepo),
		Tags: pulumi.StringMap{
			"Name":    pulumi.String("GitHub Actions Deployment Role"),
			"Project": pulumi.String("carlosaherrera.com"),
		},
	})
	if err != nil {
		return err
	}

	// Create policy for GitHub Actions role
	githubActionsPolicy, err := iam.NewPolicy(ctx, "github-actions-policy", &iam.PolicyArgs{
		Name:        pulumi.String("GitHubActionsPolicy-carlosaherrera"),
		Description: pulumi.String("Policy for GitHub Actions to deploy carlosaherrera.com"),
		Policy: pulumi.All(siteBucket.Arn, distribution.Arn).ApplyT(func(args []interface{}) string {
			bucketArn := args[0].(string)
			distributionArn := args[1].(string)

			return fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Action": [
							"s3:PutObject",
							"s3:PutObjectAcl",
							"s3:GetObject",
							"s3:DeleteObject",
							"s3:ListBucket"
						],
						"Resource": [
							"%s",
							"%s/*"
						]
					},
					{
						"Effect": "Allow",
						"Action": [
							"cloudfront:CreateInvalidation",
							"cloudfront:GetInvalidation"
						],
						"Resource": "%s"
					}
				]
			}`, bucketArn, bucketArn, distributionArn)
		}).(pulumi.StringOutput),
		Tags: pulumi.StringMap{
			"Name":    pulumi.String("GitHub Actions Policy"),
			"Project": pulumi.String("carlosaherrera.com"),
		},
	})
	if err != nil {
		return err
	}

	// Attach policy to role
	_, err = iam.NewRolePolicyAttachment(ctx, "github-actions-policy-attachment", &iam.RolePolicyAttachmentArgs{
		Role:      githubActionsRole.Name,
		PolicyArn: githubActionsPolicy.Arn,
	})
	if err != nil {
		return err
	}

	// Export GitHub Actions related values
	ctx.Export("githubActionsRoleArn", githubActionsRole.Arn)
	ctx.Export("githubOidcProviderArn", githubOidc.Arn)

	return nil
}
