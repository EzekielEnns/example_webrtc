package main

import (
	//	"fmt"
	"fmt"
	"log"
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
            noAwnser := true
            log.Println("LOOKING FOR ANWSER#########################")
            for noAwnser {
                _,anwser,_ := peers.oldPeers[i].ReadMessage()
                str := string(anwser)
                log.Println("::::::::::::::::::::::::anwser\n",string(anwser))
                conn.WriteMessage(websocket.TextMessage, []byte(string(anwser)))
                if strings.Contains(str,`"type":"answer"`) {
                    
                log.Println("got it#########################")
                    noAwnser = false
                }
            }
            newPeerState, oldPeerState := false, false
            for !(newPeerState && oldPeerState){
                if (!newPeerState){
                    _, ice, _ := conn.ReadMessage()
                    if (string(ice)!="done"){
                        peers.oldPeers[i].WriteMessage(websocket.TextMessage, []byte(string(ice)))
                    } else {
                        newPeerState = true
                    }
                }
                if (!oldPeerState) {
                    _,ice,_ := peers.oldPeers[i].ReadMessage()
                    log.Println("::::::::::::::::::::::::OldPeer\n",string(ice))
                    if (string(ice)!="done"){
                        conn.WriteMessage(websocket.TextMessage, []byte(string(ice)))
                    }else {
                        oldPeerState = true
                    }
                }
            }
        }
        
    }
    log.Println("done")
    peers.oldPeers[peers.connections] = conn
    log.Println(peers.connections)
    peers.connections++
    peers.newPeer.Unlock()
    log.Println(peers.connections)

    for {
        if peers.connections == MAX_CONN {
            return
        }
    }
}
