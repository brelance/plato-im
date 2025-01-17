package sdk

import "net"

type Chat struct {
	Nick      string
	UserId    string
	SessionId string
}

func NewChat(ip net.IP, port int, nick, userID, sessionID string) *Chat {
	chat := &Chat{
		Nick:      nick,
		UserId:    userID,
		SessionId: sessionID,
	}

	return chat
}
