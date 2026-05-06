<!--
  SetPicker — direct-jump thumbnails between products in the same drop set.
  Replaces prev/next chevrons with a small thumbnail strip so buyers can
  hop straight to the piece they want without cycling through.
  Only renders when the current product belongs to a setId.
-->
<script lang="ts">
  import { goto } from '$app/navigation';
  import type { Product } from '$lib/types/shop';

  export let allProducts: Product[] = [];
  export let currentSlug: string;
  export let setId: 'warped-reality' | undefined = undefined;

  $: members = setId
    ? allProducts.filter((p) => p.setId === setId)
    : [];

  function thumbFor(p: Product): string {
    return p.variants?.[0]?.productImage ?? p.image_url;
  }

  function go(slug: string) {
    if (slug === currentSlug) return;
    goto(`/shop/${slug}`, { noScroll: true });
  }
</script>

{#if members.length > 1}
  <nav class="set-picker" aria-label="Set members">
    <span class="picker-label">SET DROP — JUMP TO</span>
    <div class="thumbs">
      {#each members as m}
        <button
          class="thumb"
          class:active={m.slug === currentSlug}
          on:click={() => go(m.slug)}
          aria-label={m.name}
          aria-current={m.slug === currentSlug ? 'page' : undefined}
        >
          <img src={thumbFor(m)} alt={m.name} loading="lazy" />
          <span class="thumb-name">{m.name.replace('Warped Reality ', '')}</span>
        </button>
      {/each}
    </div>
  </nav>
{/if}

<style>
  .set-picker {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.85rem;
    padding: 1.2rem 0 0;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
  }

  .picker-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.50rem;
    letter-spacing: 0.4em;
    color: rgba(200, 146, 42, 0.65);
    text-transform: uppercase;
  }

  .thumbs {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
    justify-content: center;
  }

  .thumb {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.45rem;
    background: rgba(0, 0, 0, 0.4);
    border: 1px solid rgba(240, 237, 230, 0.1);
    padding: 0.55rem 0.7rem 0.5rem;
    cursor: pointer;
    transition: border-color 0.2s ease, transform 0.2s ease, background 0.2s ease;
    width: clamp(82px, 11vw, 110px);
  }

  .thumb img {
    width: 60px;
    height: 60px;
    object-fit: contain;
    display: block;
    filter: drop-shadow(0 6px 18px rgba(0,0,0,0.55));
  }

  .thumb-name {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.16em;
    color: rgba(240, 237, 230, 0.6);
    text-transform: uppercase;
    text-align: center;
    line-height: 1.3;
  }

  .thumb:hover {
    border-color: rgba(240, 237, 230, 0.3);
    transform: translateY(-2px);
  }

  .thumb.active {
    border-color: rgba(79, 195, 247, 0.7);
    background: rgba(79, 195, 247, 0.08);
  }

  .thumb.active .thumb-name {
    color: #F0EDE6;
  }
</style>
