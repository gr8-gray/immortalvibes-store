<!-- web/src/routes/shop/+page.svelte -->
<!-- WR set drop -->

<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import MissionPlanet, { type ClusterItem, type CarouselSlot } from '$lib/components/MissionPlanet.svelte';
  import { playSystemReformation, REFORMATION_DURATION_MS } from '$lib/animations/system-reformation.ts';
  import type { gsap } from 'gsap';

  // Intro toggle — flip via PUBLIC_DROP_INTRO_ENABLED env var.
  // String compare because Vite injects env values as strings.
  const introEnabled = (import.meta.env.PUBLIC_DROP_INTRO_ENABLED ?? 'true') !== 'false';

  // ── Permanent missions (post-intro state) ────────────────────────────────
  // Trucker is decommissioned — only Warped Reality Set + Tank remain.
  // WR planet shows a self-winding carousel of the blue-variant pieces.
  // One shirt only — Spires (more designed). Both tee variants live on the set landing page.
  const setCarousel: CarouselSlot[] = [
    // Blue beanie PNG sits high on its canvas — nudge down so the beanie is
    // centered on the planet sphere instead of drifting toward the top.
    { url: '/photos/_drop/beanie-blue.png', offsetY: -0.18 },
    '/photos/_drop/tee-front.png',
    '/photos/_drop/sweatpants-front.png',
  ];

  const setMission = {
    num: '001',
    name: 'Warped Reality',
    env: 'The Set Drop',
    slug: 'warped-reality-set',
    planetType: 'leo' as const,
    carousel: setCarousel,
    glow: '#4FC3F7',
    speed: 0.0018,
    tilt: 0.22,
  };

  const tankMission = {
    num: '003',
    name: 'Racerback Tanktop',
    env: 'Stellar Nursery',
    slug: 'racerback-tanktop',
    planetType: 'nebula' as const,
    product: '/photos/product-tank.png',
    productScale: 1.1,
    spriteBlending: 'additive' as const,
    glow: '#6B0FCC',
    speed: 0.0022,
    tilt: 0.32,
  };

  // ── Intro state ──────────────────────────────────────────────────────────
  let showIntroTrucker = introEnabled;
  let introPlayed = false;
  let timeline: ReturnType<typeof gsap.timeline> | null = null;

  // Bound refs for animation
  let truckerSlotEl:   HTMLElement;
  let truckerHaloEl:   HTMLElement;
  let truckerPlanetEl: HTMLElement;
  let wrSlotEl:        HTMLElement;
  let wrHaloEl:        HTMLElement;
  let wrClusterEl:     HTMLElement;
  let tankSlotEl:      HTMLElement;
  let collapseRing1El: HTMLElement;
  let collapseRing2El: HTMLElement;
  let collapseRing3El: HTMLElement;
  let singularityEl:   HTMLElement;
  let shockwaveEl:     HTMLElement;
  let shockwave2El:    HTMLElement;
  let shockwave3El:    HTMLElement;
  let flashEl:         HTMLElement;
  let streaksEl:       HTMLElement;
  let particlesEl:     HTMLElement;

  let truckerCollapsing = false;

  function startTruckerCollapse() {
    truckerCollapsing = true;
  }

  function skipIntro() {
    if (timeline) {
      timeline.progress(1).kill();
      timeline = null;
    }
    truckerCollapsing = true;
    // Wait for the reflow transition before unmounting the trucker slot.
    setTimeout(() => {
      showIntroTrucker = false;
      introPlayed = true;
    }, 1100);
  }

  onMount(() => {
    if (!browser || !introEnabled || introPlayed) return;

    const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    // Defer one frame so refs are bound after Svelte commit.
    requestAnimationFrame(() => {
      timeline = playSystemReformation({
        truckerSlot:    truckerSlotEl,
        truckerHalo:    truckerHaloEl,
        truckerPlanet:  truckerPlanetEl,
        wrSlot:         wrSlotEl,
        wrHalo:         wrHaloEl,
        wrCluster:      wrClusterEl,
        tankSlot:       tankSlotEl,
        collapseRing1:  collapseRing1El,
        collapseRing2:  collapseRing2El,
        collapseRing3:  collapseRing3El,
        singularity:    singularityEl,
        shockwave:      shockwaveEl,
        shockwave2:     shockwave2El,
        shockwave3:     shockwave3El,
        flash:          flashEl,
        streaks:        streaksEl,
        particles:      particlesEl,
      }, {
        reduceMotion,
        onTruckerGone: startTruckerCollapse,
        onComplete: () => {
          // Reflow already started ~1.95s ago via onTruckerGone — by now
          // the row has fully collapsed. Safe to unmount the slot.
          showIntroTrucker = false;
          introPlayed = true;
        },
      });
    });
  });

  onDestroy(() => {
    if (timeline) timeline.kill();
  });
