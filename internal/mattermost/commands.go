package mattermost

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/vldstkn/votely/internal/constants"
)

func GetCommands(teamId, botUrl, token string) []*model.Command {
	return []*model.Command{
		{
			TeamId:           teamId,
			Trigger:          "create",
			Method:           model.CommandMethodPost,
			AutoComplete:     true,
			AutoCompleteDesc: "Создать опрос",
			AutoCompleteHint: "",
			DisplayName:      "Создать опрос",
			Description:      "Создайте новый опрос",
			URL:              botUrl + string(constants.Create),
			Token:            token,
		},
		{
			TeamId:           teamId,
			Trigger:          "vote",
			Method:           model.CommandMethodPost,
			AutoComplete:     true,
			AutoCompleteDesc: "Проголосуйте за один из вариантов",
			AutoCompleteHint: "[poll_id option_id]",
			DisplayName:      "Проголосовать",
			Description:      "Проголосуйте за один из вариантов",
			URL:              botUrl + string(constants.Vote),
			Token:            token,
		},
		{
			TeamId:           teamId,
			Trigger:          "check",
			Method:           model.CommandMethodGet,
			AutoComplete:     true,
			AutoCompleteDesc: "Узнайте результаты голосования",
			AutoCompleteHint: "[poll_id]",
			DisplayName:      "Узнать результаты",
			Description:      "Узнайте результаты голосования",
			URL:              botUrl + string(constants.GetById),
			Token:            token,
		},
		{
			TeamId:           teamId,
			Trigger:          "end",
			Method:           model.CommandMethodPost,
			AutoComplete:     true,
			AutoCompleteDesc: "Завершите голосование досрочно",
			AutoCompleteHint: "[poll_id]",
			DisplayName:      "Завершить голосование",
			Description:      "Завершите голосование досрочно",
			URL:              botUrl + string(constants.EndVoting),
			Token:            token,
		},
		{
			TeamId:           teamId,
			Trigger:          "del",
			Method:           model.CommandMethodPost,
			AutoComplete:     true,
			AutoCompleteDesc: "Удалите голосование без подведения итогов",
			AutoCompleteHint: "[poll_id]",
			DisplayName:      "Удалить голосование",
			Description:      "Удалите голосование без подведения итогов",
			URL:              botUrl + string(constants.DeleteVoting),
			Token:            token,
		},
	}
}
