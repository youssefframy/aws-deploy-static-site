package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Clients holds all AWS service clients
type Clients struct {
	S3         *s3.Client
	CloudFront *cloudfront.Client
	IAM        *iam.Client
}

// NewClients creates new AWS service clients using the provided configuration
func NewClients(ctx context.Context, region, profile string) (*Clients, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, err
	}

	return &Clients{
		S3:         s3.NewFromConfig(cfg),
		CloudFront: cloudfront.NewFromConfig(cfg),
		IAM:        iam.NewFromConfig(cfg),
	}, nil
}