package domain

type NotificationChannel int

const (
	EmailChannel NotificationChannel = iota
	SMSChannel
)

func (nc NotificationChannel) String() string {
	switch nc {
	case EmailChannel:
		return "email"
	case SMSChannel:
		return "sms"
	default:
		return "unknown"
	}
}

type NotificationRequest struct {
	User    *User
	Channel NotificationChannel
}

func NewNotificationRequest(user *User, channel NotificationChannel) *NotificationRequest {
	return &NotificationRequest{
		User:    user,
		Channel: channel,
	}
}
