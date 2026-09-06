package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudfront"
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

		zone, err := route53.LookupZone(ctx, &route53.LookupZoneArgs{
			Name: &domain,
		})
		if err != nil {
			return fmt.Errorf("error looking up Route53 zone for domain %s: %v", domain, err)
		}

		siteBucket, err := s3.NewBucket(ctx, "site-bucket", &s3.BucketArgs{
			Bucket: pulumi.String(domain),
		})
		if err != nil {
			return err
		}

		_, err = s3.NewBucketVersioningV2(ctx, "site-bucket-versioning", &s3.BucketVersioningV2Args{
			Bucket: siteBucket.ID(),
			VersioningConfiguration: &s3.BucketVersioningV2VersioningConfigurationArgs{
				Status: pulumi.String("Enabled"),
			},
		})
		if err != nil {
			return err
		}

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

		oac, err := cloudfront.NewOriginAccessControl(ctx, "site-oac", &cloudfront.OriginAccessControlArgs{
			Name:                          pulumi.String(fmt.Sprintf("%s-oac", domain)),
			OriginAccessControlOriginType: pulumi.String("s3"),
			SigningBehavior:               pulumi.String("always"),
			SigningProtocol:               pulumi.String("sigv4"),
		})
		if err != nil {
			return err
		}

		usEast1Provider, err := aws.NewProvider(ctx, "aws-us-east-1", &aws.ProviderArgs{
			Region: pulumi.String("us-east-1"),
		})
		if err != nil {
			return err
		}

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

		certValidation, err := acm.NewCertificateValidation(ctx, "cert-validation", &acm.CertificateValidationArgs{
			CertificateArn: certificate.Arn,
		}, pulumi.Provider(usEast1Provider))
		if err != nil {
			return err
		}

		distribution, err := cloudfront.NewDistribution(ctx, "site-distribution", &cloudfront.DistributionArgs{
			Enabled: pulumi.Bool(true),
			Comment: pulumi.String("Carlos A. Herrera personal website"),

			Origins: cloudfront.DistributionOriginArray{
				&cloudfront.DistributionOriginArgs{
					DomainName:            siteBucket.BucketRegionalDomainName,
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
				DefaultTtl: pulumi.Int(3600),
				MaxTtl:     pulumi.Int(86400),
			},

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
				MinimumProtocolVersion: pulumi.String("TLSv1.2_2025"),
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

		// Bucket policy granting CloudFront OAC access only
		bucketPolicyJSON := pulumi.All(siteBucket.Arn, distribution.Arn).ApplyT(func(args []interface{}) string {
			bucketArn := args[0].(string)
			distArn := args[1].(string)
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
			}`, bucketArn, distArn)
		}).(pulumi.StringOutput)

		_, err = s3.NewBucketPolicy(ctx, "site-bucket-policy", &s3.BucketPolicyArgs{
			Bucket: siteBucket.ID(),
			Policy: bucketPolicyJSON,
		})
		if err != nil {
			return err
		}

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

		ctx.Export("bucketName", siteBucket.ID())
		ctx.Export("bucketArn", siteBucket.Arn)
		ctx.Export("distributionId", distribution.ID())
		ctx.Export("distributionArn", distribution.Arn)
		ctx.Export("distributionDomainName", distribution.DomainName)
		ctx.Export("certificateArn", certificate.Arn)
		ctx.Export("websiteUrl", pulumi.Sprintf("https://%s", domain))

		return nil
	})
}
