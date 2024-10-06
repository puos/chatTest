package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Room     string `json:"room"`
	UserName string `json:"username"`
	Message  string `json:"message"`
}

var rooms = make(map[string]map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var upgrader = websocket.Upgrader{}

func main() {

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("서버 시작: http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("서버 시작 실패: ", err)
	}

}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	var roomName string
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("메시지 읽기 오류: %v", err)
			break
		}
		if roomName == "" {
			roomName = msg.Room
			if rooms[roomName] == nil {
				rooms[roomName] = make(map[*websocket.Conn]bool)
			}
			rooms[roomName][ws] = true
		}
		broadcast <- msg
	}

	if roomName != "" {
		delete(rooms[roomName], ws)
		ws.Close()
	}
}

func handleMessages() {

	for {

		msg := <-broadcast
		roomName := msg.Room

		for client := range rooms[roomName] {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("메시지 전송 오류: %v", err)
				client.Close()
				delete(rooms[roomName], client)
			}
		}
	}
}
