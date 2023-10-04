//TODO support multiple peers

document.getElementById("conn")?.addEventListener("click", async function() {
  document.getElementById("conn")?.setAttribute("disabled", "true");

  const peer = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302", },],
  });
  //note a stream is required to start ICE
  const stream = await navigator.mediaDevices.getDisplayMedia()
  peer.addTrack(stream.getTracks()[0], stream);

  //signaling server on second connect instant trigger "knock knock"
  const server = new WebSocket("ws://localhost:8080/ws");

  //add stream to peer
  peer.addEventListener("track", function(event) {
    //console.log("track", event);
    const video = document.getElementById("video") as HTMLVideoElement;
    video.srcObject = event.streams[0];
  });

  peer.addEventListener("icecandidate", function(event) {
    //console.log("ice candidate", event);
    if (event.candidate) {
      server.send(JSON.stringify({ type: "tomWho", value: event.candidate }));
    }
  });
  peer.addEventListener("iceconnectionstatechange", function() {
    if (peer.iceConnectionState === "connected") {
      server.close();
    }
  });


  server.addEventListener("message", async function(event) {
    const data = JSON.parse(event.data);
    //console.log("message", data);
    switch (data.type) {
      case "knockKnock":
        //console.log("knock knock");
        let offer = await peer.createOffer();
        await peer.setLocalDescription(offer);
        //send offer
        server.send(JSON.stringify({ type: "whosThere", value: offer }));
        break;
      case "whosThere":
        //console.log("whos there");
        //answer offer
        await peer.setRemoteDescription(data.value);
        const answer = await peer.createAnswer();
        await peer.setLocalDescription(answer);
        server.send(JSON.stringify({ type: "tom", value: answer }));
        break;
      case "tom":
        //console.log("major tom");
        await peer.setRemoteDescription(data.value);
        break;
      case "tomWho":
        //console.log("tom who");
        await peer.addIceCandidate(data.value);
        break;
      default:
        //console.log("unknown message type", data.type);
        break;
    }
  });


})
