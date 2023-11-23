/**
 * @typedef {Object} Msg
 * @property {string} type  - type of message/how to processes
 * @property {string} value - value for message
 * @property {boolean} offer - if true is a offer if false is a anwser
 */
//TODO fix  DOMException: No remoteDescription.
    //TODO Uncaught (in promise) SyntaxError: JSON.parse: unexpected character at line 1 column 1 of the JSON data
/*
 * initalize connection to singalling server
 *     handle messages from server
 * */

//connect to singalling server
const singalling = new WebSocket("ws://localhost:8080/ws");
//TODO init peers

/**
 * @type {Array<RTCPeerConnection>}
 */
const peers = Array.from({length:4}, createPeer)
//5 connections each peer is connected to another peer ecluding itself

/**
 * a peer tobe added to peers
 * @type {number}
 */
var index = 0;

singalling.addEventListener("message", async function(event){
    console.log(event.data)
    //check for done
    let data = JSON.parse(event.data)
    console.log(data)

    switch (data.type){
        case "ready":
            const offer = await peers[index].createOffer()
            await peers[index].setLocalDescription(offer)
            console.log(offer)
            singalling.send(JSON.stringify( peers[index].localDescription ))
            break;
        case "offer":
            console.log("offer")
            await peers[index].setRemoteDescription(data)
            const awnser = await peers[index].createAnswer()
            await peers[index].setLocalDescription(awnser)
            singalling.send(JSON.stringify( peers[index].localDescription))
            break;
        case "answer":
            console.log("answer")
            console.log(data)
            await peers[index].setRemoteDescription(data)
            break;
        default:
            console.log("default")
            if (data.candidate){
                console.log("ice")
                console.log(data)
                peers[index].addIceCandidate(data)
            }
            break;
    }
})

/**
 * @returns {RTCPeerConnection}
 */
function createPeer(){
    let peer = new RTCPeerConnection({
        iceServers: [{ urls: "stun:stun.l.google.com:19302"}]
    })
    peer.createDataChannel("text").addEventListener('message',e => {
        console.log(e.data)
    })
    peer.addEventListener("icecandidate", function(event) {
        console.log("Got candidate:::::",event)
        if (event.candidate) {
            singalling.send(JSON.stringify(event.candidate))
        }
    });
    peer.addEventListener("iceconnectionstatechange", function(e) {
        console.log(e)
        console.log(peer.iceConnectionState)
        if (peer.iceConnectionState === "connected") {
            singalling.send("done")
            if(index >= peers.length){
                console.log("closed")
                singalling.close()
            }
            else {
                index++
            }
        }
    });
    return peer
}

/*
ready state
    create a new peer
    send first offerd 

sdp
    use last peer or make new one if null
    setup counter offer
    append to list
    set peer to null

establish
    peers[0] send ice offer
    index = 0
ice
    peers[index] set ice offer
    index++
    peers[index] send ice offer



here is the flow of the app

client processes
1. create peer 
    send offer
2. recive offer
    create peer
    send anwser
3. recive answer
    add to peer
    create peer 
    send offer
4.  send ice    (via event)
5.  recive ice  (via server)
    set ice
6. send done move on (via event)

server has 3 state
    setup two peers     ready
        send sdp back and fourth offer,answer
        send ice back and fourth ice
    peers done
    move to next peer in list or what ever
*/


