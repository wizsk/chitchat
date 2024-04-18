package main

import cdb "chithat/db"

type WsData struct {
	DataType string        `json:"data_type"`
	Ping     string        `json:"ping,omitempty"`
	User     *cdb.User     `json:"user,omitempty"`
	Message  *cdb.Message  `json:"message,omitempty"`
	Inbox    []cdb.Message `json:"inbox,omitempty"`
}

const (
	WsDataPing uint8 = iota
	WsDataUser
	WsDataMessage
	WsDataInbox
	WsDataArrLen
)

var wsDataNames = [WsDataArrLen]string{"ping", "user", "message", "inbox"}

func wsdt(dt uint8) string {
	if dt < WsDataArrLen {
		return wsDataNames[dt]
	}
	panic("invalid ws data type")
}
