
const socket = new WebSocket(`ws://${window.location.host}/ws`);

socket.onopen = function (event) {
    console.log("WebSocket connected");
};

let user;
let inbox;
let recived_first_msg = false;
socket.onmessage = function (event) {
    // console.log("Received message:", event.data);
    document.getElementById("output").value += event.data + "\n";
    let data;
    try {
        data = JSON.parse(event.data);
    } catch (err) {
        console.error("while parsing:", err);
        return;
    }

    switch (data.data_type) {
        case "user":
            console.log("user data:", data);
            user = data;
            break;
        case "get_inbox":
            console.log("get inbox", data);
            // inbox = constructInbox(data);
            inbox = data;
            break;
        default:
            console.log("unknown data", data);
    }

    // if (!recived_first_msg) {
    //     recived_first_msg = true;
    //     if (data.data_type !== "user") {
    //         console.error("very bad, the 1st msg is: ", data);
    //         return;
    //     }
    //     user = data.user;
    //     // request for inbox
    //     const d = { "data_type": "get_inbox" };
    //     socket.send(JSON.stringify(d));
    // }
};

socket.onclose = function (event) {
    console.log("WebSocket disconnected");
};

function sendMessage() {
    const message = document.getElementById("input").value;
    socket.send(message);
    console.log("Sent message:", message);
    document.getElementById("input").value = "";
}

function sendDM() {
    const to = Number(document.getElementById("receiver_id").value.trim());
    if (isNaN(to)) { alert("id dee vai"); return; }

    const msg = document.getElementById("msg").value.trim();
    if (msg === "") { alert("msg texxt dee vai"); return; }

    document.getElementById("input").value = "";
    const d = { "data_type": "message_send", "uuid": generateUUID(), "message": { "receiver_id": to, "message_text": msg } };
    console.log(d);
    socket.send(JSON.stringify(d));
}

function generateUUID() {
    let dt = new Date().getTime();
    const uuid = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
        const r = (dt + Math.random() * 16) % 16 | 0;
        dt = Math.floor(dt / 16);
        return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
    });
    return uuid;
}

function constructInbox(data) {
    let d = {};

    for (let i = 0; i < data.inbox.length; i++) {
        c = data.inbox[i];
        let k = c.receiver_id != user.id ? c.receiver_id : c.sender_id;

        if (d[k] === undefined) {
            d[k] = [c]
        } else {
            d[k].push(c);
        }
    }
    return d;
}
