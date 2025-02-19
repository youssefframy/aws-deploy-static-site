package main

import (
	"context"
	"fmt"
	"log"

	"aws-deploy-static-site/internal/cli"
	"aws-deploy-static-site/internal/deployment"
)

func main() {
	fmt.Println("🚀 Welcome to the CloudFront Deployment Tool!")

	// Get configuration from CLI prompts
	config, err := cli.GetUserInput()
	if err != nil {
		log.Fatalf("Error getting user input: %v", err)
	}

	// Display configuration summary
	cli.DisplayConfigSummary(config)

	// Create deployment service
	deploymentService, err := deployment.NewService(context.TODO(), config)
	if err != nil {
		log.Fatalf("Error creating deployment service: %v", err)
	}

	// Run deployment
	err = deploymentService.Deploy(context.TODO())
	if err != nil {
		log.Fatalf("Deployment failed: %v", err)
	}

	// Display success message
	deploymentService.DisplaySummary()
}