package main

import (
	_ "embed"

	configenv "guiio/backend/internal/config"
	database "guiio/backend/internal/infra/db"
	"guiio/backend/internal/repository"
	httptransport "guiio/backend/internal/transport/http"
	"guiio/backend/internal/util"

	"github.com/rs/zerolog"
	"github.com/sphynx/config"
	"github.com/sphynx/logger"
)

var (
	Version string
	//go:embed banner.txt
	banner []byte
	Mlog   *zerolog.Logger
)

// @title guiio API
// @version 1.0
// @description MinIO 스타일 객체 저장소 API
// @BasePath /api/v1
func main() {
	util.ServerInfo(banner, Version)
	Mlog = logger.New()

	env, err := configenv.GetEnv()
	if err != nil {
		Mlog.Panic().Err(err).Msg("Failed to get env")
		return
	}

	conf := config.NewConfig(env)
	db, err := database.InitDB(Mlog)
	if err != nil {
		Mlog.Panic().Err(err).Msg("Failed to init db")
		return
	}
	defer db.Close()

	repo := repository.NewObjectRepository(db)
	handler, err := httptransport.NewHttpHandler(conf, Mlog, repo)
	if err != nil {
		Mlog.Panic().Err(err).Msg("Failed to create http handler")
		return
	}

	if err := handler.Start(); err != nil {
		Mlog.Panic().Err(err).Msg("Failed to start server")
	}
}
