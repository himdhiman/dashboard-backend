package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/supabase/config"
	"github.com/himdhiman/dashboard-backend/libs/supabase/errors"
	"github.com/himdhiman/dashboard-backend/libs/supabase/models"
)

// IStorageClient defines the interface for storage operations
type IStorageClient interface {
	// Bucket operations
	CreateBucket(ctx context.Context, name string, options map[string]interface{}) (*models.StorageBucket, error)
	GetBucket(ctx context.Context, name string) (*models.StorageBucket, error)
	ListBuckets(ctx context.Context) ([]*models.StorageBucket, error)
	DeleteBucket(ctx context.Context, name string) error
	EmptyBucket(ctx context.Context, name string) error

	// File operations
	Upload(ctx context.Context, bucket, path string, file interface{}, options map[string]interface{}) (*models.StorageFile, error)
	UploadFromFile(ctx context.Context, bucket, path, filePath string, options map[string]interface{}) (*models.StorageFile, error)
	Download(ctx context.Context, bucket, path string) ([]byte, error)
	DownloadToFile(ctx context.Context, bucket, path, destination string) error
	Delete(ctx context.Context, bucket string, paths []string) error
	Move(ctx context.Context, bucket, fromPath, toPath string) error
	Copy(ctx context.Context, bucket, fromPath, toPath string) error

	// File listing and info
	List(ctx context.Context, bucket, path string, options map[string]interface{}) ([]*models.StorageFile, error)
	GetPublicURL(bucket, path string) string
	CreateSignedURL(ctx context.Context, bucket, path string, expiresIn int, options map[string]interface{}) (string, error)

	// File operations with bucket context
	From(bucket string) IStorageBucket
}

// IStorageBucket defines the interface for bucket-specific operations
type IStorageBucket interface {
	Upload(ctx context.Context, path string, file interface{}, options map[string]interface{}) (*models.StorageFile, error)
	UploadFromFile(ctx context.Context, path, filePath string, options map[string]interface{}) (*models.StorageFile, error)
	Download(ctx context.Context, path string) ([]byte, error)
	DownloadToFile(ctx context.Context, path, destination string) error
	Delete(ctx context.Context, paths []string) error
	Move(ctx context.Context, fromPath, toPath string) error
	Copy(ctx context.Context, fromPath, toPath string) error
	List(ctx context.Context, path string, options map[string]interface{}) ([]*models.StorageFile, error)
	GetPublicURL(path string) string
	CreateSignedURL(ctx context.Context, path string, expiresIn int, options map[string]interface{}) (string, error)
}

// StorageClient represents a storage client
type StorageClient struct {
	IStorageClient
	config     *config.Config
	httpClient *http.Client
	logger     logger.ILogger
}

// StorageBucket represents a storage bucket
type StorageBucket struct {
	IStorageBucket
	client *StorageClient
	bucket string
}

// NewStorageClient creates a new storage client
func NewStorageClient(cfg *config.Config, httpClient *http.Client, logger logger.ILogger) (IStorageClient, error) {
	return &StorageClient{
		config:     cfg,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// CreateBucket creates a new storage bucket
func (s *StorageClient) CreateBucket(ctx context.Context, name string, options map[string]interface{}) (*models.StorageBucket, error) {
	url := s.config.URL + "/storage/v1/bucket"

	params := map[string]interface{}{
		"id": name,
	}

	// Merge options with default params
	for k, v := range options {
		params[k] = v
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, errors.NewError(err, "failed to marshal create bucket request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewError(err, "failed to create bucket request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	// Add options as headers
	for k, v := range options {
		httpReq.Header.Set(k, fmt.Sprintf("%v", v))
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute create bucket request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("create bucket failed with status: %d", resp.StatusCode)
	}

	var bucket models.StorageBucket
	if err := json.NewDecoder(resp.Body).Decode(&bucket); err != nil {
		return nil, errors.NewError(err, "failed to decode create bucket response")
	}

	s.logger.Info("Bucket created successfully", "bucket", name)
	return &bucket, nil
}

// GetBucket retrieves a storage bucket
func (s *StorageClient) GetBucket(ctx context.Context, name string) (*models.StorageBucket, error) {
	url := s.config.URL + "/storage/v1/bucket/" + name

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewError(err, "failed to create get bucket request")
	}

	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute get bucket request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("get bucket failed with status: %d", resp.StatusCode)
	}

	var bucket models.StorageBucket
	if err := json.NewDecoder(resp.Body).Decode(&bucket); err != nil {
		return nil, errors.NewError(err, "failed to decode get bucket response")
	}

	return &bucket, nil
}

// ListBuckets lists all storage buckets
func (s *StorageClient) ListBuckets(ctx context.Context) ([]*models.StorageBucket, error) {
	url := s.config.URL + "/storage/v1/bucket"

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewError(err, "failed to create list buckets request")
	}

	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute list buckets request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("list buckets failed with status: %d", resp.StatusCode)
	}

	var buckets []*models.StorageBucket
	if err := json.NewDecoder(resp.Body).Decode(&buckets); err != nil {
		return nil, errors.NewError(err, "failed to decode list buckets response")
	}

	return buckets, nil
}

