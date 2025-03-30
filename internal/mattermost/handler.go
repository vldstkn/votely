package mattermost

import (
	"github.com/mattermost/mattermost-server/v6/model"
)

func RegisterVoteCommand(client *model.Client4, teamName, botUrl string) error {

	team, _, err := client.GetTeamByName(teamName, "")
	if err != nil {
		return err
	}

	for _, cmd := range GetCommands(team.Id, botUrl, client.AuthToken) {
		_, _, err := client.CreateCommand(cmd)
		if err != nil {
			continue
		}
	}
	return nil
}
