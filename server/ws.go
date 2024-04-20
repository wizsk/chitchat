package main

import cdb "chithat/db"

type WsData struct {
	DataType   string       `json:"data_type"`
	Error      string       `json:"error,omitempty"`
	SearchTerm string       `json:"search_term,omitempty"`
	User       *cdb.User    `json:"user,omitempty"`
	Message    *cdb.Message `json:"message,omitempty"`
	AllInboxes []*cdb.Inbox `json:"all_inboxes,omitempty"`
}

const (
	WsDataPing uint8 = iota
	WsDataUser
	WsDataGetInbox
	WsDataMessage
	WsDataMessageSend
	WsDataMessageReceive
	WsDataSearchUser
	WsDataArrLen
)

var wsDataNames = [WsDataArrLen]string{
	"ping", "user", "get_inbox",
	"message", "message_send", "message_receive",
	"search_user",
}

// ws data type
func wsdt(dt uint8) string {
	if dt < WsDataArrLen {
		return wsDataNames[dt]
	}
	panic("invalid ws data type")
}
