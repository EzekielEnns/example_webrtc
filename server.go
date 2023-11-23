package main

import (
	//	"fmt"
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
var ch = make(chan string)
const MAX_CONN = 5

func main() {
    peers = Peers{oldPeers: make([]*websocket.Conn, MAX_CONN),connections: 0}
    //when a Request is made to the path "/" server those files 
    http.HandleFunc("/",http.FileServer(http.Dir("./dist")).ServeHTTP)

    //spins up a new "thread" (goroutine) where handleSignaling() is run on new connection
    http.HandleFunc("/ws", handleSignaling)

    //listen for requests to spin "threads" up
	http.ListenAndServe(":8080", nil)
}

func handleSignaling(w http.ResponseWriter, r *http.Request) {
    conn, _ := toWebSocket.Upgrade(w,r,nil)

    //defers the call of conn close once this func ends
    defer conn.Close()

    if peers.connections >= MAX_CONN {
        log.Println("BYE")
        return
    }
    
    peers.newPeer.Lock()
    if peers.connections > 0 {
        log.Println("initalizing....")
        //let peer know they are ready 
        //reciving sdp info
        //note most cases would use json here as its more flexable
        //broadcast
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
                    log.Println("New peer done")
                    break
                } else {

                    peers.oldPeers[i].WriteMessage(websocket.TextMessage, []byte(string(ice)))
                    conn.WriteMessage(websocket.TextMessage, []byte(string(<-ch)))
                }
            }
            log.Println("next peer ",i)
        }
        
    }
    peers.oldPeers[peers.connections] = conn
    peers.connections++
    log.Println("conns ",peers.connections)
    peers.newPeer.Unlock()

    for {
        if peers.connections == MAX_CONN {
            log.Println("Closing")
            return
        }
        _,msg,_ := conn.ReadMessage()
        log.Println(string(msg))
        str := string(msg)
        if (strings.Contains(str,`candidate":"",`)){
            log.Println("skipped bad ice")
            continue
        }
        if (str != "done") {
            ch <- str
        }
    }
}
