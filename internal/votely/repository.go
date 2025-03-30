package votely

import (
	"errors"
	"fmt"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/vldstkn/votely/internal/models"
	"github.com/vldstkn/votely/pkg/db"
	"log/slog"
	"strings"
)

type RepositoryDeps struct {
	Logger *slog.Logger
	Db     *db.TrntlDb
}
type Repository struct {
	Logger *slog.Logger
	Db     *db.TrntlDb
}

func NewRepository(deps *RepositoryDeps) *Repository {
	return &Repository{
		Logger: deps.Logger,
		Db:     deps.Db,
	}
}

func (repo *Repository) Create(userId, title string, options []string) (*uint, error) {
	opt := "Repository.Create: "
	luaCode := `
        local userId, title, options = ...
        
        box.begin()
        
        local poll = box.space.polls:insert{
            nil, 
            userId,
            title,
            true,
            tonumber(tostring(os.time()))
        }
        
        for i, row in ipairs(options) do
            box.space.poll_options:insert{
                poll.id,
                i, 
                row,
                0
            }
        end
        
        box.commit()
        return poll.id
    `
	args := []interface{}{
		userId,
		title,
		options,
	}

	resp, err := repo.Db.Do(tarantool.NewEvalRequest(luaCode).Args(args)).Get()
	if err != nil {
		repo.Logger.Error(err.Error(), slog.String("location", opt+"eval transaction"))
		return nil, err
	}
	if len(resp) < 1 {
		repo.Logger.Error("Empty response", slog.String("location", opt+"response handling"))
		return nil, errors.New("transaction failed")
	}
	rawId, ok := resp[0].(int8)

	if !ok {
		repo.Logger.Error("Invalid ID type",
			slog.Any("actual_type", fmt.Sprintf("%T", resp[0])),
			slog.String("location", opt+"type conversion"))
		return nil, errors.New("invalid poll ID format")
	}

	id := uint(rawId)
	return &id, nil
}

func (repo *Repository) GetPollWithOptions(pollId uint) (*models.Poll, []models.PollOptions) {
	opt := "Repository.GetPollWithOptions: "

	luaCode := `
        local poll_id = ...
        
        if not box.space.polls or not box.space.poll_options then
            error("SPACES_NOT_FOUND")
        end
        
        local poll = box.space.polls:get{poll_id}
        if not poll then
            return nil
        end
        
        local options = box.space.poll_options.index.poll_id:select{poll_id}
        
        return {
            poll = poll,
            options = options
        }
    `

	req := tarantool.NewEvalRequest(luaCode).Args([]interface{}{pollId})

	var result []struct {
		Poll    models.Poll          `msgpack:"poll"`
		Options []models.PollOptions `msgpack:"options"`
	}

	err := repo.Db.Do(req).GetTyped(&result)
	if err != nil {
		repo.Logger.Error(err.Error(),
			slog.String("location", opt),
			slog.Int("poll_id", int(pollId)),
		)
		return nil, nil
	}
	if len(result) == 0 {
		return nil, nil
	}

	if result[0].Poll.Id == 0 {
		return nil, nil
	}

	return &result[0].Poll, result[0].Options
}

