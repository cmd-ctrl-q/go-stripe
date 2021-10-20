package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	*websocket.Conn
}

// Payload that will hold data received from the client
type WsPayload struct {
	Action      string              `json:"action"`
	Message     string              `json:"message"`
	UserName    string              `json:"username"`
	MessageType string              `json:"message_type"`
	UserID      int                 `json:"user_id"`
	Conn        WebSocketConnection `json:"-"`
}

type WsJsonResponse struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

// connection to webserver is upgraded to permit two way communication
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// To secure websocket connections
	CheckOrigin: func(r *http.Request) bool { return true },
}

// keeps track of every client connected
var clients = make(map[WebSocketConnection]string)

// stores data that is pushed to this channel
var wsChan = make(chan WsPayload)

func (app *application) WsEndPoint(w http.ResponseWriter, r *http.Request) {
	// upgrade connection
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// write to log
	app.infoLog.Println(fmt.Sprintf("Client connected from %s", r.RemoteAddr))

	var resp WsJsonResponse
	resp.Message = "Connected to server"

	err = ws.WriteJSON(resp)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// get connection
	conn := WebSocketConnection{Conn: ws}
	// add connection to map
	clients[conn] = ""

	go app.ListenForWS(&conn)
}

func (app *application) ListenForWS(conn *WebSocketConnection) {
	defer func() {
		// recover gracefully
		if r := recover(); r != nil {
			app.errorLog.Println("ERROR:", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		// execute forever
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing
		} else {
			payload.Conn = *conn
			// send connection off to channel
			wsChan <- payload
		}
	}
}

// ListenToWsChannel waits for input on channel then takes an action based the response
func (app *application) ListenToWsChannel() {
	var resp WsJsonResponse
	for {
		e := <-wsChan
		// add action
		switch e.Action {
		case "deleteUser":
			// tell end user to logout
			resp.Action = "logout"
			resp.Message = "Your account has been deleted"
			resp.UserID = e.UserID
			// broadcase to all clients
			app.broadcastToAll(resp)

		default:

		}
	}
}

func (app *application) broadcastToAll(resp WsJsonResponse) {
	for client := range clients {
		// broadcast to every connected client
		err := client.WriteJSON(resp)
		if err != nil {
			app.errorLog.Printf("Websocket err on %s: %s", resp.Action, err)
			// client disconnected
			_ = client.Close()
			// remove entry from map
			delete(clients, client)
		}
	}
}
