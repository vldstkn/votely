package models

type OpenDialogRequest struct {
	TriggerID string `json:"trigger_id"` // из slash-команды
	URL       string `json:"url"`        // куда вернётся форма при submit
	Dialog    Dialog `json:"dialog"`
}

type Dialog struct {
	CallbackID     string          `json:"callback_id"`
	Title          string          `json:"title"`
	Introduction   string          `json:"introduction,omitempty"`
	Elements       []DialogElement `json:"elements"`
	SubmitLabel    string          `json:"submit_label"`
	NotifyOnCancel bool            `json:"notify_on_cancel"`
	State          string          `json:"state,omitempty"`
}

type DialogElement struct {
	DisplayName string `json:"display_name"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Optional    bool   `json:"optional,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	HelpText    string `json:"help_text,omitempty"`
}

type DialogSubmission struct {
	Type       string            `json:"type"`        // "submission"
	CallbackID string            `json:"callback_id"` // "create_poll"
	UserID     string            `json:"user_id"`
	ChannelID  string            `json:"channel_id"`
	Submission map[string]string `json:"submission"`
}
