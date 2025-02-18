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

	"bufio"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	bucketName            string
	websiteFolderPath     string
	cloudfrontDescription string
	region                string
	profileName           string
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for AWS profile
	fmt.Print("🔑 Enter AWS profile name (press Enter for default): ")
	profileName, _ = reader.ReadString('\n')
	profileName = strings.TrimSpace(profileName)

	// Prompt for bucket name
	fmt.Print("🪣 Enter S3 bucket name: ")
	bucketName, _ = reader.ReadString('\n')
	bucketName = strings.TrimSpace(bucketName)

	// Prompt for website folder path
	fmt.Print("📂 Enter local path to your website files: ")
	websiteFolderPath, _ = reader.ReadString('\n')
	websiteFolderPath = strings.TrimSpace(websiteFolderPath)

	// Prompt for CloudFront description
	fmt.Print("💬 Enter CloudFront distribution description (press Enter for default): ")
	cloudfrontDescription, _ = reader.ReadString('\n')
	cloudfrontDescription = strings.TrimSpace(cloudfrontDescription)
	if cloudfrontDescription == "" {
		cloudfrontDescription = fmt.Sprintf("CloudFront distribution for %s", bucketName)
	}

	// Prompt for AWS region
	fmt.Print("🌐 Enter AWS region (e.g., us-east-1): ")
	region, _ = reader.ReadString('\n')
	region = strings.TrimSpace(region)

	// Validate inputs
	if bucketName == "" || websiteFolderPath == "" || region == "" {
		log.Fatal("Error: All inputs except description are required")
	}

	// Update AWS configuration to include profile option
	var cfg aws.Config
	var err error
	if profileName != "" {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithSharedConfigProfile(profileName),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
		)
	}
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Initialize service clients
	s3Client := s3.NewFromConfig(cfg)
	cfClient := cloudfront.NewFromConfig(cfg)
	iamClient := iam.NewFromConfig(cfg)

	// Create S3 bucket
	fmt.Printf("📦 Creating S3 bucket '%s'...\n", bucketName)
	bucket, err := createS3Bucket(context.TODO(), s3Client)
	if err != nil {
		log.Fatalf("Failed to create S3 bucket: %v", err)
	}
	fmt.Printf("✅ S3 bucket created successfully\n\n")

	// Create Origin Access Control
	fmt.Printf("🔒 Creating Origin Access Control...\n")
	oacId, err := createOriginAccessControl(context.TODO(), cfClient)
	if err != nil {
		log.Fatalf("Failed to create Origin Access Control: %v", err)
	}
	fmt.Printf("✅ Origin Access Control created successfully\n\n")

	// Create CloudFront distribution
	fmt.Printf("☁️  Creating CloudFront distribution...\n")
	distribution, err := createCloudFrontDistribution(context.TODO(), cfClient, bucket, oacId)
	if err != nil {
		log.Fatalf("Failed to create CloudFront distribution: %v", err)
	}
	fmt.Printf("✅ CloudFront distribution created successfully\n\n")

	// Create and attach bucket policy
	fmt.Printf("📜 Attaching bucket policy...\n")
	err = attachBucketPolicy(context.TODO(), s3Client, iamClient, bucket, distribution.Id)
	if err != nil {
		log.Fatalf("Failed to attach bucket policy: %v", err)
	}
	fmt.Printf("✅ Bucket policy attached successfully\n\n")

	// Upload website files
	fmt.Printf("📤 Uploading website files...\n")
	err = uploadWebsiteFiles(context.TODO(), s3Client)
	if err != nil {
		log.Fatalf("Failed to upload website files: %v", err)
	}
	fmt.Printf("✅ Website files uploaded successfully\n\n")

	fmt.Printf("🎉 Deployment completed successfully!\n")
	fmt.Printf("📋 Summary:\n")
	fmt.Printf("   Bucket Name: %s\n", bucketName)
	fmt.Printf("   Distribution ID: %s\n", *distribution.Id)
	fmt.Printf("   Distribution Domain Name: %s\n", *distribution.DomainName)
	fmt.Printf("\n⏳ Note: It may take up to 15 minutes for the CloudFront distribution to be fully deployed\n")
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

	fileCount := 0
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
		fmt.Printf("   Uploading: %s\n", relPath)
		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(bucketName),
			Key:         aws.String(relPath),
			Body:        bytes.NewReader(content),
			CacheControl: aws.String("public, max-age=31536000, s-maxage=31536000"),
		})
		if err != nil {
			return fmt.Errorf("failed to upload file %s: %w", relPath, err)
		}

		fileCount++
		return nil
	})

	if err == nil {
		fmt.Printf("   Total files uploaded: %d\n", fileCount)
	}

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