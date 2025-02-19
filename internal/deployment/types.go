package deployment

import (
	"aws-deploy-static-site/internal/aws"
	"aws-deploy-static-site/internal/cli"
)

// Service handles the deployment process
type Service struct {
	config         *cli.Config
	awsClients     *aws.Clients
	s3Service      *aws.S3Service
	cfService      *aws.CloudFrontService
	iamService     *aws.IAMService
	distribution   *string
	distributionID *string
}

// DeploymentSummary contains the final deployment information
type DeploymentSummary struct {
	BucketName         string
	DistributionID     string
	DistributionDomain string
}