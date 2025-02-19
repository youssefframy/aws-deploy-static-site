package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// CloudFrontService handles CloudFront-related operations
type CloudFrontService struct {
	client      *cloudfront.Client
	bucketName  string
	region      string
	description string
	deployType  string
}

// NewCloudFrontService creates a new CloudFrontService instance
func NewCloudFrontService(client *cloudfront.Client, bucketName, region, description, deployType string) *CloudFrontService {
	return &CloudFrontService{
		client:      client,
		bucketName:  bucketName,
		region:      region,
		description: description,
		deployType:  deployType,
	}
}

// CreateOriginAccessControl creates a new Origin Access Control
func (cf *CloudFrontService) CreateOriginAccessControl(ctx context.Context) (*string, error) {
	resp, err := cf.client.CreateOriginAccessControl(ctx, &cloudfront.CreateOriginAccessControlInput{
		OriginAccessControlConfig: &types.OriginAccessControlConfig{
			Name:                          aws.String(fmt.Sprintf("%s-origin-access-control", cf.bucketName)),
			OriginAccessControlOriginType: "s3",
			SigningBehavior:              "always",
			SigningProtocol:              "sigv4",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create origin access control: %w", err)
	}

	return resp.OriginAccessControl.Id, nil
}

// CreateDistribution creates a new CloudFront distribution
func (cf *CloudFrontService) CreateDistribution(ctx context.Context, oacId *string) (*types.Distribution, error) {
	distConfig := &types.DistributionConfig{
		CallerReference: aws.String(fmt.Sprintf("cli-%d", time.Now().Unix())),
		Comment:         aws.String(cf.description),
		Enabled:        aws.Bool(true),
		IsIPV6Enabled:  aws.Bool(true),
		Origins: &types.Origins{
			Items: []types.Origin{
				{
					DomainName:            aws.String(fmt.Sprintf("%s.s3.%s.amazonaws.com", cf.bucketName, cf.region)),
					Id:                    aws.String("S3Origin"),
					OriginAccessControlId: oacId,
					S3OriginConfig:        &types.S3OriginConfig{
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
	}

	if cf.deployType == "1" {
		distConfig.DefaultRootObject = aws.String("index.html")
	}

	// Add custom error responses for SPA
	if cf.deployType == "2" {
		distConfig.CustomErrorResponses = &types.CustomErrorResponses{
			Items: []types.CustomErrorResponse{
				{
					ErrorCode:          aws.Int32(403),
					ResponseCode:       aws.String("200"),
					ResponsePagePath:   aws.String("/index.html"),
					ErrorCachingMinTTL: aws.Int64(15),
				},
				{
					ErrorCode:          aws.Int32(404),
					ResponseCode:       aws.String("200"),
					ResponsePagePath:   aws.String("/index.html"),
					ErrorCachingMinTTL: aws.Int64(15),
				},
			},
			Quantity: aws.Int32(2),
		}
	}

	resp, err := cf.client.CreateDistribution(ctx, &cloudfront.CreateDistributionInput{
		DistributionConfig: distConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create distribution: %w", err)
	}

	return resp.Distribution, nil
}