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

    room := r.URL.Query().Get("room")
	if room == "" {
	   room = "default"
	 }	

	 if rooms[room] == nil {
		 rooms[room] = make(map[*websocket.Conn]bool)
	 }

	 rooms[room][ws] = true
	
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("메시지 읽기 오류: %v", err)
			delete(rooms[room], ws)
			break
		}
		msg.Room = room
		broadcast <- msg
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
