package guiio_service

import "net/http"

type StorageService struct {
}

func NewStorageService() *StorageService {
	return &StorageService{}
}

func (s *StorageService) ListBucket(w http.ResponseWriter, r *http.Request) {
}

func (s *StorageService) CreateBucket(w http.ResponseWriter, r *http.Request) {
}

func (s *StorageService) DeleteBucket(w http.ResponseWriter, r *http.Request) {
}
