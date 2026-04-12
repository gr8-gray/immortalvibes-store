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
        </div>

        <div class="mission-label">
          <span class="num">{m.num}</span>
          <span class="name">{m.name}</span>
          <span class="env">{m.env}</span>
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
