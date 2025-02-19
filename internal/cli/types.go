package cli

// Config holds all the configuration options for the deployment
type Config struct {
	DeploymentType        string
	ProfileName           string
	BucketName            string
	WebsiteFolderPath     string
	CloudFrontDescription string
	Region                string
}

// DeploymentType constants
const (
	StaticWebsite = "1"
	SPA          = "2"
)