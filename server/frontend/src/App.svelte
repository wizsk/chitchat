<script>
  import { afterUpdate, tick } from "svelte";

  /** * @type {user} */
  let user;
  /** * @type {[inbox]}  */
  let all_inboxes;
  /** * @type {inbox}  */
  let inbox_selected = undefined;

  /** * @type {string}  */
  let message_input_text = "";

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
        await msg_area_scroll_to_bottom();
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
        await msg_area_scroll_to_bottom();
        break;
    }
  };

  afterUpdate(() => {
    msg_area_scroll_to_bottom();
  });

  async function msg_area_scroll_to_bottom() {
    // scrollContainer.scrollTop = scrollContainer.scrollHeight;
    console.log(message_area.scrollTop, message_area.scrollHeight);
  }
</script>

<main>
  <p>Helloo {user?.name}</p>
  <p>Helloo @{user?.user_name}</p>
  <p>id @{user?.id}</p>
  <div></div>

  <div class="flex flex-row">
    <div>
      {#if all_inboxes}
        {#each all_inboxes as inbox, i}
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

    <div class="bg-blue-400 flex-1">
      {#if inbox_selected}
        <div bind:this={message_area} class="h-[70vh] overflow-y-auto">
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
