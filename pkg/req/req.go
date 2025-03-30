package req

import (
	"encoding/json"
	"errors"
	"github.com/mattermost/mattermost-server/v6/model"
)

func GetPost(event *model.WebSocketEvent) (*model.Post, error) {
	rawPost, ok := event.GetData()["post"].(string)
	if !ok {
		return nil, errors.New("bad post")
	}
	var post model.Post
	if err := json.Unmarshal([]byte(rawPost), &post); err != nil {
		return nil, errors.New("bad data")
	}
	return &post, nil
}
