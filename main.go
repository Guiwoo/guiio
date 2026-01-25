package main

import (
	_ "embed"

	"guiio/guiio_handler"
	guiio_repository "guiio/guiio_repository"
	"guiio/guiio_util"

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

func main() {
	guiio_util.ServerInfo(banner, Version)
	Mlog = logger.New()

	env, err := guiio_util.GetEnv()
	if err != nil {
		Mlog.Panic().Err(err).Msg("Failed to get env")
		return
	}

	conf := config.NewConfig(env)
	db, err := InitDB()
	if err != nil {
		Mlog.Panic().Err(err).Msg("Failed to init db")
		return
	}
	defer db.Close()

	repo := guiio_repository.NewObjectRepository(db)
	handler, err := guiio_handler.NewHttpHandler(conf, Mlog, repo)
	if err != nil {
		Mlog.Panic().Err(err).Msg("Failed to create http handler")
		return
	}

	if err := handler.Start(); err != nil {
		Mlog.Panic().Err(err).Msg("Failed to start server")
	}
}
