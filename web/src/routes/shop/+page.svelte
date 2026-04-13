<!-- web/src/routes/shop/+page.svelte -->
<script lang="ts">
  import MissionPlanet from '$lib/components/MissionPlanet.svelte';

  const R2 = 'https://pub-75a66fca0ddd4d93b3bb53bda5d6a29c.r2.dev';

  const missions = [
    {
      num: '001',
      name: 'Warped Reality Beanie',
      env: 'Low Earth Orbit',
      slug: 'warped-reality-beanie',
      planetType: 'leo',
      product: '/photos/product-beanie.png',
      productScale: 0.75,
      glow: '#4FC3F7',
      speed: 0.0018,
      tilt: 0.22,
    },
    {
      num: '002',
      name: 'Vanguard Trucker Hat',
      env: 'Lunar Surface',
      slug: 'vanguard-trucker-hat',
      planetType: 'lunar',
      product: '/photos/product-hat.png',
      productScale: 1.5,
      planetScale: 0.72,
      planetOffsetY: 0.04,
      glow: '#C8B89A',
      speed: 0.0011,
      tilt: 0.08,
      condemned: true, // sold out — flip to false when restocked
    },
    {
      num: '003',
      name: 'Racerback Tanktop',
      env: 'Stellar Nursery',
      slug: 'racerback-tanktop',
      planetType: 'nebula',
      product: '/photos/product-tank.png',
      productScale: 0.75,
      spriteBlending: 'additive',
      glow: '#6B0FCC',
      speed: 0.0022,
      tilt: 0.32,
    },
  ];
</script>

<svelte:head>
  <title>Immortal Vibes — Select Your Mission</title>
</svelte:head>

<a href="/" class="shop-logo" aria-label="Immortal Vibes home">
  <img src="/logo-bare.png" alt="Immortal Vibes" />
</a>

