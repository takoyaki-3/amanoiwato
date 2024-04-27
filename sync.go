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

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

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

func main() {
	loadEnv()

	bucket := os.Getenv("S3_BUCKET")
	storageDir := "storage"
	localCSVPath := "file_list.csv"
	remoteCSVPath := "file_list.csv"

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

	err = downloadFile(sess, bucket, remoteCSVPath, localCSVPath)
	if err != nil {
		log.Println("No existing CSV file on S3, or unable to download:", err)
	}

	existingFiles, err := readCSV(localCSVPath)
	if err != nil {
		log.Println("Failed to read existing CSV:", err)
		existingFiles = make(map[string][2]string)
	}

	// Create S3 client
	s3Client := s3.New(sess)

	// List all files in the S3 bucket
	err = s3Client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, object := range page.Contents {
			key := *object.Key
			// Convert the key to the local file path format
			localPath := filepath.FromSlash(key)  // Converts S3 key to OS-specific path format
			if _, err := os.Stat(localPath); os.IsNotExist(err) {
				fmt.Printf("File %s does not exist locally. Downloading...\n", key)
				if err := downloadFile(sess, bucket, key, localPath); err != nil {
					log.Printf("Failed to download %s: %v\n", key, err)
				}
			}
		}
		return !lastPage
	})
	if err != nil {
		log.Printf("Failed to list objects in bucket %s: %v\n", bucket, err)
	}

	func(){
		csvFile, err := os.Create(localCSVPath)
		if err != nil {
			log.Fatal("Error creating CSV file", err)
		}
		defer csvFile.Close()
	
		writer := csv.NewWriter(csvFile)
		defer writer.Flush()
		writer.Write([]string{"Path", "Hash", "Last Modified"})
	
		err = filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				modTime := info.ModTime().Format(time.RFC3339)
				previousData, exists := existingFiles[path]
				if !exists || previousData[1] != modTime {
					hash := hashFile(path)
					writer.Write([]string{path, hash, modTime})
					relativePath := strings.TrimPrefix(path, storageDir+"/")
					uploadFile(sess, bucket, path, relativePath)
				} else {
					writer.Write([]string{path, previousData[0], modTime})
					// fmt.Printf("Skipping upload, no changes for %s\n", path)
				}
			}
			return nil
		})
		if err != nil {
			log.Fatal("Error walking through storage directory", err)
		}
	
		if err := writer.Error(); err != nil {
			log.Fatal("Error writing CSV", err)
		}	
	}()

	uploadFile(sess, bucket, localCSVPath, remoteCSVPath)
}
