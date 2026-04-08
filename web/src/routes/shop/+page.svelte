<!-- web/src/routes/shop/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import type { PageData } from './$types';
  import MissionCard from '$lib/components/MissionCard.svelte';
  import { revealOnScroll } from '$lib/animations/reveal';

  export let data: PageData;

  let heading: HTMLHeadingElement;
  let grid: HTMLDivElement;

  function missionLabel(num: string): string {
    const labels: Record<string, string> = {
      '001': 'Low Earth Orbit',
      '002': 'Lunar Surface',
      '003': 'Stellar Nursery',
      '004': 'Deep Space',
    };
    return labels[num] ?? 'Unknown';
  }

  onMount(() => {
    if (heading) revealOnScroll(heading, 0);
    if (grid) {
      const cards = grid.querySelectorAll<HTMLElement>('.reveal-card');
      cards.forEach((card, i) => revealOnScroll(card, i * 0.1));
    }
  });
</script>

<svelte:head>
  <title>Mission Select — Immortal Vibes</title>
</svelte:head>

<main class="shop-page">
  <!-- Background gradient -->
  <div class="shop-bg" aria-hidden="true"></div>

  <!-- Static star layer (no canvas here — lightweight) -->
  <svg class="shop-stars" aria-hidden="true" xmlns="http://www.w3.org/2000/svg">
    {#each Array(60) as _, i}
      <circle
        cx="{(i * 137.508) % 100}%"
        cy="{(i * 97.3) % 100}%"
        r="{0.4 + (i % 4) * 0.35}"
        fill="rgba(240,237,230,{0.15 + (i % 5) * 0.1})"
      />
    {/each}
  </svg>

  <div class="shop-inner">
    <header class="shop-header">
      <p class="shop-eyebrow">SELECT YOUR MISSION</p>
      <h1 bind:this={heading} class="shop-title">Choose Your Orbit</h1>
    </header>

    <div bind:this={grid} class="mission-grid">
      {#each data.products as product (product.id)}
        <div class="reveal-card">
          <MissionCard
            missionNumber={product.mission_number}
            title={product.name}
            subtitle="{product.mission_number} · {missionLabel(product.mission_number)}"
            slug={product.slug}
            imageUrl={product.image_url}
            status={product.status}
          />
        </div>
      {/each}

      <!-- Placeholder card for missions without products yet -->
      {#if data.products.length < 4}
        {#each Array(4 - data.products.length) as _, i}
          <div class="reveal-card">
            <div class="mission-placeholder">
              <p class="placeholder-label">CLASSIFIED</p>
              <p class="placeholder-sub">Next drop: TBD</p>
            </div>
          </div>
        {/each}
      {/if}
    </div>
  </div>
</main>

<style>
  .shop-page {
    position: relative;
    min-height: 100vh;
    background: #030308;
    overflow: hidden;
  }

  .shop-bg {
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse 80% 50% at 50% 0%, rgba(8,8,15,0.9), #030308);
  }

  .shop-stars {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }

  .shop-inner {
    position: relative;
    z-index: 1;
    max-width: 1200px;
    margin: 0 auto;
    padding: 8rem 2rem 4rem;
  }

  .shop-header {
    text-align: center;
    margin-bottom: 5rem;
  }

  .shop-eyebrow {
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.3em;
    color: rgba(240, 237, 230, 0.4);
    margin: 0 0 1.5rem;
  }

  .shop-title {
    font-family: 'Cormorant Garamond', serif;
    font-size: clamp(2.5rem, 6vw, 5rem);
    font-weight: 300;
    color: #F0EDE6;
    margin: 0;
    line-height: 1;
  }

  .mission-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
  }

  @media (max-width: 640px) {
    .mission-grid {
      grid-template-columns: 1fr;
    }
  }

  .mission-placeholder {
    aspect-ratio: 4 / 5;
    border: 1px solid rgba(240, 237, 230, 0.05);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
  }

  .placeholder-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.3em;
    color: rgba(240, 237, 230, 0.2);
    margin: 0;
  }

  .placeholder-sub {
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    color: rgba(240, 237, 230, 0.15);
    margin: 0;
  }

  .reveal-card {
    /* initial visibility handled by revealOnScroll — opacity 0 set by GSAP */
  }
</style>
