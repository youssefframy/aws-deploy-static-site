package deployment

import (
	"context"
	"fmt"

	"aws-deploy-static-site/internal/aws"
	"aws-deploy-static-site/internal/cli"
	"aws-deploy-static-site/pkg/utils"
)

// NewService creates a new deployment service
func NewService(ctx context.Context, config *cli.Config) (*Service, error) {
	// Initialize AWS clients
	awsClients, err := aws.NewClients(ctx, config.Region, config.ProfileName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS clients: %w", err)
	}

	// Initialize service components
	s3Service := aws.NewS3Service(awsClients.S3, config.BucketName, config.Region)
	cfService := aws.NewCloudFrontService(awsClients.CloudFront, config.BucketName, config.Region, config.CloudFrontDescription, config.DeploymentType)
	iamService := aws.NewIAMService(awsClients.IAM)

	return &Service{
		config:     config,
		awsClients: awsClients,
		s3Service:  s3Service,
		cfService:  cfService,
		iamService: iamService,
	}, nil
}

// Deploy executes the deployment process
func (s *Service) Deploy(ctx context.Context) error {
	// 1. Create S3 bucket
	fmt.Printf("📦 Creating S3 bucket '%s'...\n", s.config.BucketName)
	err := s.s3Service.CreateBucket(ctx)
	if err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}
	fmt.Printf("✅ S3 bucket created successfully\n\n")

	// 2. Upload website files
	fmt.Printf("📤 Uploading website files...\n")
	err = utils.UploadWebsiteFiles(ctx, s.awsClients.S3, s.config.BucketName, s.config.WebsiteFolderPath)
	if err != nil {
		return fmt.Errorf("failed to upload website files: %w", err)
	}
	fmt.Printf("✅ Website files uploaded successfully\n\n")

	// 3. Create Origin Access Control
	fmt.Printf("🔒 Creating Origin Access Control...\n")
	oacId, err := s.cfService.CreateOriginAccessControl(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Origin Access Control: %w", err)
	}
	fmt.Printf("✅ Origin Access Control created successfully\n\n")

	// 4. Create CloudFront distribution
	fmt.Printf("☁️  Creating CloudFront distribution...\n")
	distribution, err := s.cfService.CreateDistribution(ctx, oacId)
	if err != nil {
		return fmt.Errorf("failed to create CloudFront distribution: %w", err)
	}
	s.distribution = distribution.DomainName
	s.distributionID = distribution.Id
	fmt.Printf("✅ CloudFront distribution created successfully\n\n")

	// 5. Get AWS account ID and attach bucket policy
	fmt.Printf("📜 Attaching bucket policy...\n")
	accountID, err := s.iamService.GetAWSAccountID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get AWS account ID: %w", err)
	}

	err = s.s3Service.AttachBucketPolicy(ctx, accountID, *s.distributionID)
	if err != nil {
		return fmt.Errorf("failed to attach bucket policy: %w", err)
	}
	fmt.Printf("✅ Bucket policy attached successfully\n\n")

	return nil
}

// DisplaySummary shows the deployment summary
func (s *Service) DisplaySummary() {
	fmt.Printf("🎉 Deployment completed successfully!\n")
	fmt.Printf("📋 Summary:\n")
	fmt.Printf("   Bucket Name: %s\n", s.config.BucketName)
	fmt.Printf("   Distribution ID: %s\n", *s.distributionID)
	fmt.Printf("   Distribution Domain Name: https://%s\n", *s.distribution)
	fmt.Printf("\n⏳ Note: It may take up to 15 minutes for the CloudFront distribution to be fully deployed\n")
}