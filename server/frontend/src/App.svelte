<script>
  import { afterUpdate, tick } from "svelte";

  /** * @type {user} */
  let user;
  /** * @type {[inbox]}  */
  let all_inboxes;

  /** * @type {[inbox]}  */
  let curr_all_inboxes;

  /** * @type {inbox}  */
  let inbox_selected = undefined;

  /**
   * @type {HTMLElement} chatlist
   */
  let chat_list;

  /** * @type {string}  */
  let message_input_text = "";
  /** * @type {string}  */
  let search_input = "";

  let message_area;

  const socket = new WebSocket(`ws://${window.location.host}/ws`);

  socket.onopen = function () {
    console.log("WebSocket connected");
  };

  socket.onmessage = async function (e) {
    /** @type {data} */
    let data;

    try {
      data = JSON.parse(e.data);
      console.log("WebSocket received msg:", data);
    } catch (err) {
      console.error("while parsing json form the websocket:", err);
      return;
    }

    switch (data.data_type) {
      case "user":
        user = data.user;
        break;
      case "get_inbox":
        all_inboxes = data.all_inboxes;
        curr_all_inboxes = all_inboxes;
        break;
      case "message_send":
        for (let i = 0; i < all_inboxes.length; i++) {
          if (all_inboxes[i].user.id === data.message.receiver_id) {
            all_inboxes[i].messages.push(data.message);
            if (all_inboxes[i].user.id === inbox_selected.user.id)
              inbox_selected = all_inboxes[i];
            break;
          }
        }
        msg_area_scroll_to_bottom();
        break;
      case "message_receive":
        for (let i = 0; i < all_inboxes.length; i++) {
          if (all_inboxes[i].user.id === data.message.sender_id) {
            all_inboxes[i].messages.push(data.message);

            if (all_inboxes[i].user.id === inbox_selected.user.id)
              inbox_selected = all_inboxes[i];
            break;
          }
        }
        msg_area_scroll_to_bottom();
        break;
      case "search_user":
        if (data.error) {
          console.log("no user found");
          curr_all_inboxes = all_inboxes;
          return;
        }
        // chat_list.innerHTML = "";
        // const div = document.createElement("div");
        // div.innerText = `@${data.user.user_name}`;
        // div.classList.add("block", "h-20", "bg-slate-200", "font-bold");
        // chat_list.appendChild(div);
        curr_all_inboxes = [{ user: data.user, messages: undefined }];
        console.log(curr_all_inboxes);

        break;
      default:
        console.log("received unknown data:", data);
    }
  };

  afterUpdate(() => {
    msg_area_scroll_to_bottom();
  });

  function msg_area_scroll_to_bottom() {
    if (!message_area) return;
    message_area.scroll({ top: message_area.scrollHeight, behavior: "smooth" });
  }

  let search_set_timeout_id;
  function send_search() {
    // TODO: clear stuff
    if (search_input === "") return;

    clearInterval(search_set_timeout_id);
    search_set_timeout_id = setTimeout(() => {
      const data = { data_type: "search_user", search_term: search_input };
      try {
        console.log("sending serach req:", data);
        socket.send(JSON.stringify(data));
      } catch (err) {
        console.error(err);
      }
    }, 250);
  }
</script>

<main>
  <p>Helloo {user?.name}</p>
  <p>Helloo @{user?.user_name}</p>
  <p>id @{user?.id}</p>
  <div></div>

  <div class="flex flex-row">
    <div>
      <div>
        <input
          type="text"
          class=" bg-red-400"
          bind:value={search_input}
          on:input={send_search}
        />
      </div>

      <div bind:this={chat_list}>
        {#if curr_all_inboxes}
          {#each curr_all_inboxes as inbox, i}
            <button
              on:click={() => {
                inbox_selected = inbox;
                console.log(
                  i,
                  "name",
                  inbox.user.name,
                  "user_name",
                  inbox.user.user_name,
                  message_area,
                );
              }}
              class="block h-20 bg-slate-200 font-bold"
            >
              @{inbox.user.user_name}
            </button>
          {/each}
        {/if}
      </div>
    </div>

    <div class="bg-blue-400 flex-1">
      {#if inbox_selected}
        <div bind:this={message_area} class="h-[70vh] overflow-y-auto">
          {#if inbox_selected.messages}
            {#each inbox_selected.messages as m}
              <div
                class={inbox_selected.user.id !== m.sender_id
                  ? "text-right"
                  : "text-left"}
              >
                {inbox_selected.user.id === m.sender_id
                  ? inbox_selected.user.user_name
                  : user.user_name}: {m.message_text}
              </div>
            {/each}
          {:else}
            send new message
          {/if}
        </div>
        <form
          on:submit|preventDefault={() => {
            if (!message_input_text) return;
            console.log("send_to:", inbox_selected.user.id, message_input_text);
            const d = {
              data_type: "message_send",
              message: {
                receiver_id: inbox_selected.user.id,
                message_text: message_input_text,
              },
            };

            socket.send(JSON.stringify(d));

            message_input_text = "";
          }}
        >
          <input type="text" bind:value={message_input_text} />
          <button>Send</button>
        </form>
      {:else}
        <div>select an inbox</div>
      {/if}
    </div>
  </div>
</main>
