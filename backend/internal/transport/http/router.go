package httptransport

import (
	"fmt"
	"net/http"

	"guiio/backend/internal/middleware"
	httpctx "guiio/backend/internal/port/httpctx"
	"guiio/backend/internal/repository"
	"guiio/backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/sphynx/config"
	httpSwagger "github.com/swaggo/http-swagger"

	docs "guiio/backend/docs"
)

type HttpHandler struct {
	conf          *config.GConfig
	log           *zerolog.Logger
	bucketService service.BucketService
}

func NewHttpHandler(conf *config.GConfig, log *zerolog.Logger, repo repository.ObjectRepository) (*HttpHandler, error) {
	//Todo 밖으로 빼기
	bucketService, err := service.NewStorageService(repo)
	if err != nil {
		return nil, err
	}

	return &HttpHandler{
		conf:          conf,
		log:           log,
		bucketService: bucketService,
	}, nil
}

func (h *HttpHandler) Start() error {
	port := config.Get[int]("port")
	serverName := config.Get[string]("server_name")
	allowOrigin := config.Get[string]("cors_allow_origin")

	h.log.Info().Msgf("Server Starts %d", port)

	router := chi.NewRouter()

	router.Use(middleware.HttrRequestLogger(h.log, serverName))
	router.Use(middleware.CORSMiddleware(allowOrigin))

	docs.SwaggerInfo.BasePath = "/api/v1"

	router.Get("/swagger/*", httpSwagger.Handler())

	router.Route("/api/v1/buckets", func(r chi.Router) {
		r.Get("/", h.ListBucket)
		r.Post("/", h.CreateBucket)
		r.Get("/{bucketName}", h.GetBucket)
		r.Delete("/{bucketName}", h.DeleteBucket)
		r.Post("/{bucketName}/objects", h.UploadObject)
		r.Get("/{bucketName}/objects/{objectName}", h.DownloadObject)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

// ListBucket godoc
// @Summary 버킷 목록 조회
// @Description 저장소에 존재하는 모든 버킷을 반환합니다.
// @Tags buckets
// @Produce json
// @Success 200 {object} service.BucketListResponse
// @Failure 500 {object} service.ErrorResponse
// @Router /api/v1/buckets [get]
func (h *HttpHandler) ListBucket(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.NewChiContext(w, r)
	h.bucketService.ListBucket(ctx)
}

// CreateBucket godoc
// @Summary 버킷 생성
// @Description 버킷 이름과 선택적 리전을 받아 새 버킷을 생성합니다.
// @Tags buckets
// @Accept json
// @Produce json
// @Param bucket body service.CreateBucketRequest true "버킷 생성 요청"
// @Success 201 {object} service.BucketResponse
// @Failure 400 {object} service.ErrorResponse
// @Failure 409 {object} service.ErrorResponse
// @Failure 500 {object} service.ErrorResponse
// @Router /api/v1/buckets [post]
func (h *HttpHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.NewChiContext(w, r)

	h.bucketService.CreateBucket(ctx)
}

// GetBucket godoc
// @Summary 버킷 상세 조회
// @Description 버킷 존재 여부와 생성 시점을 반환합니다.
// @Tags buckets
// @Produce json
// @Param bucketName path string true "버킷 이름"
// @Success 200 {object} service.BucketResponse
// @Failure 400 {object} service.ErrorResponse
// @Failure 404 {object} service.ErrorResponse
// @Failure 500 {object} service.ErrorResponse
// @Router /api/v1/buckets/{bucketName} [get]
func (h *HttpHandler) GetBucket(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.NewChiContext(w, r)
	h.bucketService.GetBucket(ctx)
}

// DeleteBucket godoc
// @Summary 버킷 삭제
// @Description 버킷을 삭제합니다. 비어 있지 않은 버킷은 스토리지에서 거부될 수 있습니다.
// @Tags buckets
// @Produce json
// @Param bucketName path string true "버킷 이름"
// @Success 200 {object} service.DeleteBucketResponse
// @Failure 400 {object} service.ErrorResponse
// @Failure 404 {object} service.ErrorResponse
// @Failure 500 {object} service.ErrorResponse
// @Router /api/v1/buckets/{bucketName} [delete]
func (h *HttpHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.NewChiContext(w, r)
	h.bucketService.DeleteBucket(ctx)
}

// UploadObject godoc
// @Summary 객체 업로드
// @Description 멀티파트 파일을 업로드하고 메타데이터와 함께 저장합니다.
// @Tags buckets
// @Accept multipart/form-data
// @Produce json
// @Param bucketName path string true "버킷 이름"
// @Param file formData file true "업로드 파일"
// @Param objectName formData string false "저장할 객체 이름"
// @Param meta-xxx formData string false "메타데이터 (meta- 접두사 사용)"
// @Success 201 {object} service.UploadObjectResponse
// @Failure 400 {object} service.ErrorResponse
// @Failure 404 {object} service.ErrorResponse
// @Failure 500 {object} service.ErrorResponse
// @Router /api/v1/buckets/{bucketName}/objects [post]
func (h *HttpHandler) UploadObject(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.NewChiContext(w, r)
	h.bucketService.UploadObject(ctx)
}

// DownloadObject godoc
// @Summary 객체 다운로드
// @Description 버킷의 객체를 스트리밍으로 반환합니다.
// @Tags buckets
// @Produce octet-stream
// @Param bucketName path string true "버킷 이름"
// @Param objectName path string true "객체 이름"
// @Success 200 {file} binary
// @Success 304 {string} string "Not Modified"
// @Failure 400 {object} service.ErrorResponse
// @Failure 404 {object} service.ErrorResponse
// @Failure 500 {object} service.ErrorResponse
// @Router /api/v1/buckets/{bucketName}/objects/{objectName} [get]
func (h *HttpHandler) DownloadObject(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.NewChiContext(w, r)
	h.bucketService.DownloadObject(ctx)
}
