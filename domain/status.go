package domain

type AppStatus struct {
	AppStatus  string `json:"status"`
	DbStatus   string `json:"db_status"`
	AmqpStatus string `json:"amqp_status"`
	TextError  string `json:"text_error,omitempty"`
}
