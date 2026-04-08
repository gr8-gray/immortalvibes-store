<!-- web/src/lib/components/MissionCard.svelte -->
<script lang="ts">
  import { goto } from '$app/navigation';

  export let missionNumber: '001' | '002' | '003' | '004';
  export let title: string;
  export let subtitle: string; // e.g. "001 · Low Earth Orbit"
  export let slug: string;
  export let imageUrl: string = '';
  export let status: 'available' | 'sold_out' | 'coming_soon' = 'available';

  const gradients: Record<string, string> = {
    '001': 'linear-gradient(135deg, #060a14 0%, #0a1020 60%, rgba(30,111,217,0.15) 100%)',
    '002': 'linear-gradient(135deg, #030308 0%, #1a1814 100%)',
    '003': 'linear-gradient(135deg, #030308 0%, #3d1200 100%)',
    '004': 'linear-gradient(135deg, #030308 0%, #2a0d00 100%)',
  };

  let card: HTMLDivElement;
  let rx = 0;
  let ry = 0;
  const MAX_TILT = 12;

  function handleMouseMove(e: MouseEvent) {
    const rect = card.getBoundingClientRect();
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    const dx = e.clientX - cx;
    const dy = e.clientY - cy;
    ry = (dx / (rect.width / 2)) * MAX_TILT;
    rx = -(dy / (rect.height / 2)) * MAX_TILT;
  }

  function handleMouseLeave() {
    rx = 0;
    ry = 0;
  }

  function handleClick() {
    goto(`/shop/${slug}`);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') handleClick();
  }
</script>

<div
  bind:this={card}
  class="mission-card"
  style="background: {gradients[missionNumber]}; transform: perspective(800px) rotateX({rx}deg) rotateY({ry}deg);"
  on:mousemove={handleMouseMove}
  on:mouseleave={handleMouseLeave}
  on:click={handleClick}
  on:keydown={handleKeydown}
  role="button"
  tabindex="0"
  data-magnetic
  aria-label="View {title} product page"
>
  <!-- Mini star layer — static SVG dots for card preview -->
  <svg class="card-stars" aria-hidden="true" xmlns="http://www.w3.org/2000/svg">
    {#each Array(30) as _, i}
      <circle
        cx="{(i * 137.5) % 100}%"
        cy="{(i * 97.3) % 100}%"
        r="{0.5 + (i % 3) * 0.4}"
        fill="rgba(240,237,230,{0.3 + (i % 4) * 0.15})"
      />
    {/each}
  </svg>

  <div class="card-content">
    <p class="mission-label">{subtitle}</p>
    <h2 class="product-title">{title}</h2>

    {#if status === 'sold_out'}
      <span class="badge badge--sold-out">SOLD OUT</span>
    {:else if status === 'coming_soon'}
      <span class="badge badge--coming-soon">COMING SOON</span>
    {:else}
      <span class="card-cta">EXPLORE MISSION →</span>
    {/if}

    {#if imageUrl}
      <img class="card-product-image" src={imageUrl} alt={title} />
    {/if}
  </div>
</div>

<style>
  .mission-card {
    position: relative;
    overflow: hidden;
    border: 1px solid rgba(240, 237, 230, 0.08);
    border-radius: 4px;
    cursor: pointer;
    aspect-ratio: 4 / 5;
    transition: transform 0.3s ease, box-shadow 0.3s ease;
    will-change: transform;
  }

  .mission-card:hover {
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.6);
  }

  .card-stars {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }

  .card-content {
    position: relative;
    z-index: 1;
    padding: 2rem;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
  }

  .mission-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.2em;
    text-transform: uppercase;
    color: rgba(240, 237, 230, 0.5);
    margin: 0 0 0.5rem;
  }

  .product-title {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.8rem;
    font-weight: 300;
    color: #F0EDE6;
    margin: 0 0 1rem;
    line-height: 1.15;
  }

  .card-cta {
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    letter-spacing: 0.15em;
    color: rgba(240, 237, 230, 0.6);
  }

  .badge {
    display: inline-block;
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.2em;
    padding: 0.25rem 0.75rem;
    border-radius: 2px;
  }

  .badge--sold-out {
    border: 1px solid rgba(240, 237, 230, 0.3);
    color: rgba(240, 237, 230, 0.5);
  }

  .badge--coming-soon {
    border: 1px solid rgba(240, 237, 230, 0.2);
    color: rgba(240, 237, 230, 0.4);
  }

  .card-product-image {
    position: absolute;
    top: 1.5rem;
    right: 1rem;
    width: 45%;
    object-fit: contain;
    filter: drop-shadow(0 8px 24px rgba(0, 0, 0, 0.5));
    transform: rotate(-8deg);
    transition: transform 0.4s ease;
  }

  .mission-card:hover .card-product-image {
    transform: rotate(-4deg) scale(1.04) translateY(-4px);
  }
</style>
