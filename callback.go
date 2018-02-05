package fsmtwilio

type messageReceivedCallback struct {
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

type messageSentCallback struct {
	Message Message `xml:"Message"`
}

type message struct {
	Sid                 string          `xml:"Sid"`
	DateCreated         string          `xml:"DateCreated"`
	DateUpdated         string          `xml:"DateUpdated"`
	DateSent            string          `xml:"DateSent"`
	AccountSid          string          `xml:"AccountSid"`
	To                  string          `xml:"To"`
	From                string          `xml:"From"`
	MessagingServiceSid string          `xml:"MessagingServiceSid"`
	Body                string          `xml:"Body"`
	Status              string          `xml:"Status"`
	NumSegments         string          `xml:"NumSegments"`
	NumMedia            string          `xml:"NumMedia"`
	Direction           string          `xml:"Direction"`
	APIVersion          string          `xml:"ApiVersion"`
	Price               string          `xml:"Price"`
	PriceUnit           string          `xml:"PriceUnit"`
	ErrorCode           string          `xml:"ErrorCode"`
	ErrorMessage        string          `xml:"ErrorMessage"`
	URI                 string          `xml:"Uri"`
	SubresourceUris     SubresourceUris `xml:"SubresourceUris"`
}

type subresourceUris struct {
	Media string `xml:"Media"`
}
