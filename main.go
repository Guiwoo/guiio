package main

import (
	_ "embed"

	guiio_http "guiio/guiio_server"
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
	server := guiio_http.NewGuiioHttpServer(conf, Mlog)

	if err := server.Start(); err != nil {
		Mlog.Panic().Err(err).Msg("Failed to start server")
	}
}
