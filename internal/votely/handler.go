package votely

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/vldstkn/votely/internal/config"
	"github.com/vldstkn/votely/internal/constants"
	"github.com/vldstkn/votely/internal/interfaces"
	"github.com/vldstkn/votely/internal/models"
	"github.com/vldstkn/votely/pkg/res"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

type HandlerDeps struct {
	Logger  *slog.Logger
	Service interfaces.Service
	Client  *model.Client4
	Config  *config.Config
}

type Handler struct {
	Logger  *slog.Logger
	Service interfaces.Service
	Client  *model.Client4
	Config  *config.Config
}

func NewHandler(router *chi.Mux, deps *HandlerDeps) {
	handler := &Handler{
		Logger:  deps.Logger,
		Service: deps.Service,
		Client:  deps.Client,
		Config:  deps.Config,
	}
	router.Group(func(r chi.Router) {
		r.Post(string(constants.Create), handler.Create())
		r.Post(string(constants.SubmitCreate), handler.CreateSubmit())
		r.Get(string(constants.GetById), handler.GetById())
		r.Post(string(constants.Vote), handler.Vote())
		r.Post(string(constants.EndVoting), handler.EndVoting())
		r.Post(string(constants.DeleteVoting), handler.DeleteVoting())
	})
}

func (handler *Handler) DeleteVoting() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.FormValue("user_id")
		if userId == "" {
			res.Json(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		channelId := r.FormValue("channel_id")
		if userId == "" {
			res.Json(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		text := r.FormValue("text")
		if text == "" {
			err := handler.postToChannel("Вы не указали параметр.\n\n```/end <poll_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		ids := strings.Split(text, " ")
		if len(ids) == 0 {
			err := handler.postToChannel("Вы не указали параметр <option_id>.\n\n```/end <poll_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		pollId, err := strconv.Atoi(strings.TrimSpace(ids[0]))
		if err != nil || pollId < 0 {
			err := handler.postToChannel("Параметр <poll_id> должен быть целочисленным и положительным.\n\n```/end <poll_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		out := handler.Service.DeleteVoting(userId, uint(pollId))
		err = handler.postToChannel(out, userId, channelId)
		if err != nil {
			handler.Logger.Error(err.Error())
			res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (handler *Handler) EndVoting() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.FormValue("user_id")
		if userId == "" {
			res.Json(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		channelId := r.FormValue("channel_id")
		if userId == "" {
			res.Json(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		text := r.FormValue("text")
		if text == "" {
			err := handler.postToChannel("Вы не указали параметр.\n\n```/end <poll_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		ids := strings.Split(text, " ")
		if len(ids) == 0 {
			err := handler.postToChannel("Вы не указали параметр <option_id>.\n\n```/end <poll_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		pollId, err := strconv.Atoi(strings.TrimSpace(ids[0]))
		if err != nil {
			err := handler.postToChannel("Параметр <poll_id> должен быть целочисленным.\n\n```/end <poll_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		out := handler.Service.EndVoting(userId, uint(pollId))
		err = handler.postToChannel(out, userId, channelId)
		if err != nil {
			handler.Logger.Error(err.Error())
			res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (handler *Handler) Vote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.FormValue("user_id")
		if userId == "" {
			res.Json(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		channelId := r.FormValue("channel_id")
		if userId == "" {
			res.Json(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		text := r.FormValue("text")
		if text == "" {
			err := handler.postToChannel("Вы не указали параметры.\n\n```/vote <poll_id> <option_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		ids := strings.Split(text, " ")
		if len(ids) == 1 {
			err := handler.postToChannel("Вы не указали параметр <option_id>.\n\n```/vote <poll_id> <option_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		pollIdStr, optionIdStr := ids[0], ids[1]
		pollId, err := strconv.Atoi(pollIdStr)
		if err != nil {
			err := handler.postToChannel("Параметр <poll_id> должен быть целочисленным\nЕго можно получить при создании опроса.\n\n```/vote <poll_id> <option_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		optionId, err := strconv.Atoi(optionIdStr)
		if err != nil {
			err := handler.postToChannel("Параметр <option_id> должен быть целочисленным\nОн пишется рядом с вариантом ответа.\n\n```/vote <poll_id> <option_id>```", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		out, err := handler.Service.Vote(userId, uint(pollId), uint(optionId))
		if err != nil {
			err := handler.postToChannel(out, userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error())
				res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}
		err = handler.postToChannel(out, userId, channelId)
		if err != nil {
			handler.Logger.Error(err.Error())
			res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (handler *Handler) GetById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opt := "Handler.GetById: "
		query := r.URL.Query()
		text := query.Get("text")
		userId := query.Get("user_id")
		channelId := query.Get("channel_id")
		id, err := strconv.Atoi(strings.Split(text, " ")[0])
		if err != nil {
			err := handler.postToChannel("Вы передали неправильный id.", userId, channelId)
			if err != nil {
				handler.Logger.Error(err.Error(), opt+"handler.postToChannel")
			}
			return
		}
		response := handler.Service.GetPollById(uint(id), userId)
		err = handler.postToChannel(response, userId, channelId)
		if err != nil {
			handler.Logger.Error(err.Error(), slog.Int("poll_id", id))
			res.Json(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	}
}

func (handler *Handler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		triggerID := r.FormValue("trigger_id")
		reqBody := models.OpenDialogRequest{
			TriggerID: triggerID,
			URL:       handler.Config.GetHttpUrlBot() + string(constants.SubmitCreate),
			Dialog: models.Dialog{
				CallbackID:     "create_poll",
				Title:          "Создание опроса",
				Introduction:   "Пожалуйста, заполните поля ниже",
				SubmitLabel:    "Создать",
				NotifyOnCancel: false,
				Elements: []models.DialogElement{
					{
						DisplayName: "Вопрос",
						Name:        "question",
						Type:        "text",
						Placeholder: "Например, \"Какой язык лучше?\"",
					},
					{
						DisplayName: "Варианты (через запятую)",
						Name:        "options",
						Type:        "text",
						Placeholder: "Go, Rust, Python...",
						HelpText:    "Перечислите несколько вариантов через запятую",
					},
				},
			},
		}
		err := openDialog(handler.Config.Mattermost.Url, handler.Config.Bot.Token, reqBody)
		if err != nil {
			res.Json(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (handler *Handler) CreateSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ds models.DialogSubmission
		err := json.NewDecoder(r.Body).Decode(&ds)
		if err != nil {
			res.Json(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if ds.CallbackID == "create_poll" {
			question := ds.Submission["question"]
			options := ds.Submission["options"]
			userID := ds.UserID
			channelID := ds.ChannelID
			opts := strings.Split(options, ",")
			var optsFormat []string
			for i := range opts {
				opts[i] = strings.TrimSpace(opts[i])
				if opts[i] == "" {
					continue
				}
				runes := []rune(opts[i])
				runes[0] = unicode.ToUpper(runes[0])
				optsFormat = append(optsFormat, string(runes))
			}

			id, err := handler.Service.Create(userID, question, optsFormat)
			if err != nil {
				err := handler.postToChannel("Что-то пошло не так, попробуйте еще раз.", userID, channelID)
				if err != nil {
					handler.Logger.Error(err.Error(), slog.String("location", "handler.postToChannel"))
				}
			}
			for i := 0; i < len(optsFormat); i++ {
				optsFormat[i] = fmt.Sprintf("**%d.** %s", i+1, optsFormat[i])
			}
			err = handler.postToChannel(
				fmt.Sprintf("Опрос создан!\n**id: *%d***\n### %s\n%s\n\n*Вы можете узнать результаты командой*:\n```check <id>```", *id, question, strings.Join(optsFormat, "\n")), userID, channelID)

			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			return
		}
		res.Json(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (handler *Handler) postToChannel(message, userId, channelId string) error {
	postData := map[string]interface{}{
		"user_id": userId,
		"post": map[string]string{
			"channel_id": channelId,
			"message":    message,
		},
	}
	jsonBytes, err := json.Marshal(postData)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", handler.Config.Mattermost.Url+"/api/v4/posts/ephemeral", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+handler.Config.Bot.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}
	return nil
}

func openDialog(mattermostURL, token string, reqBody models.OpenDialogRequest) error {
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", mattermostURL+"/api/v4/actions/dialogs/open", strings.NewReader(string(jsonBytes)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openDialog: unexpected status %d", resp.StatusCode)
	}

	return nil
}
