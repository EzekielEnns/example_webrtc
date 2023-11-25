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
        return
    }
    
    peers.newPeer.Lock()
    if peers.connections > 0 {
        log.Println("INITALIZING....")
        for i:=0; i<peers.connections; i++ {
            newIce := []string{}
            oldIce := []string{}
            nD,oD := true, true
            nIR,oIR := false,false
            conn.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"ready\"}"))
            for nD || oD {
                //processing new peer msgs
                if (nD) {
                    _, newPeer , _ := conn.ReadMessage()
                    str := string(newPeer)
                    log.Println("##########New:",str)
                    if strings.Contains(str,"offer"){
                        peers.oldPeers[i].WriteMessage(websocket.TextMessage, []byte(str))
                    }else if strings.Contains(str,"ready"){
                        nIR = true
                    } else if strings.Contains(str,"candidate") {
                        newIce = append(newIce, str)
                    } else if str == "done"{
                        nD = false
                    }
                }

                if nIR{
                    for _,ice := range oldIce {
                        conn.WriteMessage(websocket.TextMessage, []byte(ice))
                    }
                    oldIce = oldIce[:0]
                }

                //processing old peer
                if oD {
                    _, oldPeer, _ := peers.oldPeers[i].ReadMessage()
                    str := string(oldPeer)
                    log.Println("##########Old:",str)
                    if strings.Contains(str,"offer"){
                        conn.WriteMessage(websocket.TextMessage, []byte(str))
                    }else if strings.Contains(str,"answer"){
                        conn.WriteMessage(websocket.TextMessage, []byte(str))
                        oIR = true
                    } else if strings.Contains(str,"candidate") {
                        oldIce = append(oldIce, str)
                    } else if str == "done"{
                        oD = false
                    }

                }
                if oIR{
                    for _,ice := range newIce{
                        peers.oldPeers[i].WriteMessage(websocket.TextMessage, []byte(ice))
                    }

                    newIce = newIce[:0]
                }
            }

        }
        
    }
    peers.oldPeers[peers.connections] = conn
    peers.connections++
    log.Println("ADDED A NEW PEER",peers.connections)
    peers.newPeer.Unlock()

    for {
        if peers.connections == MAX_CONN {
            return
        }
        // _,msg,_ := conn.ReadMessage()
        // log.Println("got msg")
        // log.Println(string(msg))
        // str := string(msg)
        // if (str != "done" ) {
        //     ch <- str
        // }
        // if (strings.Contains(str,`candidate":"",`)){
        //     log.Println("skipped bad ice")
        // }
    }
}
