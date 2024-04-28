package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/joho/godotenv"
)

// Load environment variables
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Hash a file and return the hash as a string
func hashFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Download a file from an S3 bucket
func downloadFile(sess *session.Session, bucket, key, destPath string) error {
	directory := filepath.Dir(destPath)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, os.ModePerm)
	}

	downloader := s3manager.NewDownloader(sess)
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Successfully downloaded %s from %s\n", key, bucket)
	return nil
}

// Read the CSV file and return a map of file paths to their hash and last modified time
func readCSV(filePath string) (map[string][2]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	fileMap := make(map[string][2]string) // Map of path to hash and last modified
	for _, record := range records {
		if len(record) == 3 {
			fileMap[record[0]] = [2]string{record[1], record[2]}
		}
	}
	return fileMap, nil
}

// Upload a file to an S3 bucket
func uploadFile(sess *session.Session, bucket, filePath, key string) {
	uploader := s3manager.NewUploader(sess)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	linuxKey := strings.ReplaceAll(key, "\\", "/")

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(linuxKey),
		Body:   file,
	})
	if err != nil {
		log.Fatalf("Unable to upload %s to %s, %v", filePath, bucket, err)
	}
	fmt.Printf("Successfully uploaded %s to %s with key %s\n", filePath, bucket, linuxKey)
}

// Main execution function
func main() {
	loadEnv()

	bucket := os.Getenv("S3_BUCKET")
	localCSVPath := "remote_file_list.csv"
	remoteCSVPath := "file_list.csv"
	storageDir := "storage"

	// Initialize AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(os.Getenv("AWS_REGION")),
		Endpoint:         aws.String(os.Getenv("S3_ENDPOINT")),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			""),
	})
	if err != nil {
		log.Fatal("Error creating AWS session", err)
	}

	// Download the CSV file from the S3 bucket
	err = downloadFile(sess, bucket, remoteCSVPath, localCSVPath)
	if err != nil {
		log.Println("No existing CSV file on S3, or unable to download:", err)
	}

	// Read the existing CSV file to get current file data
	existingFiles, err := readCSV(localCSVPath)
	if err != nil {
		log.Println("Failed to read existing CSV:", err)
		existingFiles = make(map[string][2]string)
	}

	// Walk through the local storage directory
	err = filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			hash := hashFile(path)
			modTime := info.ModTime().Format(time.RFC3339)
			previousData, exists := existingFiles[path]
			if !exists || previousData[0] != hash || previousData[1] != modTime {
				fmt.Printf("Uploading updated or new file: %s\n", path)
				relativePath := strings.TrimPrefix(path, storageDir+"/")
				uploadFile(sess, bucket, path, relativePath)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal("Error walking through storage directory", err)
	}

	// Upload the updated CSV file back to the S3 bucket
	uploadFile(sess, bucket, localCSVPath, remoteCSVPath)

	os.Remove(localCSVPath)
}
