<script>
  let count = 0;
  const increment = () => {
    count += 1;
  };
</script>

<button class=" bg-red-400" on:click={increment}>
  count is {count}
</button>
