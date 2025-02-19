# AWS Static Site Deployment Tool

A powerful CLI tool that simplifies the deployment of static websites and Single Page Applications (SPAs) to AWS using S3 and CloudFront.

![Deploy SPA Demo](deploy-spa-demo.gif)

## Features

- 🚀 One-command deployment process
- 🔒 Secure by default with CloudFront Origin Access Control
- 📱 Support for both static websites and SPAs
- 💨 Optimal caching configuration
- 🌐 HTTPS enabled by default
- 🎯 Interactive CLI interface

## Quick Start

### Option 1: Pre-built Binary

1. Download the latest release for your platform:

   - [Windows (64-bit)](https://github.com/youssefframy/aws-deploy-static-site/releases/latest/download/aws-deploy-win-x64.exe)
   - [macOS (64-bit)](https://github.com/youssefframy/aws-deploy-static-site/releases/latest/download/aws-deploy-macos-x64)
   - [Linux (64-bit)](https://github.com/youssefframy/aws-deploy-static-site/releases/latest/download/aws-deploy-linux-x64)

2. Configure AWS credentials using one of these methods:

   - AWS CLI: `aws configure`
   - Environment variables:
     ```bash
     export AWS_ACCESS_KEY_ID="your_access_key"
     export AWS_SECRET_ACCESS_KEY="your_secret_key"
     ```
   - IAM role (if running on AWS infrastructure)

3. Run the executable and follow the interactive prompts

### Option 2: Build from Source

#### Prerequisites

- Go 1.24 or later
- AWS credentials configured
- Required AWS permissions:
  - S3: CreateBucket, PutObject, PutBucketPolicy
  - CloudFront: CreateDistribution, CreateOriginAccessControl
  - IAM: GetUser

#### Installation

1. Clone the repository:

```bash
git clone https://github.com/youssefframy/aws-deploy-static-site.git
cd aws-deploy-static-site
```

2. Build the project:

```bash
go build -o aws-deploy ./cmd/aws-deploy
```

3. Run the tool:

```bash
./aws-deploy
```

## Usage Guide

The tool provides an interactive CLI interface that will guide you through the deployment process:

1. Select deployment type:

   - Static Website (Basic)
   - Single Page Application (SPA)

2. Configure deployment settings:

   - AWS Profile (optional)
   - S3 Bucket Name
   - Website Files Location
   - CloudFront Distribution Description
   - AWS Region

3. Wait for deployment completion (typically 10-15 minutes for CloudFront propagation)

## Architecture

The deployment process:

1. Creates a private S3 bucket
2. Configures website hosting settings
3. Uploads your static files with optimal caching headers
4. Creates a CloudFront distribution with Origin Access Control
5. Configures security policies and routing rules

## Security Features

- ✅ Private S3 bucket with public access blocked
- ✅ CloudFront Origin Access Control (OAC)
- ✅ HTTPS-only access
- ✅ Secure IAM policies
- ✅ Custom error handling for SPAs

## Troubleshooting

Common issues and solutions:

1. **Access Denied**

   - Verify AWS credentials are configured correctly
   - Ensure IAM user has required permissions

2. **Bucket Creation Failed**

   - Check if bucket name is globally unique
   - Verify selected region supports all services

3. **Upload Issues**

   - Confirm website folder path is correct
   - Check file permissions

4. **CloudFront Errors**
   - Allow 10-15 minutes for distribution deployment
   - Verify domain name resolution

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For bugs and feature requests, please [open an issue](https://github.com/youssefframy/aws-deploy-static-site/issues).
