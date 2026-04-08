<!-- web/src/routes/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { gsap } from 'gsap';
  import { ScrollTrigger } from 'gsap/ScrollTrigger';
  import { revealOnScroll } from '$lib/animations/reveal';
  import { mouseParallax } from '$lib/stores/mouseParallax';

  gsap.registerPlugin(ScrollTrigger);

  let sceneInner: HTMLElement;
  let heroContent: HTMLElement;
  let heroHeading: HTMLHeadingElement;
  let heroSub: HTMLParagraphElement;
  let heroCta: HTMLAnchorElement;
  let teaserSection: HTMLElement;

  onMount(() => {
    // ── Hero entrance ──
    const tl = gsap.timeline({ defaults: { ease: 'power2.out' } });
    tl
      .from(heroHeading, { opacity: 0, y: 40, duration: 1.2, delay: 0.3 })
      .from(heroSub, { opacity: 0, y: 20, duration: 0.8 }, '-=0.6')
      .from(heroCta, { opacity: 0, y: 16, duration: 0.6 }, '-=0.4');

    // ── Mouse look-around ──
    let idleTween: gsap.core.Tween | null = null;
    let idleTimer: ReturnType<typeof setTimeout>;

    function startIdleBreathing() {
      idleTween = gsap.to(sceneInner, {
        rotateX: 1.5,
        rotateY: 1.5,
        duration: 4,
        ease: 'sine.inOut',
        yoyo: true,
        repeat: -1,
        overwrite: true
      });
    }

    function onMouseMove(e: MouseEvent) {
      clearTimeout(idleTimer);
      idleTween?.kill();
      idleTween = null;

      // Normalize to -1..1 from viewport center
      const nx = (e.clientX / window.innerWidth - 0.5) * 2;
      const ny = (e.clientY / window.innerHeight - 0.5) * 2;

      // Tilt the scene
      gsap.to(sceneInner, {
        rotateY: nx * 8,
        rotateX: -ny * 6,
        duration: 1.2,
        ease: 'power2.out',
        overwrite: 'auto'
      });

      // Counter-tilt hero content (30% — feels painted on sky)
      gsap.to(heroContent, {
        rotateY: -nx * 8 * 0.3,
        rotateX: ny * 6 * 0.3,
        duration: 1.2,
        ease: 'power2.out',
        overwrite: 'auto'
      });

      // Drive star layer parallax
      mouseParallax.set({ x: nx, y: ny });

      idleTimer = setTimeout(startIdleBreathing, 3000);
    }

    // Start idle after 3s of inactivity on load
    idleTimer = setTimeout(startIdleBreathing, 3000);
    window.addEventListener('mousemove', onMouseMove);

    // ── Scroll reveals ──
    if (teaserSection) {
      const items = teaserSection.querySelectorAll<HTMLElement>('.teaser-item');
      items.forEach((el, i) => revealOnScroll(el, i * 0.12));
    }

    return () => {
      window.removeEventListener('mousemove', onMouseMove);
      clearTimeout(idleTimer);
      idleTween?.kill();
      mouseParallax.set({ x: 0, y: 0 });
    };
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

<!-- Earth POV hero: night sky looking up -->
<div class="scene-perspective">
  <div class="scene-inner" bind:this={sceneInner}>

    <!-- Milky Way band overlay -->
    <div class="milky-way" aria-hidden="true"></div>

    <!-- Atmosphere horizon at bottom -->
    <div class="atmosphere-horizon" aria-hidden="true"></div>

    <!-- Hero content -->
    <section class="hero">
      <div class="hero-inner" bind:this={heroContent}>
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

      <!-- Scroll indicator -->
      <div class="scroll-indicator" aria-hidden="true">
        <span class="scroll-line"></span>
        <span class="scroll-label">SCROLL</span>
      </div>
    </section>

  </div>
</div>

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
  /* ── Scene wrapper ── */
  .scene-perspective {
    perspective: 800px;
    perspective-origin: 50% 50%;
    position: relative;
    width: 100%;
    min-height: 100vh;
    overflow: hidden;
  }

  .scene-inner {
    transform-style: preserve-3d;
    position: relative;
    width: 100%;
    min-height: 100vh;
    will-change: transform;
  }

  /* ── Milky Way band ── */
  .milky-way {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      135deg,
      transparent 20%,
      rgba(240, 237, 230, 0.008) 35%,
      rgba(240, 237, 230, 0.022) 42%,
      rgba(240, 237, 230, 0.032) 50%,
      rgba(240, 237, 230, 0.022) 58%,
      rgba(240, 237, 230, 0.008) 65%,
      transparent 80%
    );
    pointer-events: none;
    z-index: 1;
  }

  /* ── Atmosphere horizon ── */
  .atmosphere-horizon {
    position: absolute;
    bottom: 0;
    left: -10%;
    right: -10%;
    height: 14vh;
    background: radial-gradient(
      ellipse at 50% 100%,
      rgba(8, 22, 65, 0.7) 0%,
      rgba(15, 45, 110, 0.5) 20%,
      rgba(25, 65, 140, 0.3) 40%,
      rgba(55, 130, 190, 0.18) 60%,
      rgba(79, 195, 247, 0.08) 75%,
      transparent 90%
    );
    border-radius: 50% 50% 0 0 / 80% 80% 0 0;
    pointer-events: none;
    z-index: 1;
  }

  /* ── Hero ── */
  .hero {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    text-align: center;
    padding: 6rem 2rem 4rem;
    position: relative;
    z-index: 2;
  }

  .hero-inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2rem;
    transform-style: preserve-3d;
    will-change: transform;
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
    border: 1px solid rgba(240, 237, 230, 0.2);
    border-bottom-color: rgba(200, 146, 42, 0.35);
    color: rgba(240, 237, 230, 0.7);
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.25em;
    padding: 1rem 2.5rem;
    text-decoration: none;
    transition: border-color 0.2s, color 0.2s;
    animation: ctaPulse 2.8s ease-in-out infinite;
  }

  .hero-cta:hover {
    border-color: rgba(240, 237, 230, 0.6);
    border-bottom-color: rgba(200, 146, 42, 0.8);
    color: #F0EDE6;
  }

  @keyframes ctaPulse {
    0%, 100% { border-bottom-color: rgba(200, 146, 42, 0.15); }
    50%       { border-bottom-color: rgba(200, 146, 42, 0.6); }
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

  /* ── Teasers ── */
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
