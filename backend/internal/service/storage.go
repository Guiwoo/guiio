package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"guiio/backend/ent"
	httpctx "guiio/backend/internal/port/httpctx"
	"guiio/backend/internal/repository"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sphynx/config"
)

type StorageService struct {
	client        StorageClient
	defaultRegion string
	repo          repository.ObjectRepository
}

type minioWrapper struct {
	c *minio.Client
}

func (m *minioWrapper) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return m.c.ListBuckets(ctx)
}

func (m *minioWrapper) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return m.c.BucketExists(ctx, bucketName)
}

func (m *minioWrapper) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	return m.c.MakeBucket(ctx, bucketName, opts)
}

func (m *minioWrapper) RemoveBucket(ctx context.Context, bucketName string) error {
	return m.c.RemoveBucket(ctx, bucketName)
}

func (m *minioWrapper) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return m.c.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (m *minioWrapper) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	return m.c.GetObject(ctx, bucketName, objectName, opts)
}

func (m *minioWrapper) StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return m.c.StatObject(ctx, bucketName, objectName, opts)
}

type StorageClient interface {
	ListBuckets(ctx context.Context) ([]minio.BucketInfo, error)
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	RemoveBucket(ctx context.Context, bucketName string) error
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
	StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
}

type BucketInfo struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type BucketListResponse struct {
	Buckets []BucketInfo `json:"buckets"`
}

