package main

import (
	//	"fmt"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var toWebSocket = websocket.Upgrader{}

type Peers struct {
    newPeer sync.Mutex
    oldPeers []*websocket.Conn
    connections int
}

var peers Peers
var ch = make(chan string)
const MAX_CONN = 5

func main() {
    peers = Peers{oldPeers: make([]*websocket.Conn, MAX_CONN),connections: 0}
    //when a Request is made to the path "/" server those files 
    http.HandleFunc("/",http.FileServer(http.Dir("./dist")).ServeHTTP)

    //spins up a new "thread" (goroutine) where handleSignaling() is run on new connection
    http.HandleFunc("/ws", handleSignaling)

    fmt.Println("http://localhost:8080")
    //listen for requests to spin "threads" up
	http.ListenAndServe(":8080", nil)
}

func handleSignaling(w http.ResponseWriter, r *http.Request) {
    conn, _ := toWebSocket.Upgrade(w,r,nil)

    //defers the call of conn close once this func ends
    defer conn.Close()

    if peers.connections >= MAX_CONN {
        return
    }
    
    peers.newPeer.Lock()
    if peers.connections > 0 {
        //broad cast
        for i:=0; i<peers.connections; i++ {
            conn.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"ready\"}"))
            _, sdp , _ := conn.ReadMessage()
            //sends offer
            peers.oldPeers[i].WriteMessage(websocket.TextMessage, []byte(string(sdp)))
            //sends back awnser
            conn.WriteMessage(websocket.TextMessage, []byte(string(<-ch)))
            for {
            _, ice, _ := conn.ReadMessage()
                if (string(ice)=="done"){
                    break
                } else {

                    peers.oldPeers[i].WriteMessage(websocket.TextMessage, []byte(string(ice)))
                    conn.WriteMessage(websocket.TextMessage, []byte(string(<-ch)))
                }
            }
        }
        
    }
    peers.oldPeers[peers.connections] = conn
    peers.connections++
    peers.newPeer.Unlock()

    for {
        if peers.connections == MAX_CONN {
            return
        }
        _,msg,_ := conn.ReadMessage()
        str := string(msg)
        if (strings.Contains(str,`candidate":"",`)){
            continue
        }
        if (str != "done") {
            ch <- str
        }
    }
}
