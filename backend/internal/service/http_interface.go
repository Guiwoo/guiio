package service

import httpctx "guiio/backend/internal/port/httpctx"

type BucketService interface {
	ListBucket(ctx httpctx.Context) error
	CreateBucket(ctx httpctx.Context)
	DeleteBucket(ctx httpctx.Context)
	GetBucket(ctx httpctx.Context)
	UploadObject(ctx httpctx.Context)
	DownloadObject(ctx httpctx.Context)
}
