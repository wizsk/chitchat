package main

import (
	"sync"

	cdb "chithat/db"

	"golang.org/x/net/websocket"
)

type userAndConn struct {
	u *cdb.User
	c *websocket.Conn
}

type onlineClients struct {
	l sync.RWMutex
	m map[int]userAndConn
}

var online = onlineClients{m: make(map[int]userAndConn)}

func (o *onlineClients) add(u *cdb.User, conn *websocket.Conn) {
	o.l.Lock()
	defer o.l.Unlock()

	o.m[u.Id] = userAndConn{u, conn}
}

func (o *onlineClients) remove(userid int) {
	o.l.Lock()
	defer o.l.Unlock()

	delete(o.m, userid)
}

func (o *onlineClients) onlie() {
	o.l.RLock()
	defer o.l.RUnlock()
}

func (o *onlineClients) sendMsg(userid int, msg []byte) {
	o.l.RLock()
	defer o.l.RUnlock()

	uc, ok := o.m[userid]
	if !ok {
		return
	}
	uc.c.Write(msg)
}