// DeleteBucket deletes a storage bucket
func (s *StorageClient) DeleteBucket(ctx context.Context, name string) error {
	url := s.config.URL + "/storage/v1/bucket/" + name

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return errors.NewError(err, "failed to create delete bucket request")
	}

	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute delete bucket request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("delete bucket failed with status: %d", resp.StatusCode)
	}

	s.logger.Info("Bucket deleted successfully", "bucket", name)
	return nil
}

// EmptyBucket empties a storage bucket
func (s *StorageClient) EmptyBucket(ctx context.Context, name string) error {
	url := s.config.URL + "/storage/v1/bucket/" + name + "/empty"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return errors.NewError(err, "failed to create empty bucket request")
	}

	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute empty bucket request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("empty bucket failed with status: %d", resp.StatusCode)
	}

	s.logger.Info("Bucket emptied successfully", "bucket", name)
	return nil
}

// Upload uploads a file to storage
func (s *StorageClient) Upload(ctx context.Context, bucket, path string, file interface{}, options map[string]interface{}) (*models.StorageFile, error) {
	url := s.config.URL + "/storage/v1/object/" + bucket + "/" + path

	var body io.Reader
	var contentType string

	switch f := file.(type) {
	case []byte:
		body = bytes.NewBuffer(f)
		contentType = "application/octet-stream"
	case string:
		body = strings.NewReader(f)
		contentType = "text/plain"
	case io.Reader:
		body = f
		contentType = "application/octet-stream"
	default:
		return nil, errors.NewError(errors.ErrInvalidFileType, "unsupported file type")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, errors.NewError(err, "failed to create upload request")
	}

	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	// Add options as headers
	for k, v := range options {
		httpReq.Header.Set(k, fmt.Sprintf("%v", v))
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute upload request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("upload failed with status: %d", resp.StatusCode)
	}

	var fileInfo models.StorageFile
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		return nil, errors.NewError(err, "failed to decode upload response")
	}

	s.logger.Info("File uploaded successfully", "bucket", bucket, "path", path)
	return &fileInfo, nil
}

// UploadFromFile uploads a file from the local filesystem
func (s *StorageClient) UploadFromFile(ctx context.Context, bucket, path, filePath string, options map[string]interface{}) (*models.StorageFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.NewError(err, "failed to open file")
	}
	defer file.Close()

	return s.Upload(ctx, bucket, path, file, options)
}

// Download downloads a file from storage
func (s *StorageClient) Download(ctx context.Context, bucket, path string) ([]byte, error) {
	url := s.config.URL + "/storage/v1/object/" + bucket + "/" + path

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewError(err, "failed to create download request")
	}

	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute download request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("download failed with status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewError(err, "failed to read download response")
	}

	s.logger.Info("File downloaded successfully", "bucket", bucket, "path", path)
	return data, nil
}

// DownloadToFile downloads a file to the local filesystem
func (s *StorageClient) DownloadToFile(ctx context.Context, bucket, path, destination string) error {
	data, err := s.Download(ctx, bucket, path)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewError(err, "failed to create destination directory")
	}

	if err := os.WriteFile(destination, data, 0644); err != nil {
		return errors.NewError(err, "failed to write file")
	}

	s.logger.Info("File downloaded to local filesystem", "bucket", bucket, "path", path, "destination", destination)
	return nil
}

