<!-- web/src/lib/components/SizeSelector.svelte -->
<script lang="ts">
  export let sizes: string[] = ['S', 'M', 'L', 'XL'];
  export let selected: string = '';
  export let onChange: (size: string) => void = () => {};

  function select(size: string) {
    selected = size;
    onChange(size);
  }
</script>

<div class="size-selector" role="radiogroup" aria-label="Select size">
  {#each sizes as size}
    <button
      class="size-btn"
      class:selected={selected === size}
      on:click={() => select(size)}
      role="radio"
      aria-checked={selected === size}
      data-magnetic
    >
      {size}
    </button>
  {/each}
</div>

<style>
  .size-selector {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .size-btn {
    width: 3rem;
    height: 3rem;
    border: 1px solid rgba(240, 237, 230, 0.2);
    background: transparent;
    color: rgba(240, 237, 230, 0.6);
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    letter-spacing: 0.1em;
    cursor: none;
    transition: border-color 0.2s, color 0.2s, background 0.2s;
  }

  .size-btn:hover {
    border-color: rgba(240, 237, 230, 0.5);
    color: #F0EDE6;
  }

  .size-btn.selected {
    border-color: #F0EDE6;
    color: #F0EDE6;
    background: rgba(240, 237, 230, 0.06);
  }
</style>
