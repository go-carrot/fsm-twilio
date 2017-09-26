package fsmtwilio

type MessageReceivedCallback struct {
	ApiVersion string
	AccountSid string

	NumSegments int
	NumMedia    string

	MessageSid          string
	MessagingServiceSid string

	SmsMessageSid string
	SmsSid        string
	SmsStatus     string

	ToCountry string
	ToState   string
	ToCity    string
	ToZip     string

	FromCountry string
	FromState   string
	FromCity    string
	FromZip     string

	To   string
	From string
	Body string
}
