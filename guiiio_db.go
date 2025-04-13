package main

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/guiio/ent"
	"github.com/sphynx/config"

	_ "github.com/lib/pq"
)

var db *ent.Client

func InitDB() (db *ent.Client, err error) {
	tdb, err := sql.Open("postgres", config.Get[string]("db_dsn"))
	if err != nil {
		Mlog.Panic().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	// 🔧 커넥션 풀 설정
	postgres := tdb.DB()
	postgres.SetMaxIdleConns(config.Get[int]("db_max_idle_conns"))
	postgres.SetMaxOpenConns(config.Get[int]("db_max_conns"))
	postgres.SetConnMaxLifetime(time.Duration(config.Get[int]("db_max_timout")) * time.Minute)

	options := []ent.Option{
		ent.Debug(),
		ent.Log(func(i ...any) {
			Mlog.Info().Msgf("DB: %v", i)
		}),
		ent.Driver(tdb),
	}

	db = ent.NewClient(options...)

	return db, nil
}
