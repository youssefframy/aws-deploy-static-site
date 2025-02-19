package cli

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

// GetUserInput collects all necessary configuration from user input
func GetUserInput() (*Config, error) {
	config := &Config{}

	// Deployment type selection
	deployType, err := getDeploymentType()
	if err != nil {
		return nil, err
	}
	config.DeploymentType = deployType

	// AWS Profile
	profile, err := getAWSProfile()
	if err != nil {
		return nil, err
	}
	config.ProfileName = profile

	// Bucket name
	bucket, err := getBucketName()
	if err != nil {
		return nil, err
	}
	config.BucketName = bucket

	// Website folder path
	path, err := getWebsitePath()
	if err != nil {
		return nil, err
	}
	config.WebsiteFolderPath = path

	// CloudFront description
	desc, err := getCloudFrontDescription(bucket)
	if err != nil {
		return nil, err
	}
	config.CloudFrontDescription = desc

	// Region
	region, err := getRegion()
	if err != nil {
		return nil, err
	}
	config.Region = region

	return config, nil
}

func getDeploymentType() (string, error) {
	prompt := promptui.Select{
		Label: "Select deployment type",
		Items: []string{
			"Static Website (Basic)",
			"Single Page Application (SPA)",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . | cyan }}?",
			Active:   "➜ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✔ {{ . | green }}",
		},
		HideHelp: true,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	return fmt.Sprintf("%d", index+1), nil
}

func getAWSProfile() (string, error) {
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | cyan }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "",
	}

	prompt := promptui.Prompt{
		Label:     "🔑 AWS Profile",
		Default:   "default",
		Templates: templates,
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("profile name cannot be empty")
			}
			return nil
		},
	}

	return prompt.Run()
}

func getBucketName() (string, error) {
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | cyan }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "",
	}

	prompt := promptui.Prompt{
		Label:     "🪣 S3 Bucket Name",
		Templates: templates,
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("bucket name cannot be empty")
			}
			return nil
		},
	}

	return prompt.Run()
}

func getWebsitePath() (string, error) {
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | cyan }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "",
	}

	prompt := promptui.Prompt{
		Label:     "📂 Website Folder Path",
		Templates: templates,
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("folder path cannot be empty")
			}
			return nil
		},
	}

	return prompt.Run()
}

func getCloudFrontDescription(bucketName string) (string, error) {
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | cyan }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "",
	}

	prompt := promptui.Prompt{
		Label:     "💬 CloudFront Distribution Description",
		Default:   fmt.Sprintf("Distribution for %s", bucketName),
		Templates: templates,
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("description cannot be empty")
			}
			return nil
		},
	}

	return prompt.Run()
}

func getRegion() (string, error) {
	prompt := promptui.Select{
		Label: "🌐 Select AWS Region",
		Items: []string{
			"us-east-1", "us-east-2", "us-west-1", "us-west-2",
			"eu-west-1", "eu-west-2", "eu-central-1",
			"ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . | cyan }}?",
			Active:   "➜ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✔ {{ . | green }}",
		},
		HideHelp: true,
	}

	_, result, err := prompt.Run()
	return result, err
}

// Additional prompt functions...

func DisplayConfigSummary(config *Config) {
	fmt.Println("--------------------------------")
	fmt.Printf("\n%s\n", "Configuration Summary:")
	fmt.Printf("⚙️ Deployment Type: %s\n", config.DeploymentType)
	fmt.Printf("🔑 AWS Profile: %s\n", config.ProfileName)
	fmt.Printf("🪣 S3 Bucket: %s\n", config.BucketName)
	fmt.Printf("📂 Website Path: %s\n", config.WebsiteFolderPath)
	fmt.Printf("💬 CloudFront Description: %s\n", config.CloudFrontDescription)
	fmt.Printf("🌐 Region: %s\n\n", config.Region)
	fmt.Println("--------------------------------")
}