package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// IAMService handles IAM-related operations
type IAMService struct {
	client *iam.Client
}

// NewIAMService creates a new IAMService instance
func NewIAMService(client *iam.Client) *IAMService {
	return &IAMService{
		client: client,
	}
}

// GetAWSAccountID retrieves the AWS account ID
func (i *IAMService) GetAWSAccountID(ctx context.Context) (string, error) {
	result, err := i.client.GetUser(ctx, &iam.GetUserInput{})
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