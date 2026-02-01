package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
)

type fakeStorageClient struct {
	listResp     []minio.BucketInfo
	listErr      error
	existsMap    map[string]bool
	existsErr    error
	makeErr      error
	removeErr    error
	makeCalled   []string
	removeCalled []string
	putCalled    []string
	putErr       error
	objects      map[string][]byte
}

func (f *fakeStorageClient) ListBuckets(_ context.Context) ([]minio.BucketInfo, error) {
	return f.listResp, f.listErr
}

func (f *fakeStorageClient) BucketExists(_ context.Context, bucketName string) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.existsMap[bucketName], nil
}

func (f *fakeStorageClient) MakeBucket(_ context.Context, bucketName string, _ minio.MakeBucketOptions) error {
	f.makeCalled = append(f.makeCalled, bucketName)
	return f.makeErr
}

func (f *fakeStorageClient) RemoveBucket(_ context.Context, bucketName string) error {
	f.removeCalled = append(f.removeCalled, bucketName)
	return f.removeErr
}

func (f *fakeStorageClient) PutObject(_ context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, _ minio.PutObjectOptions) (minio.UploadInfo, error) {
	if f.putErr != nil {
		return minio.UploadInfo{}, f.putErr
	}
	data, _ := io.ReadAll(reader)
	if objectSize == 0 {
		objectSize = int64(len(data))
	}
	f.putCalled = append(f.putCalled, fmt.Sprintf("%s/%s", bucketName, objectName))
	if f.objects == nil {
		f.objects = map[string][]byte{}
	}
	f.objects[fmt.Sprintf("%s/%s", bucketName, objectName)] = data
	return minio.UploadInfo{Size: objectSize, ETag: "etag"}, nil
}