// Delete deletes files from storage
func (s *StorageClient) Delete(ctx context.Context, bucket string, paths []string) error {
	url := s.config.URL + "/storage/v1/object/" + bucket

	body, err := json.Marshal(map[string][]string{"prefixes": paths})
	if err != nil {
		return errors.NewError(err, "failed to marshal delete request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.NewError(err, "failed to create delete request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute delete request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("delete failed with status: %d", resp.StatusCode)
	}

	s.logger.Info("Files deleted successfully", "bucket", bucket, "paths", paths)
	return nil
}

// Move moves a file in storage
func (s *StorageClient) Move(ctx context.Context, bucket, fromPath, toPath string) error {
	url := s.config.URL + "/storage/v1/object/move"

	body, err := json.Marshal(map[string]string{
		"bucketId":       bucket,
		"sourceKey":      fromPath,
		"destinationKey": toPath,
	})
	if err != nil {
		return errors.NewError(err, "failed to marshal move request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.NewError(err, "failed to create move request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute move request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("move failed with status: %d", resp.StatusCode)
	}

	s.logger.Info("File moved successfully", "bucket", bucket, "from", fromPath, "to", toPath)
	return nil
}

// Copy copies a file in storage
func (s *StorageClient) Copy(ctx context.Context, bucket, fromPath, toPath string) error {
	url := s.config.URL + "/storage/v1/object/copy"

	body, err := json.Marshal(map[string]string{
		"bucketId":       bucket,
		"sourceKey":      fromPath,
		"destinationKey": toPath,
	})
	if err != nil {
		return errors.NewError(err, "failed to marshal copy request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.NewError(err, "failed to create copy request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute copy request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("copy failed with status: %d", resp.StatusCode)
	}

	s.logger.Info("File copied successfully", "bucket", bucket, "from", fromPath, "to", toPath)
	return nil
}

// List lists files in a bucket
func (s *StorageClient) List(ctx context.Context, bucket, path string, options map[string]interface{}) ([]*models.StorageFile, error) {
	url := s.config.URL + "/storage/v1/object/list/" + bucket

	// Add path as query parameter
	if path != "" {
		url += "?prefix=" + path
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewError(err, "failed to create list request")
	}

	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute list request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("list failed with status: %d", resp.StatusCode)
	}

	var files []*models.StorageFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, errors.NewError(err, "failed to decode list response")
	}

	return files, nil
}

// GetPublicURL returns the public URL for a file
func (s *StorageClient) GetPublicURL(bucket, path string) string {
	return s.config.URL + "/storage/v1/object/public/" + bucket + "/" + path
}

// CreateSignedURL creates a signed URL for a file
func (s *StorageClient) CreateSignedURL(ctx context.Context, bucket, path string, expiresIn int, options map[string]interface{}) (string, error) {
	url := s.config.URL + "/storage/v1/object/sign/" + bucket + "/" + path

	params := map[string]interface{}{
		"expiresIn": expiresIn,
	}

	// Merge options with default params
	for k, v := range options {
		params[k] = v
	}

	body, err := json.Marshal(params)
	if err != nil {
		return "", errors.NewError(err, "failed to marshal signed URL request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", errors.NewError(err, "failed to create signed URL request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+s.config.Key)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", errors.NewError(err, "failed to execute signed URL request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.NewErrorf("signed URL creation failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.NewError(err, "failed to decode signed URL response")
	}

	signedURL, ok := result["signedURL"].(string)
	if !ok {
		return "", errors.NewError(err, "invalid signed URL response format")
	}

	return signedURL, nil
}

// From returns a bucket-specific storage client
func (s *StorageClient) From(bucket string) IStorageBucket {
	return &StorageBucket{
		client: s,
		bucket: bucket,
	}
}

// StorageBucket methods
func (sb *StorageBucket) Upload(ctx context.Context, path string, file interface{}, options map[string]interface{}) (*models.StorageFile, error) {
	return sb.client.Upload(ctx, sb.bucket, path, file, options)
}

func (sb *StorageBucket) UploadFromFile(ctx context.Context, path, filePath string, options map[string]interface{}) (*models.StorageFile, error) {
	return sb.client.UploadFromFile(ctx, sb.bucket, path, filePath, options)
}

func (sb *StorageBucket) Download(ctx context.Context, path string) ([]byte, error) {
	return sb.client.Download(ctx, sb.bucket, path)
}

func (sb *StorageBucket) DownloadToFile(ctx context.Context, path, destination string) error {
	return sb.client.DownloadToFile(ctx, sb.bucket, path, destination)
}

func (sb *StorageBucket) Delete(ctx context.Context, paths []string) error {
	return sb.client.Delete(ctx, sb.bucket, paths)
}

func (sb *StorageBucket) Move(ctx context.Context, fromPath, toPath string) error {
	return sb.client.Move(ctx, sb.bucket, fromPath, toPath)
}

func (sb *StorageBucket) Copy(ctx context.Context, fromPath, toPath string) error {
	return sb.client.Copy(ctx, sb.bucket, fromPath, toPath)
}

func (sb *StorageBucket) List(ctx context.Context, path string, options map[string]interface{}) ([]*models.StorageFile, error) {
	return sb.client.List(ctx, sb.bucket, path, options)
}

func (sb *StorageBucket) GetPublicURL(path string) string {
	return sb.client.GetPublicURL(sb.bucket, path)
}

func (sb *StorageBucket) CreateSignedURL(ctx context.Context, path string, expiresIn int, options map[string]interface{}) (string, error) {
	return sb.client.CreateSignedURL(ctx, sb.bucket, path, expiresIn, options)
}
