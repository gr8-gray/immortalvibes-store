<!-- web/src/routes/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { gsap } from 'gsap';
  import { ScrollTrigger } from 'gsap/ScrollTrigger';
  import { revealOnScroll } from '$lib/animations/reveal';
  import HeroScene from '$lib/components/HeroScene.svelte';

  gsap.registerPlugin(ScrollTrigger);

  let teaserSection: HTMLElement;

  onMount(() => {
    if (teaserSection) {
      const items = teaserSection.querySelectorAll<HTMLElement>('.teaser-item');
      items.forEach((el, i) => revealOnScroll(el, i * 0.12));
    }
  });

  const missions = [
    { number: '001', label: 'Low Earth Orbit', product: 'Warped Reality Beanie', slug: 'warped-reality-beanie', image: '/photos/blue-beanie.jpeg' },
    { number: '002', label: 'Lunar Surface',   product: 'Vanguard Trucker Hat',  slug: 'vanguard-trucker-hat', image: null },
    { number: '003', label: 'Stellar Nursery', product: 'Racerback Tanktop',     slug: 'racerback-tanktop',    image: '/photos/tank-front.png' },
    { number: '004', label: 'Deep Space',      product: 'Next Drop',             slug: null,                   image: null },
  ] as const;
</script>

<svelte:head>
  <title>Immortal Vibes — Rise Beyond the Mortal Plane</title>
  <meta name="description" content="Garments built for those who orbit higher. Limited drops, infinite purpose." />
</svelte:head>

<HeroScene />

<!-- Scroll indicator — anchored to bottom of viewport during hero -->
<section class="hero-spacer">
  <div class="scroll-indicator" aria-hidden="true">
    <span class="scroll-line"></span>
    <span class="scroll-label">SCROLL</span>
  </div>
</section>

<!-- Mission teasers below the fold -->
<section bind:this={teaserSection} class="teasers">
  <div class="teasers-inner">
    <p class="teasers-eyebrow">ACTIVE MISSIONS</p>

    {#each missions as mission}
      <div class="teaser-item">
        <span class="teaser-number">{mission.number}</span>
        <div class="teaser-body">
          <p class="teaser-label">{mission.label}</p>
          <p class="teaser-product">{mission.product}</p>
        </div>
        {#if mission.image}
          <img class="teaser-image" src={mission.image} alt={mission.product} />
        {/if}
        {#if mission.slug}
          <a href="/shop/{mission.slug}" class="teaser-link" data-magnetic>
            VIEW →
          </a>
        {:else}
          <span class="teaser-tbd">TBD</span>
        {/if}
      </div>
    {/each}
  </div>
</section>

<style>
  .hero-spacer {
    position: relative;
    z-index: 10;
    min-height: 100vh;
    pointer-events: none;
  }

  .scroll-indicator {
    position: absolute;
    bottom: 2.5rem;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    animation: scrollBob 2.5s ease-in-out infinite;
  }

  .scroll-line {
    display: block;
    width: 1px;
    height: 40px;
    background: linear-gradient(to bottom, transparent, rgba(240, 237, 230, 0.3));
  }

  .scroll-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.25em;
    color: rgba(240, 237, 230, 0.25);
  }

  @keyframes scrollBob {
    0%, 100% { transform: translateX(-50%) translateY(0); }
    50%       { transform: translateX(-50%) translateY(6px); }
  }

  /* ── Teasers ── */
  .teasers {
    position: relative;
    z-index: 10;
    background: #030308;
    padding: 8rem 2rem;
  }

  .teasers-inner {
    max-width: 800px;
    margin: 0 auto;
  }

  .teasers-eyebrow {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.3em;
    color: rgba(240, 237, 230, 0.3);
    margin: 0 0 3rem;
  }

  .teaser-item {
    display: flex;
    align-items: center;
    gap: 2rem;
    padding: 1.75rem 0;
    border-bottom: 1px solid rgba(240, 237, 230, 0.06);
  }

  .teaser-number {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.2);
    width: 2.5rem;
    flex-shrink: 0;
  }

  .teaser-body { flex: 1; }

  .teaser-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.15em;
    color: rgba(240, 237, 230, 0.3);
    margin: 0 0 0.25rem;
    text-transform: uppercase;
  }

  .teaser-product {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.4rem;
    font-weight: 300;
    color: #F0EDE6;
    margin: 0;
  }

  .teaser-link {
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.15em;
    color: rgba(240, 237, 230, 0.4);
    text-decoration: none;
    transition: color 0.2s;
    flex-shrink: 0;
  }

  .teaser-link:hover { color: #F0EDE6; }

  .teaser-image {
    width: 56px;
    height: 56px;
    object-fit: contain;
    filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.5));
    transform: rotate(-6deg);
    flex-shrink: 0;
    opacity: 0.85;
    transition: opacity 0.2s, transform 0.3s;
  }

  .teaser-item:hover .teaser-image {
    opacity: 1;
    transform: rotate(-3deg) scale(1.06);
  }

  .teaser-tbd {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.15);
    flex-shrink: 0;
  }
</style>
