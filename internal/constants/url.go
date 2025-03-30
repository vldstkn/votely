package constants

type URL string

const prefix = "/api"
const (
	Create       URL = prefix + "/create"
	Vote         URL = prefix + "/vote"
	GetById      URL = prefix + "/"
	EndVoting    URL = prefix + "/end"
	DeleteVoting URL = prefix + "/delete"
	SubmitCreate URL = prefix + "/create/submit"
)