<div class="shop">

  <header class="top-bar">
    <span class="select-label">MISSION SELECT</span>
  </header>

  <div class="planets-row">
    {#each missions as m}
      <a href="/shop/{m.slug}" class="mission-slot">

        <div class="planet-wrap">
          <MissionPlanet
            planetType={m.planetType}
            productUrl={m.product}
            productScale={m.planetScale ?? m.productScale ?? 1.0}
            productOffsetY={m.planetOffsetY ?? 0}
            productBlending={m.spriteBlending ?? 'normal'}
            glowColor={m.glow}
            rotationSpeed={m.speed}
            axialTilt={m.tilt}
          />
          <div class="planet-halo" style="--glow:{m.glow}"></div>

          {#if m.condemned}
            <!-- rings sit outside the clipping shell -->
            <div class="ring-outer"></div>
            <div class="ring-inner"></div>
            <div class="alert-dot"></div>
            <!-- shell clips tape + overlays to the circle -->
            <div class="condemned-shell">
              <div class="condemned-dim"></div>
              <div class="scanlines"></div>
              <div class="danger-atmo"></div>
              <div class="tape tape-1"></div>
              <div class="tape tape-2"></div>
              <div class="tape-text">CONDEMNED</div>
            </div>
          {/if}
        </div>

        <div class="mission-label" class:is-condemned={m.condemned}>
          <span class="num">{m.num}</span>
          <span class="name">{m.name}</span>
          {#if m.condemned}
            <div class="condemned-badge">Condemned</div>
            <span class="mission-terminated">Mission Terminated</span>
          {:else}
            <span class="env">{m.env}</span>
          {/if}
        </div>

      </a>
    {/each}
  </div>

  <footer class="bottom-bar">
    <!-- Clicking the Earth returns home -->
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

</div>

<style>
  :global(body) { overflow: hidden; }

  .shop-logo {
    position: fixed;
    top: 1.2rem;
    left: 1.8rem;
    z-index: 20;
    opacity: 0.7;
    transition: opacity 0.2s ease;
  }

  .shop-logo:hover { opacity: 1; }

  .shop-logo img {
    width: 52px;
    height: 52px;
    object-fit: contain;
  }

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
  .top-bar {
    width: 100%;
    display: flex;
    justify-content: center;
  }

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
    .mission-slot:nth-child(1) { transform: translateY(-3vh); }
    .mission-slot:nth-child(2) { transform: translateY(2vh);  }
    .mission-slot:nth-child(3) { transform: translateY(-2vh); }
  }

  @media (max-width: 640px) {
    :global(body) { overflow-y: auto; }
    .shop { position: relative; min-height: 100dvh; padding: 5rem 1.5rem 3rem; justify-content: flex-start; }
    .planets-row { flex-direction: column; gap: 2rem; align-items: center; justify-content: flex-start; flex: unset; padding: 1.5rem 0 2rem; }
    .planet-wrap { width: clamp(150px, 60vw, 220px); height: clamp(150px, 60vw, 220px); }
    .mission-slot:hover { transform: none !important; }
    .shop-logo { top: 0.9rem; left: 1rem; }
    .shop-logo img { width: 40px; height: 40px; }
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

  .mission-slot:hover { transform: translateY(-8px) !important; }

  /* ── Planet + overlay ── */
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

  /* ── Condemned overlay ── */
  .condemned-shell {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    overflow: hidden;
    pointer-events: none;
  }

  .condemned-dim {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.65);
    border-radius: 50%;
  }

  .scanlines {
    position: absolute;
    inset: 0;
    background: repeating-linear-gradient(
      0deg,
      transparent 0px, transparent 3px,
      rgba(0, 0, 0, 0.18) 3px, rgba(0, 0, 0, 0.18) 4px
    );
    border-radius: 50%;
  }

  .danger-atmo {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: radial-gradient(circle at 50% 50%,
      transparent 40%, rgba(160, 30, 30, 0.22) 100%);
  }

  .tape {
    position: absolute;
    width: 230%;
    height: 14px;
    left: -65%;
    transform-origin: center center;
    background: repeating-linear-gradient(
      90deg,
      rgba(14, 14, 0, 0.72) 0px,  rgba(14, 14, 0, 0.72) 14px,
      rgba(200, 146, 42, 0.65) 14px, rgba(200, 146, 42, 0.65) 28px
    );
  }

  .tape-1 { top: calc(50% - 7px); transform: rotate(-33deg); }
  .tape-2 { top: calc(50% - 7px); transform: rotate(33deg); }

  .tape-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%) rotate(-33deg);
    font-family: 'Inter', sans-serif;
    font-size: 0.38rem;
    font-weight: 700;
    letter-spacing: 0.35em;
    color: rgba(14, 14, 0, 0.85);
    background: rgba(200, 146, 42, 0.75);
    padding: 0.1rem 0.5rem;
    white-space: nowrap;
    text-transform: uppercase;
    z-index: 2;
  }

  .ring-outer {
    position: absolute;
    inset: -6px;
    border-radius: 50%;
    border: 1px solid rgba(190, 45, 45, 0.55);
    box-shadow: 0 0 12px rgba(190,45,45,0.25), inset 0 0 10px rgba(190,45,45,0.08);
    pointer-events: none;
    animation: flicker 2.6s ease-in-out infinite;
  }

  .ring-inner {
    position: absolute;
    inset: -2px;
    border-radius: 50%;
    border: 1px solid rgba(190, 45, 45, 0.18);
    pointer-events: none;
    animation: flicker 2.6s ease-in-out infinite reverse;
  }

  .alert-dot {
    position: absolute;
    top: 10px;
    right: 10px;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #cc2222;
    box-shadow: 0 0 8px rgba(200,30,30,0.9);
    z-index: 10;
    pointer-events: none;
    animation: blink 1.4s ease-in-out infinite;
  }

  @keyframes flicker {
    0%, 100% { opacity: 0.4; }
    30%       { opacity: 1; }
    65%       { opacity: 0.6; }
    80%       { opacity: 0.9; }
  }

  @keyframes blink {
    0%, 100% { opacity: 1; }
    50%       { opacity: 0.1; }
  }

  .mission-label.is-condemned .name { color: rgba(240, 237, 230, 0.25); }
  .mission-label.is-condemned .num  { color: rgba(200, 146, 42, 0.25); }

  .condemned-badge {
    font-family: 'Inter', sans-serif;
    font-size: 0.46rem;
    letter-spacing: 0.28em;
    color: rgba(190, 45, 45, 0.7);
    border: 1px solid rgba(190, 45, 45, 0.3);
    padding: 0.2rem 0.8rem;
    text-transform: uppercase;
    position: relative;
  }

  .condemned-badge::before,
  .condemned-badge::after {
    content: '';
    position: absolute;
    width: 5px;
    height: 5px;
  }
  .condemned-badge::before {
    top: -1px; left: -1px;
    border-top: 1px solid rgba(190,45,45,0.55);
    border-left: 1px solid rgba(190,45,45,0.55);
  }
  .condemned-badge::after {
    bottom: -1px; right: -1px;
    border-bottom: 1px solid rgba(190,45,45,0.55);
    border-right: 1px solid rgba(190,45,45,0.55);
  }

  .mission-terminated {
    font-family: 'Inter', sans-serif;
    font-size: 0.4rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.18);
    text-transform: uppercase;
  }

  /* ── Footer / Earth ── */
  .bottom-bar {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
  }

  .earth-link {
    display: block;
    text-decoration: none;
    transition: transform 0.3s ease;
  }

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
</style>
