package main

import (
	"github.com/gorilla/websocket"
  r "github.com/dancannon/gorethink"
  "log"
)

type FindHandler func(string) (Handler, bool)

type Client struct {
	send   chan Message
	socket *websocket.Conn
  findHandler FindHandler
  session *r.Session
  stopChannels map[int]chan bool
  id string
  userName string
}

func (c *Client) NewStopChannel(stopKey int) chan bool {
  c.StopForKey(stopKey)
  stop := make(chan bool)
  c.stopChannels[stopKey] = stop
  return stop
}

func (c *Client) StopForKey(key int) {
  if ch, found := c.stopChannels[key]; found {
    ch <- true
    delete(c.stopChannels, key)
  }
}

func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}
    if handler, found := client.findHandler(message.Name); found {
      handler(client, message.Data)
    }
	}
  client.socket.Close()
}

func (client *Client) Write() {
	for msg := range client.send {
		if err := client.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	client.socket.Close()
}

func (client *Client) Close() {
  for _, ch := range client.stopChannels {
    ch <- true
  }
  close(client.send)
  r.Table("user").Get(client.id).Delete().Exec(client.session)
}

func NewClient(socket *websocket.Conn, findHandler FindHandler, session *r.Session) *Client {
  var user User
  user.Name = "anonymous"
  response, err := r.Table("user").Insert(user).RunWrite(session)
  if err != nil {
    log.Println(err.Error())
  }
  var id string
  if len(response.GeneratedKeys) > 0 {
    id = response.GeneratedKeys[0]
  }
	return &Client{
		send:   make(chan Message),
		socket: socket,
    findHandler: findHandler,
    session: session,
    stopChannels: make(map[int]chan bool),
    id: id,
    userName: user.Name,
	}
}
