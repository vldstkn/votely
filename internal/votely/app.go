package votely

import (
	"github.com/go-chi/chi/v5"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/vldstkn/votely/internal/config"
	"github.com/vldstkn/votely/internal/mattermost"
	"github.com/vldstkn/votely/pkg/db"
	"log/slog"
	"net/http"
)

type AppDeps struct {
	Logger *slog.Logger
	Config *config.Config
	Db     *db.TrntlDb
	Mode   string
}

type App struct {
	Logger *slog.Logger
	Config *config.Config
	Client *model.Client4
	Db     *db.TrntlDb
	Mode   string
}

func NewApp(deps *AppDeps) (*App, error) {
	client := model.NewAPIv4Client(deps.Config.Mattermost.Url)
	client.SetToken(deps.Config.Bot.Token)
	_, _, err := client.GetUser("me", "")
	if err != nil {
		deps.Logger.Error(err.Error(),
			slog.String("Error location", "NewApp.client.GetUser"),
			slog.String("Mattermost URL", deps.Config.Mattermost.Url),
			slog.String("Mode", deps.Mode),
		)
		return nil, err
	}
	err = mattermost.RegisterVoteCommand(client, deps.Config.Mattermost.TeamName, deps.Config.GetHttpUrlBot())
	if err != nil {
		deps.Logger.Error(err.Error(),
			slog.String("Error location", "mattermost.RegisterVoteCommand"),
			slog.String("Mattermost URL", deps.Config.Mattermost.Url),
			slog.String("Mode", deps.Mode),
		)
	}

	return &App{
		Logger: deps.Logger,
		Config: deps.Config,
		Client: client,
		Mode:   deps.Mode,
		Db:     deps.Db,
	}, nil
}

func (app *App) Run() error {
	repository := NewRepository(&RepositoryDeps{
		Logger: app.Logger,
		Db:     app.Db,
	})
	service := NewService(&ServiceDeps{
		Logger:     app.Logger,
		Repository: repository,
	})
	router := chi.NewRouter()

	NewHandler(router, &HandlerDeps{
		Logger:  app.Logger,
		Service: service,
		Client:  app.Client,
		Config:  app.Config,
	})
	server := http.Server{
		Addr:    app.Config.GetUrlBot(),
		Handler: router,
	}
	defer server.Close()

	app.Logger.Info("Bot is starting...",
		slog.String("Bot name", app.Config.Bot.Name),
		slog.String("Mode", app.Mode),
		slog.String("Bot address", app.Config.GetUrlBot()),
	)

	err := server.ListenAndServe()
	if err != nil {
		app.Logger.Error(err.Error(),
			slog.String("Error location", "server.ListenAndServe"),
			slog.String("Bot url", app.Config.GetUrlBot()),
		)
	}
	return err
}
