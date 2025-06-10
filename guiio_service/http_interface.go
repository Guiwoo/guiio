package guiio_service

import guiio_http "guiio/guiio_server/http"

type BucketService interface {
	ListBucket(ctx guiio_http.Context) error
	CreateBucket(ctx guiio_http.Context)
	DeleteBucket(ctx guiio_http.Context)
}
