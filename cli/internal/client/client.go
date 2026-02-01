package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseURL string
	http    HTTPClient
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

type CreateBucketRequest struct {
	Name   string `json:"name"`
	Region string `json:"region,omitempty"`
}

type UploadObjectResponse struct {
	Bucket      string `json:"bucket"`
	Object      string `json:"object"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	ETag        string `json:"etag"`
	StoragePath string `json:"storage_path"`
}

type DownloadResult struct {
	Data        []byte
	ContentType string
	FileName    string
}

func New(baseURL string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api/v1"
	}
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) ListBuckets(ctx context.Context) (BucketListResponse, error) {
	var out BucketListResponse
	if err := c.get(ctx, "/buckets", &out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) GetBucket(ctx context.Context, name string) (BucketResponse, error) {
	var out BucketResponse
	if err := c.get(ctx, fmt.Sprintf("/buckets/%s", url.PathEscape(name)), &out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) CreateBucket(ctx context.Context, name, region string) (BucketResponse, error) {
	payload := CreateBucketRequest{Name: name, Region: region}
	var out BucketResponse
	if err := c.post(ctx, "/buckets", payload, &out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) DeleteBucket(ctx context.Context, name string) (DeleteBucketResponse, error) {
	var out DeleteBucketResponse
	if err := c.delete(ctx, fmt.Sprintf("/buckets/%s", url.PathEscape(name)), &out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) UploadObject(ctx context.Context, bucket, filePath, objectName string, meta map[string]string) (UploadObjectResponse, error) {
	var out UploadObjectResponse

	f, err := os.Open(filePath)
	if err != nil {
		return out, err
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fh, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return out, err
	}
	if _, err := io.Copy(fh, f); err != nil {
		return out, err
	}

	if objectName != "" {
		_ = writer.WriteField("objectName", objectName)
	}

	for k, v := range meta {
		_ = writer.WriteField("meta-"+k, v)
	}

	if err := writer.Close(); err != nil {
		return out, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/buckets/%s/objects", c.baseURL, url.PathEscape(bucket)), body)
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := c.do(req, &out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) DownloadObject(ctx context.Context, bucket, object string) (DownloadResult, error) {
	var out DownloadResult
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/buckets/%s/objects/%s", c.baseURL, url.PathEscape(bucket), url.PathEscape(object)), nil)
	if err != nil {
		return out, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var er ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&er)
		if er.Error == "" {
			er.Error = resp.Status
		}
		return out, fmt.Errorf("%s", er.Error)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}
	out.Data = data
	out.ContentType = resp.Header.Get("Content-Type")
	out.FileName = object
	return out, nil
}

func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, out interface{}) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) delete(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var er ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&er)
		if er.Error == "" {
			er.Error = resp.Status
		}
		return fmt.Errorf("%s", er.Error)
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
