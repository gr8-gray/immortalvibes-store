<!-- web/src/routes/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { gsap } from 'gsap';
  import { ScrollTrigger } from 'gsap/ScrollTrigger';
  import MissionScene from '$lib/components/MissionScene.svelte';
  import { revealOnScroll } from '$lib/animations/reveal';

  gsap.registerPlugin(ScrollTrigger);

  let heroHeading: HTMLHeadingElement;
  let heroSub: HTMLParagraphElement;
  let heroCta: HTMLAnchorElement;
  let teaserSection: HTMLElement;

  onMount(() => {
    // Hero entrance sequence — not scroll triggered, runs immediately
    const tl = gsap.timeline({ defaults: { ease: 'power2.out' } });
    tl
      .from(heroHeading, { opacity: 0, y: 40, duration: 1.2, delay: 0.3 })
      .from(heroSub, { opacity: 0, y: 20, duration: 0.8 }, '-=0.6')
      .from(heroCta, { opacity: 0, y: 16, duration: 0.6 }, '-=0.4');

    // Scroll reveals for mission teasers
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

<!-- Hero: full-bleed mission 001 scene as base -->
<MissionScene missionNumber="001">
  <section class="hero">
    <div class="hero-inner">
      <p class="hero-eyebrow">IMMORTAL VIBES</p>

      <h1 bind:this={heroHeading} class="hero-heading">
        RISE BEYOND<br />THE MORTAL PLANE
      </h1>

      <p bind:this={heroSub} class="hero-sub">
        Garments built for those who orbit higher.<br />
        Limited drops. Infinite purpose.
      </p>

      <a
        bind:this={heroCta}
        href="/shop"
        class="hero-cta"
        data-magnetic
      >
        SELECT YOUR MISSION
      </a>
    </div>

    <!-- Scroll indicator -->
    <div class="scroll-indicator" aria-hidden="true">
      <span class="scroll-line"></span>
      <span class="scroll-label">SCROLL</span>
    </div>
  </section>
</MissionScene>

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
  .hero {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    padding: 6rem 2rem 4rem;
    position: relative;
  }

  .hero-inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2rem;
  }

  .hero-eyebrow {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.4em;
    color: rgba(240, 237, 230, 0.35);
    margin: 0;
  }

  .hero-heading {
    font-family: 'Cormorant Garamond', serif;
    font-size: clamp(3rem, 10vw, 9rem);
    font-weight: 300;
    color: #F0EDE6;
    margin: 0;
    line-height: 0.95;
    letter-spacing: -0.01em;
  }

  .hero-sub {
    font-family: 'Inter', sans-serif;
    font-size: 0.875rem;
    line-height: 1.8;
    color: rgba(240, 237, 230, 0.45);
    margin: 0;
  }

  .hero-cta {
    display: inline-block;
    border: 1px solid rgba(240, 237, 230, 0.3);
    color: rgba(240, 237, 230, 0.7);
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.25em;
    padding: 1rem 2.5rem;
    text-decoration: none;
    transition: border-color 0.2s, color 0.2s;
  }

  .hero-cta:hover {
    border-color: rgba(240, 237, 230, 0.7);
    color: #F0EDE6;
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
    background: linear-gradient(to bottom, transparent, rgba(240,237,230,0.3));
  }

  .scroll-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.25em;
    color: rgba(240, 237, 230, 0.25);
  }

  @keyframes scrollBob {
    0%, 100% { transform: translateX(-50%) translateY(0); }
    50% { transform: translateX(-50%) translateY(6px); }
  }

  /* Teaser section */
  .teasers {
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

  .teaser-body {
    flex: 1;
  }

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

  .teaser-link:hover {
    color: #F0EDE6;
  }

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