</script>

<svelte:head>
  <title>Immortal Vibes — Select Your Mission</title>
</svelte:head>

<div class="shop">

  <header class="top-bar">
    <span class="select-label">MISSION SELECT</span>
  </header>

  <div class="planets-row">

    <!-- ── Warped Reality SET (permanent slot 1) ── -->
    <a href="/shop/{setMission.slug}" class="mission-slot wr-slot" bind:this={wrSlotEl}>
      <div class="planet-wrap">
        <MissionPlanet
          planetType={setMission.planetType}
          carousel={setMission.carousel}
          carouselInterval={3500}
          carouselFade={850}
          productScale={0.72}
          productOpacity={0.95}
          glowColor={setMission.glow}
          rotationSpeed={setMission.speed}
          axialTilt={setMission.tilt}
        />
        <div class="planet-halo" style="--glow:{setMission.glow}" bind:this={wrHaloEl}></div>
        <div class="set-cluster-marker" bind:this={wrClusterEl} aria-hidden="true"></div>
      </div>

      <div class="mission-label">
        <span class="num">{setMission.num}</span>
        <span class="name">{setMission.name}</span>
        <span class="env">{setMission.env}</span>
        <span class="set-badge">SET DROP · 3 PIECES</span>
      </div>
    </a>

    <!-- ── Trucker (transient — only during intro; explodes during act 1) ── -->
    {#if showIntroTrucker}
      <div class="mission-slot trucker-slot" class:is-collapsing={truckerCollapsing} bind:this={truckerSlotEl} aria-hidden="true">
        <div class="planet-wrap">
          <!-- planet wrapper is the explosion target — scales, blurs, rotates -->
          <div class="planet-explode-target" bind:this={truckerPlanetEl}>
            <MissionPlanet
              planetType="lunar"
              productUrl="/photos/product-hat.png"
              productScale={0.9}
              productOffsetY={0.18}
              glowColor="#C8B89A"
              rotationSpeed={0.0011}
              axialTilt={0.08}
            />
          </div>
          <div class="planet-halo" style="--glow:#C8B89A" bind:this={truckerHaloEl}></div>

          <!-- IMPLOSION layers — collapse rings + singularity dot -->
          <div class="collapse-ring c1" bind:this={collapseRing1El}></div>
          <div class="collapse-ring c2" bind:this={collapseRing2El}></div>
          <div class="collapse-ring c3" bind:this={collapseRing3El}></div>
          <div class="singularity"      bind:this={singularityEl}></div>

          <!-- SUPERNOVA layers — flash + 3 shockwaves -->
          <div class="flash"      bind:this={flashEl}></div>
          <div class="shockwave"  bind:this={shockwaveEl}></div>
          <div class="shockwave2" bind:this={shockwave2El}></div>
          <div class="shockwave3" bind:this={shockwave3El}></div>

          <!-- Radial streaks — 24 lines fanning outward -->
          <div class="streaks" bind:this={streaksEl}>
            {#each Array(24) as _, i}
              <span
                class="streak"
                style="
                  --angle: {(i * 15)}deg;
                  --delay: {(i % 6) * 0.02}s;
                "
              ></span>
            {/each}
          </div>

          <div class="particles" bind:this={particlesEl}>
            {#each Array(36) as _, i}
              <span
                class="particle"
                style="
                  --angle: {(i * 10)}deg;
                  --delay: {(i % 8) * 0.025}s;
                  --dist: {110 + (i % 5) * 38}px;
                  --size: {3 + (i % 4)}px;
                  --hue: {i % 3 === 0 ? '#fff5d8' : i % 3 === 1 ? '#ffb060' : '#ff7028'};
                "
              ></span>
            {/each}
            {#each Array(8) as _, i}
              <span
                class="ember"
                style="
                  --angle: {(i * 45) + 22}deg;
                  --delay: {i * 0.04}s;
                  --dist: {180 + i * 18}px;
                "
              ></span>
            {/each}
          </div>
        </div>
        <div class="mission-label is-decommissioning">
          <span class="num">002</span>
          <span class="name">Vanguard Trucker Hat</span>
          <span class="env decommission-text">DECOMMISSIONED</span>
        </div>
      </div>
    {/if}

    <!-- ── Tank (permanent slot) ── -->
    <a href="/shop/{tankMission.slug}" class="mission-slot tank-slot" bind:this={tankSlotEl}>
      <div class="planet-wrap">
        <MissionPlanet
          planetType={tankMission.planetType}
          productUrl={tankMission.product}
          productScale={tankMission.productScale}
          productBlending={tankMission.spriteBlending}
          glowColor={tankMission.glow}
          rotationSpeed={tankMission.speed}
          axialTilt={tankMission.tilt}
        />
        <div class="planet-halo" style="--glow:{tankMission.glow}"></div>
      </div>
      <div class="mission-label">
        <span class="num">{tankMission.num}</span>
        <span class="name">{tankMission.name}</span>
        <span class="env">{tankMission.env}</span>
      </div>
    </a>

  </div>

  <footer class="bottom-bar">
    <a href="/" class="earth-link" aria-label="Return to Earth">
      <div class="earth-wrap">
        <MissionPlanet
          planetType="earth"
          photoUrl="/planet-leo.jpg"
          glowColor="#4FC3F7"
          rotationSpeed={0.0008}
          axialTilt={0.41}
        />
        <div class="earth-halo"></div>
      </div>
    </a>
    <span class="earth-label">EARTH</span>
    <a href="/" class="return-btn">↓ RETURN TO EARTH</a>
  </footer>

  {#if showIntroTrucker && !introPlayed}
    <button class="skip-intro" on:click={skipIntro} type="button">SKIP INTRO →</button>
  {/if}

</div>

<style>
  :global(body) { overflow: hidden; }

  .shop {
    position: fixed;
    inset: 0;
    background: #000005;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: space-between;
    z-index: 10;
    padding: 2rem 3rem 2.5rem;
  }

  /* ── Header ── */
  .top-bar { width: 100%; display: flex; justify-content: center; }

  .select-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.55em;
    color: rgba(200, 146, 42, 0.65);
  }

  /* ── Planets row ── */
  .planets-row {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: clamp(3rem, 7vw, 7rem);
    flex: 1;
    width: 100%;
  }

  @media (min-width: 641px) {
    .wr-slot      { transform: translateY(-3vh); }
    .trucker-slot { transform: translateY(2vh);  }
    .tank-slot    { transform: translateY(-2vh); }
  }

  @media (max-width: 640px) {
    :global(body) { overflow-y: auto; }
    .shop { position: relative; min-height: 100dvh; padding: 5rem 1.5rem 3rem; justify-content: flex-start; }
    .planets-row { flex-direction: column; gap: 2rem; align-items: center; justify-content: flex-start; flex: unset; padding: 1.5rem 0 2rem; }
    .planet-wrap { width: clamp(150px, 60vw, 220px); height: clamp(150px, 60vw, 220px); }
    .mission-slot:hover { transform: none !important; }
    .bottom-bar { margin-top: 1rem; }
  }

  .mission-slot {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.4rem;
    text-decoration: none;
    cursor: pointer;
    transition: transform 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  }

  /* Trucker slot collapse — smooth reflow instead of display:none snap.
     Width + margin animate to 0; sibling slots reposition naturally via flex. */
  .trucker-slot {
    transition:
      width   1.1s cubic-bezier(0.5, 0, 0.25, 1),
      max-width 1.1s cubic-bezier(0.5, 0, 0.25, 1),
      margin  1.1s cubic-bezier(0.5, 0, 0.25, 1),
      opacity 0.5s ease;
  }
  .trucker-slot.is-collapsing {
    width: 0 !important;
    max-width: 0 !important;
    margin-left:  calc(clamp(3rem, 7vw, 7rem) * -1) !important;
    overflow: hidden;
    pointer-events: none;
  }

  .mission-slot:hover { transform: translateY(-8px) !important; }

  /* Set + Tank are matched siblings post-decommission. */
  .wr-slot .planet-wrap,
  .tank-slot .planet-wrap {
    width: clamp(220px, 25vw, 340px);
    height: clamp(220px, 25vw, 340px);
  }

  /* Trucker uses the base size during intro. */
  .planet-wrap {
    position: relative;
    width: clamp(160px, 18vw, 260px);
    height: clamp(160px, 18vw, 260px);
  }

  /* Diffuse outer glow */
  .planet-halo {
    position: absolute;
    inset: -40%;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      color-mix(in srgb, var(--glow) 28%, transparent) 0%,
      color-mix(in srgb, var(--glow) 10%, transparent) 35%,
      transparent 68%
    );
    pointer-events: none;
    transition: opacity 0.4s ease;
    opacity: 0.55;
    z-index: -1;
  }

  .mission-slot:hover .planet-halo { opacity: 0.95; }

  /* ── Labels ── */
  .mission-label {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.3rem;
    text-align: center;
  }

  .num {
    font-family: 'Inter', sans-serif;
    font-size: 0.5rem;
    letter-spacing: 0.35em;
    color: rgba(200, 146, 42, 0.6);
  }

  .name {
    font-family: 'Cormorant Garamond', serif;
    font-size: clamp(0.9rem, 1.4vw, 1.25rem);
    font-weight: 300;
    color: #F0EDE6;
    letter-spacing: 0.04em;
  }

  .env {
    font-family: 'Inter', sans-serif;
    font-size: 0.46rem;
    letter-spacing: 0.22em;
    color: rgba(240, 237, 230, 0.3);
    text-transform: uppercase;
  }

  .set-badge {
    margin-top: 0.2rem;
    font-family: 'Inter', sans-serif;
    font-size: 0.45rem;
    letter-spacing: 0.30em;
    color: rgba(200, 146, 42, 0.85);
    border: 1px solid rgba(200, 146, 42, 0.45);
    padding: 0.2rem 0.7rem;
    text-transform: uppercase;
  }

  /* ── Explosion VFX (act 1 of intro) ── */

  /* Wrapper that gsap scales/rotates/blurs to "shatter" the planet. */
  .planet-explode-target {
    position: absolute;
    inset: 0;
    transform-origin: 50% 50%;
    will-change: transform, filter;
  }

  /* Bright white radial flash — the burst itself. */
  .flash {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: radial-gradient(circle,
      rgba(255, 255, 255, 1.0) 0%,
      rgba(255, 240, 200, 0.95) 18%,
      rgba(255, 180, 90, 0.7) 38%,
      rgba(255, 90, 30, 0.35) 60%,
      transparent 78%);
    transform-origin: 50% 50%;
    pointer-events: none;
    mix-blend-mode: screen;
    will-change: transform, opacity;
    z-index: 6;
    filter: blur(2px);
  }

  /* Primary shockwave — bright orange ring. */
  .shockwave {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 3px solid rgba(255, 170, 80, 0.95);
    box-shadow:
      0 0 40px rgba(255, 130, 50, 0.7),
      inset 0 0 20px rgba(255, 200, 120, 0.4);
    transform-origin: 50% 50%;
    pointer-events: none;
    will-change: transform, opacity;
    z-index: 5;
  }

  /* Secondary shockwave — softer, trails primary. */
  .shockwave2 {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid rgba(255, 220, 160, 0.6);
    box-shadow: 0 0 60px rgba(255, 150, 70, 0.4);
    transform-origin: 50% 50%;
    pointer-events: none;
    will-change: transform, opacity;
    z-index: 4;
  }

  /* Tertiary shockwave — faintest, biggest, last to fade. */
  .shockwave3 {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid rgba(255, 230, 180, 0.4);
    box-shadow: 0 0 80px rgba(255, 140, 60, 0.25);
    transform-origin: 50% 50%;
    pointer-events: none;
    will-change: transform, opacity;
    z-index: 3;
  }

  /* ── Implosion VFX ── */

  /* Collapse rings — start outside the planet, crash inward to centre. */
  .collapse-ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    transform-origin: 50% 50%;
    pointer-events: none;
    will-change: transform, opacity;
    z-index: 5;
  }
  .collapse-ring.c1 {
    border: 2px solid rgba(255, 200, 140, 0.95);
    box-shadow: 0 0 24px rgba(255, 160, 80, 0.7),
                inset 0 0 12px rgba(255, 200, 120, 0.4);
  }
  .collapse-ring.c2 {
    border: 1px solid rgba(255, 220, 160, 0.85);
    box-shadow: 0 0 18px rgba(255, 180, 100, 0.6);
  }
  .collapse-ring.c3 {
    border: 1px solid rgba(255, 240, 200, 0.65);
    box-shadow: 0 0 14px rgba(255, 200, 120, 0.5);
  }

  /* Singularity — bright pinpoint at the implosion peak. */
  .singularity {
    position: absolute;
    top: 50%; left: 50%;
    width: 8px; height: 8px;
    margin: -4px 0 0 -4px;
    border-radius: 50%;
    background: #fff;
    box-shadow:
      0 0 20px #fff,
      0 0 40px rgba(255, 220, 160, 0.95),
      0 0 80px rgba(255, 140, 50, 0.7),
      0 0 140px rgba(200, 80, 30, 0.4);
    transform-origin: 50% 50%;
    pointer-events: none;
    mix-blend-mode: screen;
    will-change: transform, opacity;
    z-index: 8;
  }

  /* Radial debris streaks — lines fanning out from centre at the burst. */
  .streaks {
    position: absolute;
    inset: 0;
    pointer-events: none;
    will-change: opacity;
    z-index: 6;
  }
  .streak {
    position: absolute;
    top: 50%; left: 50%;
    width: 110px; height: 2px;
    background: linear-gradient(90deg,
      rgba(255, 255, 255, 0) 0%,
      rgba(255, 220, 160, 0.95) 18%,
      rgba(255, 140, 60, 0.95) 60%,
      rgba(255, 80, 30, 0) 100%);
    box-shadow: 0 0 8px rgba(255, 140, 60, 0.7);
    transform-origin: 0% 50%;
    opacity: 0;
    animation: streak-fly 0.85s cubic-bezier(0.18, 0.7, 0.4, 1) var(--delay, 0s) forwards;
  }
  @keyframes streak-fly {
    0%   { opacity: 0;   transform: rotate(var(--angle)) translateX(0)    scaleX(0.2); }
    18%  { opacity: 1;   transform: rotate(var(--angle)) translateX(20px)  scaleX(0.5); }
    60%  { opacity: 0.85;transform: rotate(var(--angle)) translateX(120px) scaleX(1.4); }
    100% { opacity: 0;   transform: rotate(var(--angle)) translateX(220px) scaleX(2.2); }
  }

  .particles {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 7;
    will-change: opacity;
  }

  .particle {
    position: absolute;
    top: 50%;
    left: 50%;
    width: var(--size, 4px);
    height: var(--size, 4px);
    border-radius: 50%;
    background: var(--hue, #f0d7a3);
    box-shadow: 0 0 12px var(--hue, #f0d7a3),
                0 0 24px color-mix(in srgb, var(--hue, #f0d7a3) 60%, transparent);
    transform: translate(-50%, -50%);
    animation: particle-fly 1.05s cubic-bezier(0.18, 0.7, 0.4, 1) var(--delay, 0s) forwards;
    opacity: 0;
  }

  @keyframes particle-fly {
    0%   { transform: translate(-50%, -50%) rotate(var(--angle)) translateX(0)                scale(1.4); opacity: 1; }
    20%  { opacity: 1; }
    70%  { opacity: 0.9; }
    100% { transform: translate(-50%, -50%) rotate(var(--angle)) translateX(var(--dist, 80px)) scale(0.4); opacity: 0; }
  }

  /* Larger flung debris embers — fewer, slower, longer reach. */
  .ember {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: radial-gradient(circle, #fff8d8 0%, #ffaa50 40%, #c84418 100%);
    box-shadow: 0 0 16px rgba(255, 130, 60, 0.85),
                0 0 32px rgba(255, 80, 30, 0.4);
    transform: translate(-50%, -50%);
    animation: ember-fly 1.4s cubic-bezier(0.15, 0.65, 0.45, 1) var(--delay, 0s) forwards;
    opacity: 0;
  }

  @keyframes ember-fly {
    0%   { transform: translate(-50%, -50%) rotate(var(--angle)) translateX(0)                  scale(1.6); opacity: 1; }
    100% { transform: translate(-50%, -50%) rotate(var(--angle)) translateX(var(--dist, 180px)) scale(0.2); opacity: 0; }
  }

  .decommission-text {
    color: rgba(200, 60, 60, 0.7) !important;
    letter-spacing: 0.3em !important;
  }

  /* ── Skip intro button ── */
  .skip-intro {
    position: absolute;
    bottom: 1.2rem;
    right: 1.4rem;
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.32em;
    color: rgba(240, 237, 230, 0.55);
    background: rgba(0, 0, 0, 0.5);
    border: 1px solid rgba(240, 237, 230, 0.18);
    padding: 0.55rem 1.1rem;
    backdrop-filter: blur(4px);
    cursor: pointer;
    text-transform: uppercase;
    transition: color 0.2s, border-color 0.2s;
    z-index: 50;
  }

  .skip-intro:hover {
    color: #F0EDE6;
    border-color: rgba(240, 237, 230, 0.5);
  }

  /* ── Footer / Earth ── */
  .bottom-bar {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
  }

  .earth-link { display: block; text-decoration: none; transition: transform 0.3s ease; }
  .earth-link:hover { transform: scale(1.12); }

  .earth-wrap {
    position: relative;
    width: clamp(44px, 5vw, 64px);
    height: clamp(44px, 5vw, 64px);
  }

  .earth-halo {
    position: absolute;
    inset: -50%;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      rgba(79,195,247,0.22) 0%,
      rgba(79,195,247,0.06) 40%,
      transparent 68%
    );
    pointer-events: none;
    z-index: -1;
  }

  .earth-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.46rem;
    letter-spacing: 0.35em;
    color: rgba(240,237,230,0.25);
  }

  .return-btn {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.22em;
    color: rgba(240,237,230,0.4);
    text-decoration: none;
    border: 1px solid rgba(240,237,230,0.12);
    padding: 0.6rem 1.6rem;
    background: rgba(0,0,0,0.4);
    backdrop-filter: blur(4px);
    transition: color 0.2s, border-color 0.2s;
  }

  .return-btn:hover {
    color: #F0EDE6;
    border-color: rgba(240,237,230,0.4);
  }

  /* set-cluster-marker is a hidden hook for the GSAP timeline to fade in.
     Visual cluster comes from the in-scene Three.js sprites. */
  .set-cluster-marker {
    position: absolute; inset: 0; pointer-events: none;
  }
</style>
