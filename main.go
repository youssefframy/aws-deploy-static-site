package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	bucketName            = "example-bucket-name"
	websiteFolderPath     = "relative/path/to/your/website"
	cloudfrontDescription = "description"
	region                = "region"
)

func main() {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Initialize service clients
	s3Client := s3.NewFromConfig(cfg)
	cfClient := cloudfront.NewFromConfig(cfg)
	iamClient := iam.NewFromConfig(cfg)

	// Create S3 bucket
	bucket, err := createS3Bucket(context.TODO(), s3Client)
	if err != nil {
		log.Fatalf("Failed to create S3 bucket: %v", err)
	}

	// Create Origin Access Control
	oacId, err := createOriginAccessControl(context.TODO(), cfClient)
	if err != nil {
		log.Fatalf("Failed to create Origin Access Control: %v", err)
	}

	// Create CloudFront distribution
	distribution, err := createCloudFrontDistribution(context.TODO(), cfClient, bucket, oacId)
	if err != nil {
		log.Fatalf("Failed to create CloudFront distribution: %v", err)
	}

	// Create and attach bucket policy
	err = attachBucketPolicy(context.TODO(), s3Client, iamClient, bucket, distribution.Id)
	if err != nil {
		log.Fatalf("Failed to attach bucket policy: %v", err)
	}

	// Upload website files
	err = uploadWebsiteFiles(context.TODO(), s3Client)
	if err != nil {
		log.Fatalf("Failed to upload website files: %v", err)
	}

	fmt.Printf("Deployment completed successfully!\n")
	fmt.Printf("Bucket Name: %s\n", bucketName)
	fmt.Printf("Distribution ID: %s\n", *distribution.Id)
	fmt.Printf("Distribution Domain Name: %s\n", *distribution.DomainName)
}

func createS3Bucket(ctx context.Context, client *s3.Client) (*string, error) {
	bucket := bucketName
	input := &s3.CreateBucketInput{
		Bucket: &bucket,
	}
	
	// Only add LocationConstraint for regions other than us-east-1
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(region),
		}
	}
	
	_, err := client.CreateBucket(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	// Block public access
	_, err = client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: &bucket,
		PublicAccessBlockConfiguration: &s3Types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to block public access: %w", err)
	}

	return &bucket, nil
}

func createOriginAccessControl(ctx context.Context, client *cloudfront.Client) (*string, error) {
	resp, err := client.CreateOriginAccessControl(ctx, &cloudfront.CreateOriginAccessControlInput{
		OriginAccessControlConfig: &types.OriginAccessControlConfig{
			Name: aws.String(fmt.Sprintf("%s-origin-access-control", bucketName)),
			OriginAccessControlOriginType: "s3",
			SigningBehavior: "always",
			SigningProtocol: "sigv4",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create origin access control: %w", err)
	}

	return resp.OriginAccessControl.Id, nil
}

func createCloudFrontDistribution(ctx context.Context, client *cloudfront.Client, bucketName, oacId *string) (*types.Distribution, error) {
	resp, err := client.CreateDistribution(ctx, &cloudfront.CreateDistributionInput{
		DistributionConfig: &types.DistributionConfig{
			CallerReference:    aws.String(fmt.Sprintf("cli-%d", time.Now().Unix())),
			Comment:           aws.String(cloudfrontDescription),
			Enabled:           aws.Bool(true),
			IsIPV6Enabled:     aws.Bool(true),
			Origins: &types.Origins{
				Items: []types.Origin{
					{
						DomainName:              aws.String(fmt.Sprintf("%s.s3.%s.amazonaws.com", *bucketName, region)),
						Id:                      aws.String("S3Origin"),
						OriginAccessControlId:   oacId,
						S3OriginConfig:          &types.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
				Quantity: aws.Int32(1),
			},
			DefaultCacheBehavior: &types.DefaultCacheBehavior{
				TargetOriginId:       aws.String("S3Origin"),
				ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps,
				Compress:             aws.Bool(true),
				CachePolicyId:        aws.String("658327ea-f89d-4fab-a63d-7e88639e58f6"),
			},
			CustomErrorResponses: &types.CustomErrorResponses{
				Items: []types.CustomErrorResponse{
					{
						ErrorCode:          aws.Int32(403),
						ResponseCode:       aws.String("200"),
						ResponsePagePath:   aws.String("/index.html"),
						ErrorCachingMinTTL: aws.Int64(0),
					},
					{
						ErrorCode:          aws.Int32(404),
						ResponseCode:       aws.String("200"),
						ResponsePagePath:   aws.String("/index.html"),
						ErrorCachingMinTTL: aws.Int64(0),
					},
				},
				Quantity: aws.Int32(2),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create distribution: %w", err)
	}

	return resp.Distribution, nil
}

func attachBucketPolicy(ctx context.Context, s3Client *s3.Client, iamClient *iam.Client, bucketName, distributionId *string) error {
	accountID, err := getAWSAccountID(ctx, iamClient)
	if err != nil {
		return fmt.Errorf("failed to get AWS account ID: %w", err)
	}

	policyDocument := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {
				"Service": "cloudfront.amazonaws.com"
			},
			"Action": "s3:GetObject",
			"Resource": "arn:aws:s3:::%s/*",
			"Condition": {
				"StringEquals": {
					"AWS:SourceArn": "arn:aws:cloudfront::%s:distribution/%s"
				}
			}
		}]
	}`, *bucketName, accountID, *distributionId)

	_, err = s3Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: bucketName,
		Policy: &policyDocument,
	})
	return err
}

func uploadWebsiteFiles(ctx context.Context, client *s3.Client) error {
	absPath, err := filepath.Abs(websiteFolderPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Calculate the key (path relative to the website folder)
		relPath, err := filepath.Rel(absPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Upload file to S3
		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(relPath),
			Body:   bytes.NewReader(content),
			CacheControl: aws.String("public, max-age=31536000, s-maxage=31536000"),
		})
		if err != nil {
			return fmt.Errorf("failed to upload file %s: %w", relPath, err)
		}

		return nil
	})

	return err
}

func getAWSAccountID(ctx context.Context, client *iam.Client) (string, error) {
	result, err := client.GetUser(ctx, &iam.GetUserInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	// Extract account ID from user ARN
	// ARN format: arn:aws:iam::ACCOUNT-ID:user/USER-NAME
	parts := strings.Split(*result.User.Arn, ":")
	if len(parts) < 5 {
		return "", fmt.Errorf("invalid ARN format")
	}

	return parts[4], nil
}