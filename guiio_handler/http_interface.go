package guiio_handler

import "net/http"

type BucketService interface {
	ListBucket(w http.ResponseWriter, r *http.Request)
	CreateBucket(w http.ResponseWriter, r *http.Request)
	DeleteBucket(w http.ResponseWriter, r *http.Request)
}
