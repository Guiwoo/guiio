package guiio_service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	guiio_repository "guiio/guiio_repository"
	guiio_http "guiio/guiio_server/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sphynx/config"
)

type StorageService struct {
	client        StorageClient
	defaultRegion string
	repo          guiio_repository.ObjectRepository
}

type StorageClient interface {
	ListBuckets(ctx context.Context) ([]minio.BucketInfo, error)
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	RemoveBucket(ctx context.Context, bucketName string) error
}

type bucketInfo struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type createBucketRequest struct {
	Name   string `json:"name"`
	Region string `json:"region,omitempty"`
}

func NewStorageService(repo guiio_repository.ObjectRepository) (*StorageService, error) {
	endpoint := strings.TrimSpace(config.Get[string]("storage_endpoint"))
	accessKey := strings.TrimSpace(config.Get[string]("storage_access_key"))
	secretKey := strings.TrimSpace(config.Get[string]("storage_secret_key"))
	useSSL := config.Get[bool]("storage_use_ssl")
	region := strings.TrimSpace(config.Get[string]("storage_region"))

	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("storage configuration is missing")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage client: %w", err)
	}

	return NewStorageServiceWithClient(client, region, repo), nil
}

func NewStorageServiceWithClient(client StorageClient, defaultRegion string, repo guiio_repository.ObjectRepository) *StorageService {
	return &StorageService{
		client:        client,
		defaultRegion: defaultRegion,
		repo:          repo,
	}
}

func (s *StorageService) ListBucket(ctx guiio_http.Context) error {
	buckets, err := s.client.ListBuckets(ctx.Context())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("list buckets failed: %v", err),
		})
	}

	result := make([]bucketInfo, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, bucketInfo{
			Name:      b.Name,
			CreatedAt: b.CreationDate,
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"buckets": result,
	})
}

func (s *StorageService) CreateBucket(ctx guiio_http.Context) {
	var req createBucketRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Region = strings.TrimSpace(req.Region)

	if req.Name == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "bucket name is required"})
		return
	}

	region := req.Region
	if region == "" {
		region = s.defaultRegion
	}

	reqCtx := ctx.Context()

	exists, err := s.client.BucketExists(reqCtx, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("check bucket failed: %v", err),
		})
		return
	}

	if exists {
		ctx.JSON(http.StatusConflict, map[string]string{"error": "bucket already exists"})
		return
	}

	if err := s.client.MakeBucket(reqCtx, req.Name, minio.MakeBucketOptions{Region: region}); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("create bucket failed: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusCreated, map[string]any{
		"name":   req.Name,
		"region": region,
	})
}

func (s *StorageService) DeleteBucket(ctx guiio_http.Context) {
	bucketName := strings.TrimSpace(ctx.Param("bucketName"))
	if bucketName == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "bucket name is required"})
		return
	}

	reqCtx := ctx.Context()

	exists, err := s.client.BucketExists(reqCtx, bucketName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("check bucket failed: %v", err),
		})
		return
	}

	if !exists {
		ctx.JSON(http.StatusNotFound, map[string]string{"error": "bucket not found"})
		return
	}

	if err := s.client.RemoveBucket(reqCtx, bucketName); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("delete bucket failed: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{"deleted": bucketName})
}
