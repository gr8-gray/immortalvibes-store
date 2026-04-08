<!-- web/src/routes/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { gsap } from 'gsap';
  import { ScrollTrigger } from 'gsap/ScrollTrigger';
  import { revealOnScroll } from '$lib/animations/reveal';
  import HeroScene from '$lib/components/HeroScene.svelte';

  gsap.registerPlugin(ScrollTrigger);

  let heroHeading: HTMLHeadingElement;
  let heroSub: HTMLParagraphElement;
  let heroCta: HTMLAnchorElement;
  let teaserSection: HTMLElement;

  // CTA visibility — show when looking up, only hide when looking significantly down
  // Hysteresis prevents it vanishing when mouse drifts down to click
  let ctaShown = false;
  function handleCameraUpdate(camY: number) {
    if (!heroCta) return;
    const shouldShow = ctaShown ? camY > -0.18 : camY > 0.10;
    if (shouldShow === ctaShown) return;
    ctaShown = shouldShow;
    gsap.to(heroCta, {
      opacity: shouldShow ? 1 : 0,
      y: shouldShow ? 0 : 10,
      duration: 0.5,
      ease: 'power2.out',
      overwrite: 'auto'
    });
  }

  onMount(() => {
    // CTA starts hidden — revealed only when looking up
    gsap.set(heroCta, { opacity: 0, y: 10 });

    // Hero text entrance — delayed to let scene fade in first
    const tl = gsap.timeline({ defaults: { ease: 'power2.out' } });
    tl
      .from(heroHeading, { opacity: 0, y: 30, duration: 1.4, delay: 2.0 })
      .from(heroSub,     { opacity: 0, y: 20, duration: 0.9 }, '-=0.6');

    // Scroll reveals for teasers
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

<!-- Panoramic environment canvas -->
<HeroScene onCameraUpdate={handleCameraUpdate} />

<!-- Hero text overlay — centered above the scene -->
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
      ENTER THE MISSIONS
    </a>
  </div>

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
  /* ── Hero overlay — vertically centered in the sky ── */
  .hero {
    position: relative;
    z-index: 10;
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    padding: 2rem;
    pointer-events: none;
  }

  /* Re-enable pointer events only on interactive elements */
  .hero-cta {
    pointer-events: all;
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
    text-shadow: 0 0 80px rgba(0,0,0,0.8);
  }

  .hero-sub {
    font-family: 'Inter', sans-serif;
    font-size: 0.875rem;
    line-height: 1.8;
    color: rgba(240, 237, 230, 0.45);
    margin: 0;
    text-shadow: 0 0 20px rgba(0,0,0,0.9);
  }

  .hero-cta {
    display: inline-block;
    border: 1px solid rgba(240, 237, 230, 0.3);
    border-bottom-color: rgba(200, 146, 42, 0.5);
    color: rgba(240, 237, 230, 0.9);
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.25em;
    padding: 1rem 2.5rem;
    text-decoration: none;
    transition: border-color 0.2s, color 0.2s, box-shadow 0.2s;
    animation: ctaPulse 2.8s ease-in-out infinite;
    background: rgba(0,0,0,0.4);
    backdrop-filter: blur(8px);
    box-shadow: 0 0 30px rgba(240,237,230,0.08), 0 0 60px rgba(200,146,42,0.06);
  }

  .hero-cta:hover {
    border-color: rgba(240, 237, 230, 0.8);
    border-bottom-color: rgba(200, 146, 42, 1);
    color: #F0EDE6;
    box-shadow: 0 0 50px rgba(240,237,230,0.18), 0 0 100px rgba(200,146,42,0.14);
  }

  .hero-cta:hover {
    border-color: rgba(240, 237, 230, 0.6);
    border-bottom-color: rgba(200, 146, 42, 0.8);
    color: #F0EDE6;
  }

  @keyframes ctaPulse {
    0%, 100% { border-bottom-color: rgba(200, 146, 42, 0.15); }
    50%       { border-bottom-color: rgba(200, 146, 42, 0.65); }
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
    filter: drop-shadow(0 4px 12px rgba(0,0,0,0.5));
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
