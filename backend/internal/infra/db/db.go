package database

import (
	"time"

	"guiio/backend/ent"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog"
	"github.com/sphynx/config"

	_ "github.com/lib/pq"
)

func InitDB(log *zerolog.Logger) (*ent.Client, error) {
	tdb, err := sql.Open("postgres", config.Get[string]("db_dsn"))
	if err != nil {
		log.Panic().Err(err).Msg("Failed to connect to database")
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
			log.Info().Msgf("DB: %v", i)
		}),
		ent.Driver(tdb),
	}

	db := ent.NewClient(options...)

	return db, nil
}
