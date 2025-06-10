package guiio_service

import (
	guiio_http "guiio/guiio_server/http"
)

type StorageService struct {
}

func NewStorageService() *StorageService {
	return &StorageService{}
}

func (s *StorageService) ListBucket(ctx guiio_http.Context) error {
	return nil
}

func (s *StorageService) CreateBucket(ctx guiio_http.Context) {
}

func (s *StorageService) DeleteBucket(ctx guiio_http.Context) {
}
