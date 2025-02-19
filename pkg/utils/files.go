package utils

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// getContentType determines the content type based on file extension
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	
	// Add common web content types
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".txt":
		return "text/plain"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	}

	// For unknown extensions, try to detect using mime package
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

// UploadWebsiteFiles uploads all files from the specified directory to S3
func UploadWebsiteFiles(ctx context.Context, client *s3.Client, bucketName, websitePath string) error {
	absPath, err := filepath.Abs(websitePath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	fileCount := 0
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(absPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		key := filepath.ToSlash(relPath)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		contentType := getContentType(path)
		fmt.Printf("   Uploading: %s (Content-Type: %s)\n", key, contentType)
		
		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(bucketName),
			Key:          aws.String(key),
			Body:         bytes.NewReader(content),
			ContentType:  aws.String(contentType),
			CacheControl: aws.String("public, max-age=31536000, s-maxage=31536000"),
		})
		if err != nil {
			return fmt.Errorf("failed to upload file %s: %w", key, err)
		}

		fileCount++
		return nil
	})

	if err == nil {
		fmt.Printf("   Total files uploaded: %d\n", fileCount)
	}

	return err
}