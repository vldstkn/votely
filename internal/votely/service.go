package votely

import (
	"fmt"
	"github.com/vldstkn/votely/internal/interfaces"
	"github.com/vldstkn/votely/internal/models"
	"log/slog"
	"strings"
	"unicode"
)

type ServiceDeps struct {
	Logger     *slog.Logger
	Repository interfaces.Repository
}
type Service struct {
	Logger     *slog.Logger
	Repository interfaces.Repository
}

func NewService(deps *ServiceDeps) *Service {
	return &Service{
		Logger:     deps.Logger,
		Repository: deps.Repository,
	}
}

func (service *Service) Create(userId, title string, options []string) (*uint, error) {
	id, err := service.Repository.Create(userId, title, options)
	return id, err
}
func (service *Service) GetPollById(pollId uint, userId string) string {
	poll, options := service.Repository.GetPollWithOptions(pollId)
	if poll == nil {
		return "Нет опроса с таким id.\nПроверьте id и повторите попытку"
	}
	vote := service.Repository.GetVote(userId, pollId)
	out := service.buildPoll(*poll, options, vote)
	return out
}

func (service *Service) EndVoting(userId string, pollId uint) string {
	err := service.Repository.EndVoting(userId, pollId)
	if err != nil {
		runes := []rune(err.Error())
		runes[0] = unicode.ToUpper(runes[0])
		return string(runes)
	}
	poll, options := service.Repository.GetPollWithOptions(pollId)
	if poll == nil {
		return "Получите результаты голосования командой\n```/check <poll_id>```"
	}
	txt := fmt.Sprintf("## Результаты голосования\n%s", service.buildPoll(*poll, options, nil))

	return txt
}
func (service *Service) DeleteVoting(creatorId string, pollId uint) string {
	err := service.Repository.DeleteVoting(creatorId, pollId)
	if err != nil {
		runes := []rune(err.Error())
		runes[0] = unicode.ToUpper(runes[0])
		return string(runes)
	}
	return "Голосование успешно удалено."
}

func (service *Service) Vote(userId string, pollId, optionId uint) (string, error) {
	vote := &models.Vote{
		PollId:   pollId,
		UserId:   userId,
		OptionId: optionId,
	}
	err := service.Repository.Vote(vote)
	if err != nil {
		runes := []rune(err.Error())
		runes[0] = unicode.ToUpper(runes[0])
		return string(runes), err
	}
	poll, options := service.Repository.GetPollWithOptions(pollId)
	out := service.buildPoll(*poll, options, vote)
	return out, nil
}

func (service *Service) buildPoll(poll models.Poll, options []models.PollOptions, vote *models.Vote) string {
	var allVotes uint
	maxLen := 0
	res := make([]string, len(options))
	var voteExists bool
	if vote == nil {
		voteExists = false
	} else {
		voteExists = true
	}
	for i := 0; i < len(options); i++ {
		allVotes += options[i].VotesCount
		options[i].OptionText = fmt.Sprintf("**%d. **%s", options[i].OptionId, options[i].OptionText)
		if len(options[i].OptionText) > maxLen {
			maxLen = len(options[i].OptionText)
		}
	}
	for i := 0; i < len(options); i++ {
		opt := options[i]
		var votes float32
		if allVotes == 0 {
			votes = 0
		} else {
			votes = (float32(opt.VotesCount) / float32(allVotes)) * 100
		}
		res[i] = fmt.Sprintf("%-*s  %.2f%%    %d",
			maxLen, opt.OptionText, votes, opt.VotesCount)
		if voteExists && int(vote.OptionId) == i+1 {
			res[i] += "\t***ваш выбор***"
		}
	}
	var status string
	if poll.IsActive {
		status = "активный"
	} else {
		status = "завершен"
	}
	txt := fmt.Sprintf("### %s\n%s\n\n\n***%s***", poll.Question, strings.Join(res, "\n\n"), status)
	return txt
}