type BucketResponse struct {
	Name      string    `json:"name"`
	Region    string    `json:"region,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type DeleteBucketResponse struct {
	Deleted string `json:"deleted"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type UploadObjectResponse struct {
	Bucket      string `json:"bucket"`
	Object      string `json:"object"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	ETag        string `json:"etag"`
	StoragePath string `json:"storage_path"`
}

type CreateBucketRequest struct {
	Name   string `json:"name"`
	Region string `json:"region,omitempty"`
}

var bucketNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

func NewStorageService(repo repository.ObjectRepository) (*StorageService, error) {
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

	return NewStorageServiceWithClient(&minioWrapper{c: client}, region, repo), nil
}

func NewStorageServiceWithClient(client StorageClient, defaultRegion string, repo repository.ObjectRepository) *StorageService {
	return &StorageService{
		client:        client,
		defaultRegion: defaultRegion,
		repo:          repo,
	}
}

func (s *StorageService) ListBucket(ctx httpctx.Context) error {
	buckets, err := s.client.ListBuckets(ctx.Context())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("list buckets failed: %v", err)})
	}

	result := make([]BucketInfo, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, BucketInfo{
			Name:      b.Name,
			CreatedAt: b.CreationDate,
		})
	}

	return ctx.JSON(http.StatusOK, BucketListResponse{Buckets: result})
}

func (s *StorageService) CreateBucket(ctx httpctx.Context) {
	var req CreateBucketRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Region = strings.TrimSpace(req.Region)

	if err := validateBucketName(req.Name); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	region := req.Region
	if region == "" {
		region = s.defaultRegion
	}

	reqCtx := ctx.Context()

	exists, err := s.client.BucketExists(reqCtx, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("check bucket failed: %v", err)})
		return
	}

	if exists {
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: "bucket already exists"})
		return
	}

	if err := s.client.MakeBucket(reqCtx, req.Name, minio.MakeBucketOptions{Region: region}); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("create bucket failed: %v", err)})
		return
	}

	ctx.JSON(http.StatusCreated, BucketResponse{
		Name:   req.Name,
		Region: region,
	})
}

func (s *StorageService) DeleteBucket(ctx httpctx.Context) {
	bucketName := strings.TrimSpace(ctx.Param("bucketName"))
	if err := validateBucketName(bucketName); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	reqCtx := ctx.Context()

	exists, err := s.client.BucketExists(reqCtx, bucketName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("check bucket failed: %v", err)})
		return
	}

	if !exists {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "bucket not found"})
		return
	}

	if err := s.client.RemoveBucket(reqCtx, bucketName); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("delete bucket failed: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, DeleteBucketResponse{Deleted: bucketName})
}

func (s *StorageService) GetBucket(ctx httpctx.Context) {
	bucketName := strings.TrimSpace(ctx.Param("bucketName"))
	if err := validateBucketName(bucketName); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	reqCtx := ctx.Context()
	buckets, err := s.client.ListBuckets(reqCtx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("list buckets failed: %v", err)})
		return
	}

	for _, b := range buckets {
		if b.Name == bucketName {
			ctx.JSON(http.StatusOK, BucketResponse{
				Name:      b.Name,
				CreatedAt: b.CreationDate,
			})
			return
		}
	}

	ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "bucket not found"})
}

func validateBucketName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("bucket name is required")
	}

	if len(name) < 3 || len(name) > 63 {
		return errors.New("bucket name must be between 3 and 63 characters")
	}

	if !bucketNameRegex.MatchString(name) {
		return errors.New("bucket name must use lowercase letters, numbers, hyphen or dot, and start/end with alphanumeric")
	}

	if strings.Contains(name, "..") {
		return errors.New("bucket name cannot contain consecutive dots")
	}

	return nil
}

func validateObjectName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("object name is required")
	}
	return nil
}

func encodeObjectKey(name string) string {
	segments := strings.Split(name, "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	return strings.Join(segments, "/")
}

func normalizeStoragePath(bucketName, storedPath, fallback string) string {
	if storedPath == "" {
		return fallback
	}
	if strings.HasPrefix(storedPath, bucketName+"/") {
		return strings.TrimPrefix(storedPath, bucketName+"/")
	}
	return storedPath
}

func (s *StorageService) DownloadObject(ctx httpctx.Context) {
	bucketName := strings.TrimSpace(ctx.Param("bucketName"))
	objectName := strings.TrimSpace(ctx.Param("objectName"))
	if err := validateBucketName(bucketName); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := validateObjectName(objectName); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	encodedName := encodeObjectKey(objectName)
	cacheControl := config.Get[string]("object_cache_control")
	req := ctx.Request()

	if s.repo != nil {
		if obj, err := s.repo.GetObject(ctx.Context(), bucketName, objectName); err == nil {
			etag := obj.Etag
			storageKey := normalizeStoragePath(bucketName, obj.StoragePath, encodedName)
			if cacheControl != "" {
				ctx.SetHeader("Cache-Control", cacheControl)
			}
			ctx.SetHeader("ETag", etag)
			ctx.SetHeader("Last-Modified", obj.UpdatedAt.UTC().Format(http.TimeFormat))

			if req != nil {
				clientETag := strings.Trim(req.Header.Get("If-None-Match"), "\"")
				serverETag := strings.Trim(etag, "\"")
				if clientETag != "" && clientETag == serverETag {
					_ = ctx.Stream(http.StatusNotModified, "", bytes.NewReader(nil))
					return
				}
				if ims := req.Header.Get("If-Modified-Since"); ims != "" {
					if t, err := http.ParseTime(ims); err == nil && !obj.UpdatedAt.After(t) {
						_ = ctx.Stream(http.StatusNotModified, "", bytes.NewReader(nil))
						return
					}
				}
			}

			objReader, err := s.client.GetObject(ctx.Context(), bucketName, storageKey, minio.GetObjectOptions{})
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("download failed: %v", err)})
				return
			}
			defer objReader.Close()

			if obj.ContentType != "" {
				ctx.SetHeader("Content-Type", obj.ContentType)
			}
			ctx.SetHeader("Content-Length", fmt.Sprintf("%d", obj.Size))

			if err := ctx.Stream(http.StatusOK, obj.ContentType, objReader); err != nil {
				ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("stream failed: %v", err)})
				return
			}
			return
		} else if !ent.IsNotFound(err) {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("get object metadata failed: %v", err)})
			return
		}
	}

	info, err := s.client.StatObject(ctx.Context(), bucketName, encodedName, minio.StatObjectOptions{})
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "object not found"})
		return
	}

	ctx.SetHeader("ETag", info.ETag)
	if cacheControl != "" {
		ctx.SetHeader("Cache-Control", cacheControl)
	}
	ctx.SetHeader("Last-Modified", info.LastModified.UTC().Format(http.TimeFormat))

	if req != nil {
		clientETag := strings.Trim(req.Header.Get("If-None-Match"), "\"")
		serverETag := strings.Trim(info.ETag, "\"")
		if clientETag != "" && clientETag == serverETag {
			_ = ctx.Stream(http.StatusNotModified, "", bytes.NewReader(nil))
			return
		}
		if ims := req.Header.Get("If-Modified-Since"); ims != "" {
			if t, err := http.ParseTime(ims); err == nil && !info.LastModified.After(t) {
				_ = ctx.Stream(http.StatusNotModified, "", bytes.NewReader(nil))
				return
			}
		}
	}

	obj, err := s.client.GetObject(ctx.Context(), bucketName, encodedName, minio.GetObjectOptions{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("download failed: %v", err)})
		return
	}
	defer obj.Close()

	if info.ContentType != "" {
		ctx.SetHeader("Content-Type", info.ContentType)
	}
	ctx.SetHeader("Content-Length", fmt.Sprintf("%d", info.Size))

	if err := ctx.Stream(http.StatusOK, info.ContentType, obj); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("stream failed: %v", err)})
		return
	}
}

func (s *StorageService) UploadObject(ctx httpctx.Context) {
	bucketName := strings.TrimSpace(ctx.Param("bucketName"))
	if err := validateBucketName(bucketName); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	r := ctx.Request()
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	objectName := strings.TrimSpace(r.FormValue("objectName"))
	if objectName == "" {
		objectName = header.Filename
	}
	if err := validateObjectName(objectName); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	encodedName := encodeObjectKey(objectName)

	metadata := map[string]string{}
	for k, vals := range r.PostForm {
		if strings.HasPrefix(k, "meta-") && len(vals) > 0 {
			metadata[strings.TrimPrefix(k, "meta-")] = vals[0]
		}
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// We need size; use header.Size if available, else read into buffer
	var reader io.Reader = file
	size := header.Size

	if size <= 0 {
		data, err := io.ReadAll(file)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("read file: %v", err)})
			return
		}
		size = int64(len(data))
		reader = bytes.NewReader(data)
	}

	uinfo, err := s.client.PutObject(ctx.Context(), bucketName, encodedName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("upload failed: %v", err)})
		return
	}

	storagePath := encodedName
	if s.repo != nil {
		_, _ = s.repo.UpsertObject(ctx.Context(), repository.ObjectUpsertInput{
			BucketName:  bucketName,
			ObjectName:  objectName,
			StoragePath: storagePath,
			ContentType: contentType,
			Size:        uinfo.Size,
			ETag:        uinfo.ETag,
			Metadata:    metadata,
		})
	}

	ctx.JSON(http.StatusCreated, UploadObjectResponse{
		Bucket:      bucketName,
		Object:      objectName,
		ContentType: contentType,
		Size:        uinfo.Size,
		ETag:        uinfo.ETag,
		StoragePath: storagePath,
	})
}
