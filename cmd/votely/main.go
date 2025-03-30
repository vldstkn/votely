package main

import (
	"github.com/tarantool/go-tarantool/v2"
	"github.com/vldstkn/votely/internal/config"
	"github.com/vldstkn/votely/internal/votely"
	"github.com/vldstkn/votely/pkg/db"
	"github.com/vldstkn/votely/pkg/logger"
	"log/slog"
	"os"
)

func main() {
	mode := os.Getenv("APP_ENV")
	if mode == "" {
		mode = "dev"
	}
	cnf := config.LoadConfig("configs", mode)
	log := logger.NewLogger(os.Stdout)
	database, err := db.InitTrntlDb(tarantool.NetDialer{
		Address:  cnf.Database.Tarantool.Address,
		User:     cnf.Database.Tarantool.User,
		Password: cnf.Database.Tarantool.Password,
	}, tarantool.Opts{})
	if err != nil {

		log.Error(err.Error(),
			slog.String("Address", cnf.Database.Tarantool.Address),
			slog.String("User", cnf.Database.Tarantool.User),
		)
		return
	}
	app, err := votely.NewApp(&votely.AppDeps{
		Logger: log,
		Config: cnf,
		Mode:   mode,
		Db:     database,
	})
	if err != nil {
		return
	}
	_ = app.Run()
	log.Info("Bot has been stopped...",
		slog.String("Mode", mode),
		slog.String("Bot name", cnf.Bot.Name),
	)
}
