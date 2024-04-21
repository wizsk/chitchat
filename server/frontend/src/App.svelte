<script>
  import { afterUpdate, tick } from "svelte";
  /** * @type {any} */
  let array_any = []; // just to get rid of the warning :)

  /** * @type {user} */
  let user;
  /** * @type {[inbox]}  */
  let all_inboxes;

  // /** * @type {[inbox]}  */
  // let curr_all_inboxes;

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

        if (!all_inboxes) {
          all_inboxes = array_any;
        }
        break;
      case "message_send":
        if (data.error) {
          return;
        }

        for (let i = 0; i < all_inboxes.length; i++) {
          if (all_inboxes[i].user.id === data.message.receiver_id) {
            const n = all_inboxes.splice(i, 1);

            if (n[0].messages) {
              n[0].messages.push(data.message);
            } else {
              n[0].messages = [data.message];
            }

            if (n[0].user.id === inbox_selected.user.id) {
              inbox_selected = n[0];
              msg_area_scroll_to_bottom();
            }
            all_inboxes.unshift(n[0]);
            all_inboxes = all_inboxes;
            return;
          }
        }

        if (all_inboxes) {
          all_inboxes.unshift({ user: data.user, messages: [data.message] });
          all_inboxes = all_inboxes;
          inbox_selected = all_inboxes[0];
        } else {
          all_inboxes = [{ user: data.user, messages: [data.message] }];
          inbox_selected = all_inboxes[0];
        }
        break;

      case "message_receive":
        for (let i = 0; i < all_inboxes.length; i++) {
          if (all_inboxes[i].user.id === data.message.sender_id) {
            const n = all_inboxes.splice(i, 1);
            if (n[0].messages) {
              n[0].messages.push(data.message);
            } else {
              n[0].messages = [data.message];
            }

            if (n[0].user.id === inbox_selected.user.id) {
              inbox_selected = n[0];
              msg_area_scroll_to_bottom();
            }
            all_inboxes.unshift(n[0]);
            all_inboxes = all_inboxes;
            return;
          }
        }
        all_inboxes.unshift({ user: data.user, messages: [data.message] });
        all_inboxes = all_inboxes;
        break;
      case "search_user":
        tmp_all_inboxes = all_inboxes;
        searched_user = true;

        if (data.error) {
          all_inboxes = undefined;
        } else {
          all_inboxes = [{ user: data.user, messages: undefined }];
        }
        break;
      default:
        console.log("received unknown data:", data);
    }
  };

  let tmp_all_inboxes = undefined;
  let searched_user = false;

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
    search_input = search_input.trim();
    if (search_input === "") {
      all_inboxes = tmp_all_inboxes;
      tmp_all_inboxes = undefined;
      searched_user = false;
      return;
    }

    clearInterval(search_set_timeout_id);
    search_set_timeout_id = setTimeout(() => {
      for (let i = 0; i < all_inboxes.length; i++) {
        if (all_inboxes[i].user.user_name === search_input) {
          tmp_all_inboxes = all_inboxes;
          searched_user = true;
          all_inboxes = [all_inboxes[i]];
          return;
        }
      }

      const data = { data_type: "search_user", search_term: search_input };
      try {
        socket.send(JSON.stringify(data));
      } catch (err) {
        console.error(err);
      }
    }, 250);
  }
</script>

<main class="mx-4">
  <div class="font-bold text-xl">Helloo {user?.name}</div>
  <!-- <p>Helloo @{user?.user_name}</p>
  <p>id @{user?.id}</p> -->

  <div class="flex flex-row">
    <div class="w-[20vw] h-[90vh]">
      <div>
        <input
          type="text"
          class="text-xl w-full p-2 border border-red-500"
          placeholder="find user by username"
          bind:value={search_input}
          on:input={send_search}
        />
      </div>

      <div bind:this={chat_list} class="h-full overflow-y-scroll">
        {#if all_inboxes}
          {#each all_inboxes as inbox, i}
            <button
              on:click={() => {
                inbox_selected = inbox;
                if (searched_user) {
                  searched_user = false;
                  all_inboxes = tmp_all_inboxes;
                }
              }}
              class="h-20 w-full text-left flex flex-col"
            >
              <span class="font-bold text-xl">@{inbox.user.user_name}</span>
              <span class="opacity-80"
                >{inbox.messages
                  ? inbox.messages[inbox.messages.length - 1].message_text
                  : ""}</span
              >
            </button>
          {/each}
        {:else}
          <div class="w-full h-full flex justify-center items-center">
            No user found
          </div>
        {/if}
      </div>
    </div>

    <div class="bg-blue-100 flex-1 h-[90vh]">
      {#if inbox_selected}
        <div bind:this={message_area} class="h-full overflow-y-auto">
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
            <div class="w-full h-full flex justify-center items-center">
              send new message
            </div>
          {/if}
        </div>
        <form
          class="w-full flex"
          on:submit|preventDefault={() => {
            if (!message_input_text) return;
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
          <input
            class="w-full flex-1"
            type="text"
            placeholder="send a message"
            bind:value={message_input_text}
          />
          <button>Send</button>
        </form>
      {:else}
        <div class="w-full h-full flex justify-center items-center">
          select an inbox
        </div>
      {/if}
    </div>
  </div>
</main>
