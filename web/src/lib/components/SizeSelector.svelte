<!-- web/src/lib/components/SizeSelector.svelte -->
<script lang="ts">
  export let sizes: string[] = ['S', 'M', 'L', 'XL'];
  export let selected: string = '';
  export let onChange: (size: string) => void = () => {};
  /** Sizes with explicit stock=0 variant rows — renders disabled + struck-through.
   *  Sizes absent from this set stay enabled (no data ≠ sold out). */
  export let soldOut: Set<string> = new Set();

  function select(size: string) {
    if (soldOut.has(size)) return;
    selected = size;
    onChange(size);
  }
</script>

<div class="size-selector" role="radiogroup" aria-label="Select size">
  {#each sizes as size}
    {@const out = soldOut.has(size)}
    <button
      class="size-btn"
      class:selected={selected === size}
      class:sold-out={out}
      on:click={() => select(size)}
      role="radio"
      aria-checked={selected === size}
      aria-disabled={out}
      disabled={out}
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

  .size-btn:hover:not(:disabled) {
    border-color: rgba(240, 237, 230, 0.5);
    color: #F0EDE6;
  }

  .size-btn.selected {
    border-color: #F0EDE6;
    color: #F0EDE6;
    background: rgba(240, 237, 230, 0.06);
  }

  /* Sold-out: Lunar White at reduced opacity, struck-through — matches design language */
  .size-btn.sold-out {
    opacity: 0.3;
    text-decoration: line-through;
    cursor: not-allowed;
    border-color: rgba(240, 237, 230, 0.1);
  }
</style>
