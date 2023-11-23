/**
 * @typedef {Object} Msg
 * @property {string} type  - value found on sdp offers
 */

/**
 * @type{HTMLInputElement}
 */
const input = document.getElementById("input");
/**
 * @type{HTMLDivElement}
 */
const output = document.getElementById("output");
/**
 * @type{HTMLParagraphElement}
 */
const log = document.getElementById("status");

/**
 * @type {Array<RTCDataChannel>}
 */
const channels = [];
/**
 * @type {Array<RTCPeerConnection>}
 */
const peers = Array.from({ length: 4 }, createPeer);
//5 connections each peer is connected to another peer ecluding itself


/**
 * a counter to keep track of what peer we are on
 * @type {number}
 */
var index = 0;

//singalling server setup
const singalling = new WebSocket("ws://localhost:8080/ws");
singalling.addEventListener("message", async function (event) {
  //check for done
  let data = JSON.parse(event.data);

  switch (data.type) {
    case "ready":
      const offer = await peers[index].createOffer();
      await peers[index].setLocalDescription(offer);
      console.log(offer);
      singalling.send(JSON.stringify(peers[index].localDescription));
      break;
    case "offer":
      console.log("offer");
      await peers[index].setRemoteDescription(data);
      const awnser = await peers[index].createAnswer();
      await peers[index].setLocalDescription(awnser);
      singalling.send(JSON.stringify(peers[index].localDescription));
      break;
    case "answer":
      console.log("answer");
      console.log(data);
      await peers[index].setRemoteDescription(data);
      break;
    default:
      console.log("default");
      if (data.candidate) {
        console.log("ice");
        console.log(data);
        peers[index].addIceCandidate(data);
      }
      break;
  }
});

/**
 * @returns {RTCPeerConnection}
 */
function createPeer() {
  let peer = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
  });

  peer.addEventListener("icecandidate", function (event) {
    console.log("Got candidate:::::", event);
    if (event.candidate) {
      singalling.send(JSON.stringify(event.candidate));
    }
  });
  peer.addEventListener("iceconnectionstatechange", function (e) {
    console.log(e);
    console.log(peer.iceConnectionState);
    if (peer.iceConnectionState === "connected") {
      singalling.send("done");
      console.log(index);
      index++;
    }
  });

  //setting up local channel
  let local = peer.createDataChannel("chat");
  local.addEventListener("open", ()=>{ log.textContent = `opened local`; console.log("opened local") });
  local.addEventListener("close", ()=>{ log.textContent = `closed local`;console.log("closed local")});
  channels.push(local);
  //setting up remote channel
  peer.addEventListener("datachannel", function ({ channel }) {
    channel.addEventListener("message", function (e) {
        console.log("CHANNEL DATA")
        console.log(e)
        console.log(e.data)
      let el = document.createElement("p");
      let txt = document.createTextNode(e.data);
      el.appendChild(txt);
      output.appendChild(el);
    });
    
    channel.addEventListener("open", ()=>{
        log.textContent = `opened local`;console.log("opened channel")});
    channel.addEventListener("close", ()=>{
          log.textContent = `closed channel`;console.log("closed channel ")});
  });
  return peer;
}

document.getElementById("send").addEventListener("click", function () {
  for (let ch of channels) {
    if (ch.readyState == "open"){
        ch.send(input.value);
    }
  }
    input.value = "";
    input.focus();
});
