package interfaces

import "github.com/vldstkn/votely/internal/models"

type Service interface {
	Create(userId, title string, options []string) (*uint, error)
	GetPollById(id uint, userId string) string
	Vote(userID string, pollId, optionId uint) (string, error)
	EndVoting(userId string, pollId uint) string
	DeleteVoting(creatorId string, pollId uint) string
}

type Repository interface {
	Create(userId, title string, options []string) (*uint, error)
	GetPollById(id uint) *models.Poll
	GetPollOptions(id uint) []models.PollOptions
	Vote(vote *models.Vote) error
	GetVote(userId string, pollId uint) *models.Vote
	GetPollWithOptions(pollId uint) (*models.Poll, []models.PollOptions)
	EndVoting(userId string, pollId uint) error
	DeleteVoting(creatorId string, pollId uint) error
}
