const inbox_data_id_name = "data-inbox-user-id";
const socket = new WebSocket(`ws://${window.location.host}/ws`);

socket.onopen = function (event) {
    console.log("WebSocket connected");
};

let user;
let inboxes;
let selected_inbox_user;
let message_send = {};

let recived_first_msg = false;
socket.onmessage = function (event) {
    let data;
    try {
        data = JSON.parse(event.data);
    } catch (err) {
        console.error("while parsing:", err);
        return;
    }
    console.log(data)

    switch (data.data_type) {
        case "user":
            console.log("user data:", data);
            const me = document.getElementById("me");
            user = data.user;
            me.innerHTML = `Welcome back: ${user.name}<br>username: @${user.user_name}`;
            break;
        case "get_inbox":
            console.log("get inbox", data);
            // inbox = constructInbox(data);
            inboxes = data.all_inbox;

            const friends = document.getElementById("friends");
            friends.innerHTML = inboxes.reverse().map((v) => `<div ${inbox_data_id_name}="${v.user.id}">${v.user.name}</div>`).join("");

            document.querySelectorAll(`[${inbox_data_id_name}]`).forEach((elm) => {
                elm.addEventListener("click", (e) => {
                    const inbox = document.getElementById("inbox");
                    const id = Number(e.target.getAttribute(inbox_data_id_name));
                    for (let i = 0; i < inboxes.length; i++) {
                        if (inboxes[i].user.id === id) {
                            ibx = inboxes[i];
                            inbox.innerHTML = ibx.messages.map(v => {
                                return `${v.sender_id === user.id ? user.user_name : ibx.user.user_name}: ${v.message_text}`
                            }).join("<br>");
                            selected_inbox_user = ibx.user;
                            break;
                        }
                    }
                })
            })
            break;
        case "message_send":
            {
                const rid = data.message.receiver_id;
                for (let i = 0; i < inboxes.length; i++) {
                    if (inboxes[i].user.id === rid) {
                        inboxes[i].messages.push(data.message);
                        break;
                    }
                }
                if (selected_inbox_user.id !== rid) return;
                const inbox = document.getElementById("inbox");
                inbox.innerHTML += `<br>${user.user_name}: ${data.message.message_text}`;
            }
            break;
        case "message_receive":
            console.log(data)
            {
                const rid = data.message.sender_id;
                for (let i = 0; i < inboxes.length; i++) {
                    if (inboxes[i].user.id === rid) {
                        inboxes[i].messages.push(data.message);
                        break;
                    }
                }

                if (selected_inbox_user.id !== rid) return;
                const inbox = document.getElementById("inbox");
                inbox.innerHTML += `<br>${data.user.user_name}: ${data.message.message_text}`;
            }
            break;
        default:
            console.log("unknown data", data);
    }

    let x = [1, 2, 3, 4];
    x.map((v) => {

    })

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
    const msg = document.getElementById("msg").value.trim();
    if (msg === "") { alert("msg texxt dee vai"); return; }

    if (selected_inbox_user == undefined) { alert("select an inbox"); return; }

    document.getElementById("input").value = "";
    const uuid = generateUUID();
    const d = { "data_type": "message_send", "uuid": uuid, "message": { "receiver_id": selected_inbox_user.id, "message_text": msg } };
    message_send[uuid] = { message_text: msg, receiver_id: selected_inbox_user.id };
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
