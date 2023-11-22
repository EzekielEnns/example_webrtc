/**
 * @typedef {Object} Msg
 * @property {string} type  - type of message/how to processes
 * @property {string} value - value for message
 * @property {boolean} offer - if true is a offer if false is a anwser
 */

/*
 * initalize connection to singalling server
 *     handle messages from server
 * */

const stream = await navigator.mediaDevices.getDisplayMedia()
//connect to singalling server
const singalling = new WebSocket("ws://localhost:8080/ws");
//TODO init peers

/**
 * @type {Array<RTCPeerConnection>}
 */
const peers = []    //TODO fill array

/**
 * a peer tobe added to peers
 * @type {number}
 */
var index;

singalling.addEventListener("message", async function(event){
    /**
     * @type {Msg}
     */
    const data = JSON.parse(event.data)

    switch (data.type){
        case "ready":
            const offer = await peers[index].createOffer()
            await peers[index].setLocalDescription(offer)
            singalling.send(offer.sdp)
            break;
        case "offer":
            await peers[index].setRemoteDescription(data.value)
            const awnser = await peers[index].createAnswer()
            singalling.send(awnser.sdp)
            break;
        case "awnser":
            await peers[index].setRemoteDescription(data.value)
            break;
        case "ice":
            peers[index].addIceCandidate(data.value)
            singalling.send("done")
            index++;
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
    peer.addTrack(stream.getTracks()[0], stream);
    peer.addEventListener("icecandidate", function(event) {
        singalling.send(JSON.stringify(event.candidate))
    });
    peer.addEventListener("iceconnectionstatechange", function() {
        if (peer.iceConnectionState === "connected") {
            singalling.send("done")
            index++
        }
    });
    peer.addEventListener("track", function(event) {
        //add track to videos
        // const video = document.getElementById("video") as HTMLVideoElement;
        // video.srcObject = event.streams[0];
    });
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


