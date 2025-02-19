package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Service handles S3-related operations
type S3Service struct {
	client     *s3.Client
	bucketName string
	region     string
}

// NewS3Service creates a new S3Service instance
func NewS3Service(client *s3.Client, bucketName, region string) *S3Service {
	return &S3Service{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}
}

// CreateBucket creates a new S3 bucket with the specified configuration
func (s *S3Service) CreateBucket(ctx context.Context) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName),
	}

	if s.region != "us-east-1" {
		input.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(s.region),
		}
	}

	_, err := s.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	// Block public access
	_, err = s.client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(s.bucketName),
		PublicAccessBlockConfiguration: &s3Types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to block public access: %w", err)
	}

	return nil
}

// AttachBucketPolicy attaches the CloudFront access policy to the bucket
func (s *S3Service) AttachBucketPolicy(ctx context.Context, accountID, distributionID string) error {
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
	}`, s.bucketName, accountID, distributionID)

	_, err := s.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(s.bucketName),
		Policy: aws.String(policyDocument),
	})
	return err
}