func (f *fakeStorageClient) GetObject(_ context.Context, bucketName, objectName string, _ minio.GetObjectOptions) (io.ReadCloser, error) {
	key := fmt.Sprintf("%s/%s", bucketName, objectName)
	data := f.objects[key]
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (f *fakeStorageClient) StatObject(_ context.Context, bucketName, objectName string, _ minio.StatObjectOptions) (minio.ObjectInfo, error) {
	key := fmt.Sprintf("%s/%s", bucketName, objectName)
	data, ok := f.objects[key]
	if !ok {
		return minio.ObjectInfo{}, errors.New("not found")
	}
	return minio.ObjectInfo{Size: int64(len(data)), ContentType: "application/octet-stream", ETag: "etag"}, nil
}

type fakeContext struct {
	body    []byte
	params  map[string]string
	status  int
	resp    interface{}
	bindErr error
	stream  []byte
}

func (c *fakeContext) JSON(code int, v interface{}) error {
	c.status = code
	c.resp = v
	return nil
}

func (c *fakeContext) Bind(v interface{}) error {
	if c.bindErr != nil {
		return c.bindErr
	}
	return json.Unmarshal(c.body, v)
}

func (c *fakeContext) Param(name string) string {
	return c.params[name]
}

func (c *fakeContext) Query(string) string      { return "" }
func (c *fakeContext) GetHeader(string) string  { return "" }
func (c *fakeContext) SetHeader(string, string) {}
func (c *fakeContext) Context() context.Context { return context.Background() }
func (c *fakeContext) Request() *http.Request   { return nil }
func (c *fakeContext) Stream(code int, _ string, r io.Reader) error {
	c.status = code
	data, _ := io.ReadAll(r)
	c.stream = data
	return nil
}

func TestCreateBucket(t *testing.T) {
	client := &fakeStorageClient{existsMap: map[string]bool{"dup": true}}
	svc := NewStorageServiceWithClient(client, "default", nil)

	t.Run("success", func(t *testing.T) {
		ctx := &fakeContext{body: []byte(`{"name":"ok1"}`)}
		svc.CreateBucket(ctx)

		if ctx.status != http.StatusCreated {
			t.Fatalf("expected %d got %d", http.StatusCreated, ctx.status)
		}
		if len(client.makeCalled) != 1 || client.makeCalled[0] != "ok1" {
			t.Fatalf("make bucket not called correctly: %+v", client.makeCalled)
		}
		resp := ctx.resp.(BucketResponse)
		if resp.Name != "ok1" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("already exists", func(t *testing.T) {
		ctx := &fakeContext{body: []byte(`{"name":"dup"}`)}
		svc.CreateBucket(ctx)
		if ctx.status != http.StatusConflict {
			t.Fatalf("expected conflict got %d", ctx.status)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		ctx := &fakeContext{body: []byte(`{`)}
		svc.CreateBucket(ctx)
		if ctx.status != http.StatusBadRequest {
			t.Fatalf("expected bad request got %d", ctx.status)
		}
	})

	t.Run("invalid name", func(t *testing.T) {
		ctx := &fakeContext{body: []byte(`{"name":"UPPER"}`)}
		svc.CreateBucket(ctx)
		if ctx.status != http.StatusBadRequest {
			t.Fatalf("expected bad request got %d", ctx.status)
		}
	})

	t.Run("bucket check error", func(t *testing.T) {
		client.existsErr = errors.New("boom")
		ctx := &fakeContext{body: []byte(`{"name":"err"}`)}
		svc.CreateBucket(ctx)
		if ctx.status != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", ctx.status)
		}
		client.existsErr = nil
	})
}

func TestDeleteBucket(t *testing.T) {
	client := &fakeStorageClient{existsMap: map[string]bool{"keep": true}}
	svc := NewStorageServiceWithClient(client, "", nil)

	t.Run("success", func(t *testing.T) {
		ctx := &fakeContext{params: map[string]string{"bucketName": "keep"}}
		svc.DeleteBucket(ctx)
		if ctx.status != http.StatusOK {
			t.Fatalf("expected ok got %d", ctx.status)
		}
		if len(client.removeCalled) != 1 || client.removeCalled[0] != "keep" {
			t.Fatalf("remove not called: %+v", client.removeCalled)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctx := &fakeContext{params: map[string]string{"bucketName": "none"}}
		svc.DeleteBucket(ctx)
		if ctx.status != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", ctx.status)
		}
	})

	t.Run("invalid name", func(t *testing.T) {
		ctx := &fakeContext{params: map[string]string{"bucketName": "Nope"}}
		svc.DeleteBucket(ctx)
		if ctx.status != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", ctx.status)
		}
	})

	t.Run("exists error", func(t *testing.T) {
		client.existsErr = errors.New("boom")
		ctx := &fakeContext{params: map[string]string{"bucketName": "keep"}}
		svc.DeleteBucket(ctx)
		if ctx.status != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", ctx.status)
		}
		client.existsErr = nil
	})
}

func TestListBucket(t *testing.T) {
	now := time.Now()
	client := &fakeStorageClient{
		listResp: []minio.BucketInfo{{Name: "a", CreationDate: now}},
	}
	svc := NewStorageServiceWithClient(client, "", nil)
	ctx := &fakeContext{}

	err := svc.ListBucket(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.status != http.StatusOK {
		t.Fatalf("expected 200 got %d", ctx.status)
	}
	resp := ctx.resp.(BucketListResponse).Buckets
	if len(resp) != 1 || resp[0].Name != "a" || !resp[0].CreatedAt.Equal(now) {
		t.Fatalf("unexpected buckets: %+v", resp)
	}
}

func TestGetBucket(t *testing.T) {
	now := time.Now()
	client := &fakeStorageClient{
		listResp: []minio.BucketInfo{{Name: "alive", CreationDate: now}},
	}
	svc := NewStorageServiceWithClient(client, "", nil)

	t.Run("found", func(t *testing.T) {
		ctx := &fakeContext{params: map[string]string{"bucketName": "alive"}}
		svc.GetBucket(ctx)
		if ctx.status != http.StatusOK {
			t.Fatalf("expected 200 got %d", ctx.status)
		}
		resp := ctx.resp.(BucketResponse)
		if resp.Name != "alive" || !resp.CreatedAt.Equal(now) {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctx := &fakeContext{params: map[string]string{"bucketName": "missing"}}
		svc.GetBucket(ctx)
		if ctx.status != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", ctx.status)
		}
	})

	t.Run("invalid name", func(t *testing.T) {
		ctx := &fakeContext{params: map[string]string{"bucketName": "AA"}}
		svc.GetBucket(ctx)
		if ctx.status != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", ctx.status)
		}
	})

	t.Run("list error", func(t *testing.T) {
		client.listErr = errors.New("boom")
		ctx := &fakeContext{params: map[string]string{"bucketName": "alive"}}
		svc.GetBucket(ctx)
		if ctx.status != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", ctx.status)
		}
		client.listErr = nil
	})
}