func (repo *Repository) GetPollById(id uint) *models.Poll {
	opt := "Repository.GetPollById: "
	luaCode := `
        local id = ...
        local result = box.space.polls:select{id}
        if #result == 0 then
            return nil
        end
        return result[1]  
    `

	req := tarantool.NewEvalRequest(luaCode).Args([]interface{}{id})

	var polls []models.Poll
	err := repo.Db.Do(req).GetTyped(&polls)
	if err != nil {
		repo.Logger.Warn(err.Error(),
			slog.String("location", opt+" repo.Db.Do(req).GetTyped"),
			slog.Int("id", int(id)),
		)
		return nil
	}

	return &polls[0]
}
func (repo *Repository) GetPollOptions(id uint) []models.PollOptions {
	opt := "Repository.GetPollOptions: "
	luaCode := `
        local poll_id = ...
        
        if not box.space.poll_options then
            error("Space poll_options not found")
        end
        
        local result = box.space.poll_options.index.poll_id:select{poll_id}
        return result
    `
	req := tarantool.NewEvalRequest(luaCode).Args([]interface{}{id})
	var pollsOpt [][]models.PollOptions
	err := repo.Db.Do(req).GetTyped(&pollsOpt)
	if err != nil {
		repo.Logger.Warn(err.Error(),
			slog.String("location", opt+" tarantool.NewEvalRequest"),
			slog.Int("id", int(id)),
		)
		return nil
	}

	if len(pollsOpt) == 0 {
		return nil
	}

	return pollsOpt[0]
}
func (repo *Repository) Vote(vote *models.Vote) error {
	opt := "Repository.Vote: "
	luaCode := `
        local vote = ...
        
        if not box.space.votes or not box.space.poll_options then
            error("SPACES_NOT_EXIST")
        end
        
        if type(vote.poll_id) ~= 'number' or
           type(vote.user_id) ~= 'string' or
           type(vote.option_id) ~= 'number' then
            error("INVALID_DATA_TYPES")
        end
        
        box.begin()
        
        local poll = box.space.polls:get{vote.poll_id}
        if not poll then
            box.rollback()
            error("POLL_NOT_FOUND")
        end
        if not poll.is_active then
            box.rollback()
            error("POLL_INACTIVE")
        end

        local option = box.space.poll_options:get{vote.poll_id, vote.option_id}
        if not option then
            box.rollback()
            error("OPTION_NOT_FOUND")
        end
        

        local old_vote = box.space.votes:get{vote.poll_id, vote.user_id}
        local old_option_id = nil
        
        if old_vote ~= nil then
            old_option_id = old_vote.option_id
            local old_option = box.space.poll_options:get{vote.poll_id, old_option_id}
            if old_option then
                box.space.poll_options:update(
                    {vote.poll_id, old_option_id},
                    {{'-', 'votes_count', 1}}
                )
            end
        end
        
        box.space.poll_options:update(
            {vote.poll_id, vote.option_id},
            {{'+', 'votes_count', 1}}
        )
        
        local new_vote = {
            vote.poll_id,
            vote.user_id,
            vote.option_id,
            os.time()
        }
        
        if old_vote then
            box.space.votes:replace(new_vote)
        else
            box.space.votes:insert(new_vote)
        end
        
        box.commit()
        return true
    `

	data := map[string]interface{}{
		"poll_id":   vote.PollId,
		"user_id":   vote.UserId,
		"option_id": vote.OptionId,
	}

	req := tarantool.NewEvalRequest(luaCode).Args([]interface{}{data})
	_, err := repo.Db.Do(req).Get()

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "VOTE_EXISTS"):
			return errors.New("вы уже голосовали в этом опросе")
		case strings.Contains(err.Error(), "OPTION_NOT_FOUND"):
			return errors.New("указанный вариант ответа не найден")
		case strings.Contains(err.Error(), "POLL_NOT_FOUND"):
			return errors.New("опрос не найден")
		case strings.Contains(err.Error(), "POLL_INACTIVE"):
			return errors.New("опрос уже завершен, голоса больше не принимаются")
		case strings.Contains(err.Error(), "SPACES_NOT_EXIST"):
			repo.Logger.Error("SPACES_NOT_EXIST"+err.Error(), slog.String("location", opt))
			return errors.New("внутренняя ошибка базы данных")
		case strings.Contains(err.Error(), "INVALID_DATA_TYPES"):
			return errors.New("неверный формат данных")
		default:
			repo.Logger.Error(err.Error(), slog.String("location", opt))
			return fmt.Errorf("ошибка голосования: %v", err)
		}
	}

	return nil
}
func (repo *Repository) GetVote(userId string, pollId uint) *models.Vote {
	opt := "Repository.GetVote: "
	luaCode := `
        local poll_id, user_id = ...
        local result = box.space.votes:select{poll_id, user_id}
        if #result == 0 then
            return nil
        end
        return result[1]  
    `
	req := tarantool.NewEvalRequest(luaCode).Args([]interface{}{pollId, userId})

	var votes []models.Vote
	err := repo.Db.Do(req).GetTyped(&votes)
	if err != nil {
		repo.Logger.Warn(err.Error(),
			slog.String("location", opt+" repo.Db.Do(req).GetTyped"),
			slog.Int("poll_id", int(pollId)),
			slog.String("user_id", userId),
		)
		return nil
	}
	if len(votes) == 0 {
		return nil
	}
	return &votes[0]
}
func (repo *Repository) EndVoting(userId string, pollId uint) error {
	luaCode := `
        local poll_id, user_id = ...
        
        if not box.space.polls then
            error("POLLS_SPACE_NOT_FOUND")
        end
        
        if type(poll_id) ~= 'number' or 
           type(user_id) ~= 'string' then
            error("INVALID_ARGUMENT_TYPES")
        end
        
        box.begin()
        
        local poll = box.space.polls:get(poll_id)
        if not poll then
            box.rollback()
            error("POLL_NOT_FOUND")
        end
        
        if poll.creator_id ~= user_id then
            box.rollback()
            error("PERMISSION_DENIED")
        end
        
        local updated = box.space.polls:update(
            poll_id,
            {{'=', 'is_active', false}}
        )
        
        box.commit()
        return updated
    `

	req := tarantool.NewEvalRequest(luaCode).Args([]interface{}{pollId, userId})
	_, err := repo.Db.Do(req).Get()

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "POLL_NOT_FOUND"):
			return errors.New("опрос не найден")
		case strings.Contains(err.Error(), "PERMISSION_DENIED"):
			return errors.New("недостаточно прав для деактивации опроса")
		case strings.Contains(err.Error(), "POLLS_SPACE_NOT_FOUND"):
			return errors.New("таблица опросов не существует")
		case strings.Contains(err.Error(), "INVALID_ARGUMENT_TYPES"):
			return errors.New("неверный формат аргументов")
		default:
			return fmt.Errorf("ошибка деактивации: %v", err)
		}
	}

	return nil
}
func (repo *Repository) DeleteVoting(creatorId string, pollId uint) error {
	luaCode := `
        local poll_id, user_id = ...
        
        if not box.space.polls or 
           not box.space.poll_options or 
           not box.space.votes then
            error("SPACES_NOT_FOUND")
        end
        
        if type(poll_id) ~= 'number' or 
           type(user_id) ~= 'string' then
            error("INVALID_ARGUMENT_TYPES")
        end
        
        box.begin()
        
        local poll = box.space.polls:get(poll_id)
        if not poll then
            box.rollback()
            error("POLL_NOT_FOUND")
        end
        
        if poll.creator_id ~= user_id then
            box.rollback()
            error("PERMISSION_DENIED")
        end
        
        local options = box.space.poll_options.index.poll_id:select(poll_id)
        for _, option in ipairs(options) do
            box.space.poll_options:delete{option.poll_id, option.option_id}
        end
        
        local votes = box.space.votes.index.by_option:select(poll_id, {iterator='GE'})
        for _, vote in ipairs(votes) do
            box.space.votes:delete{vote.poll_id, vote.user_id}
        end
        
        box.space.polls:delete(poll_id)
        
        box.commit()
        return true
    `

	req := tarantool.NewEvalRequest(luaCode).Args([]interface{}{pollId, creatorId})
	_, err := repo.Db.Do(req).Get()

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "POLL_NOT_FOUND"):
			return errors.New("опрос не найден")
		case strings.Contains(err.Error(), "PERMISSION_DENIED"):
			return errors.New("недостаточно прав для удаления опроса")
		case strings.Contains(err.Error(), "SPACES_NOT_FOUND"):
			return errors.New("ошибка сервера")
		case strings.Contains(err.Error(), "INVALID_ARGUMENT_TYPES"):
			return errors.New("неверный формат аргументов")
		default:
			repo.Logger.Error(err.Error(), "Repository.DeleteVoting")
			return fmt.Errorf("ошибка удаления")
		}
	}

	return nil
}
