package models

type Poll struct {
	Id        uint   `msgpack:"id"`
	CreatorId string `msgpack:"creator_id"`
	Question  string `msgpack:"question"`
	IsActive  bool   `msgpack:"is_active"`
	CreatedAt uint   `msgpack:"created_at"`
}

type Vote struct {
	PollId    uint   `msgpack:"poll_id"`
	UserId    string `msgpack:"user_id"`
	OptionId  uint   `msgpack:"option_id"`
	CreatedAt uint   `msgpack:"created_at"`
}

type PollOptions struct {
	PollId     uint   `msgpack:"1"`
	OptionId   uint   `msgpack:"2"`
	OptionText string `msgpack:"3"`
	VotesCount uint   `msgpack:"4"`
}
