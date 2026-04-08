<!-- web/src/lib/components/MissionCard.svelte -->
<script lang="ts">
  import { goto } from '$app/navigation';
  import { transitionStore, MISSION_ACCENT } from '$lib/stores/transition';

  export let missionNumber: '001' | '002' | '003' | '004';
  export let title: string;
  export let subtitle: string;
  export let slug: string;
  export let imageUrl: string = '';
  export let status: 'available' | 'sold_out' | 'coming_soon' = 'available';

  interface MissionEnv {
    bg: string;
    accentColor: string;
    planetStyle: string;
    label: string;
  }

  const envs: Record<string, MissionEnv> = {
    '001': {
      bg: 'linear-gradient(160deg, #000814 0%, #050d1e 45%, #0a1628 100%)',
      accentColor: '#4FC3F7',
      label: 'LOW EARTH ORBIT',
      planetStyle: 'position:absolute;bottom:-15%;left:-10%;width:65%;aspect-ratio:1;border-radius:50%;background:radial-gradient(ellipse at 42% 38%,rgba(79,195,247,0.55) 0%,rgba(30,100,200,0.38) 35%,rgba(10,40,100,0.22) 58%,transparent 75%);box-shadow:0 0 40px rgba(79,195,247,0.12);',
    },
    '002': {
      bg: 'linear-gradient(160deg, #050505 0%, #0e0c0a 50%, #1a1814 100%)',
      accentColor: 'rgba(200,190,180,0.8)',
      label: 'LUNAR SURFACE',
      planetStyle: 'position:absolute;bottom:-5%;left:-5%;right:-5%;height:28%;background:linear-gradient(to top,rgba(160,150,140,0.18),transparent);border-top:1px solid rgba(160,150,140,0.08);',
    },
    '003': {
      bg: 'linear-gradient(160deg, #030308 0%, #1a0800 50%, #2a0d00 100%)',
      accentColor: 'rgba(255,130,50,0.9)',
      label: 'STELLAR NURSERY',
      planetStyle: 'position:absolute;inset:0;background:radial-gradient(ellipse at 65% 25%,rgba(255,80,20,0.22) 0%,transparent 55%),radial-gradient(ellipse at 30% 60%,rgba(180,40,0,0.18) 0%,transparent 50%);',
    },
    '004': {
      bg: 'linear-gradient(160deg, #030308 0%, #1a0500 50%, #2a0800 100%)',
      accentColor: 'rgba(180,80,40,0.7)',
      label: 'DEEP SPACE',
      planetStyle: 'position:absolute;bottom:-20%;left:50%;transform:translateX(-50%);width:80%;aspect-ratio:1;border-radius:50%;background:radial-gradient(ellipse at 45% 40%,rgba(180,60,20,0.35) 0%,rgba(120,30,5,0.2) 40%,transparent 65%);',
    },
  };

  $: env = envs[missionNumber] ?? envs['004'];
  $: showEarthrise = missionNumber === '002';

  let card: HTMLDivElement;
  let rx = 0;
  let ry = 0;
  const MAX_TILT = 10;

  function handleMouseMove(e: MouseEvent) {
    const rect = card.getBoundingClientRect();
    ry = ((e.clientX - rect.left) / rect.width - 0.5) * 2 * MAX_TILT;
    rx = -((e.clientY - rect.top) / rect.height - 0.5) * 2 * MAX_TILT;
  }

  function handleMouseLeave() {
    rx = 0;
    ry = 0;
  }

  function handleClick(e: MouseEvent) {
    if (!slug) return;
    const rect = card.getBoundingClientRect();
    transitionStore.update((s) => ({
      ...s,
      clickX: (rect.left + rect.width / 2) / window.innerWidth,
      clickY: (rect.top + rect.height / 2) / window.innerHeight,
      missionAccent: MISSION_ACCENT[slug] ?? '#4FC3F7',
    }));
    goto(`/shop/${slug}`);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') handleClick(e as unknown as MouseEvent);
  }
</script>

<div
  bind:this={card}
  class="mission-card"
  style="background: {env.bg}; transform: perspective(900px) rotateX({rx}deg) rotateY({ry}deg); --accent: {env.accentColor};"
  on:mousemove={handleMouseMove}
  on:mouseleave={handleMouseLeave}
  on:click={handleClick}
  on:keydown={handleKeydown}
  role="button"
  tabindex="0"
  aria-label="Explore {title}"
>
  <div class="env-layer" style={env.planetStyle}></div>

  {#if showEarthrise}
    <div class="earthrise"></div>
  {/if}

  <svg class="card-stars" aria-hidden="true" xmlns="http://www.w3.org/2000/svg">
    {#each Array(35) as _, i}
      <circle
        cx="{(i * 137.508) % 100}%"
        cy="{(i * 97.3) % 100}%"
        r="{0.5 + (i % 3) * 0.35}"
        fill="rgba(240,237,230,{0.25 + (i % 5) * 0.12})"
      />
    {/each}
  </svg>

  {#if imageUrl}
    <div class="product-zone">
      <img
        class="product-img"
        src={imageUrl}
        alt={title}
        style="filter: drop-shadow(0 20px 50px {env.accentColor});"
      />
    </div>
  {:else}
    <div
      class="product-placeholder"
      style="background: radial-gradient(ellipse, {env.accentColor.replace(/[\d.]+\)$/, '0.12)')}, transparent 70%);"
    ></div>
  {/if}

  <div class="scanlines" aria-hidden="true"></div>
  <div class="accent-rim"></div>

  <div class="card-info">
    <p class="mission-num">MISSION {missionNumber}</p>
    <p class="location" style="color: {env.accentColor};">{env.label}</p>
    <h2 class="product-title">{title}</h2>
    <div class="card-footer">
      {#if status === 'sold_out'}
        <span class="badge badge-sold">SOLD OUT</span>
      {:else if status === 'coming_soon'}
        <span class="badge badge-soon">COMING SOON</span>
      {:else}
        <span class="cta">EXPLORE MISSION →</span>
      {/if}
    </div>
  </div>
</div>

<style>
  .mission-card {
    position: relative;
    overflow: hidden;
    border: 1px solid rgba(240,237,230,0.06);
    border-radius: 4px;
    cursor: pointer;
    aspect-ratio: 3/4;
    transition: transform 0.3s ease, box-shadow 0.3s ease, border-color 0.3s ease;
    will-change: transform;
  }
  .mission-card:hover {
    border-color: color-mix(in srgb, var(--accent) 35%, transparent);
    box-shadow: 0 20px 60px color-mix(in srgb, var(--accent) 15%, transparent);
  }
  .env-layer { pointer-events: none; }
  .earthrise {
    position: absolute;
    top: 8%; right: 10%;
    width: 22%; aspect-ratio: 1;
    border-radius: 50%;
    background: radial-gradient(ellipse at 40% 38%, rgba(79,195,247,0.5) 0%, rgba(30,80,180,0.38) 40%, rgba(10,30,100,0.2) 60%, transparent 75%);
    box-shadow: 0 0 16px rgba(79,195,247,0.2);
  }
  .card-stars {
    position: absolute; inset: 0;
    width: 100%; height: 100%;
    pointer-events: none;
  }
  .product-zone {
    position: absolute;
    top: 6%; left: 50%; transform: translateX(-50%);
    width: 70%;
    display: flex; align-items: center; justify-content: center;
    pointer-events: none;
    transition: transform 0.4s ease;
  }
  .mission-card:hover .product-zone {
    transform: translateX(-50%) translateY(-8px) scale(1.04);
  }
  .product-img {
    width: 100%;
    max-height: 52%;
    object-fit: contain;
    transform: rotate(-5deg);
    transition: transform 0.4s ease;
  }
  .mission-card:hover .product-img {
    transform: rotate(-2deg);
  }
  .product-placeholder {
    position: absolute;
    top: 10%; left: 20%; right: 20%; height: 45%;
    border-radius: 50%;
    pointer-events: none;
  }
  .scanlines {
    position: absolute; inset: 0;
    background: repeating-linear-gradient(to bottom, transparent, transparent 2px, rgba(0,0,0,0.025) 2px, rgba(0,0,0,0.025) 4px);
    pointer-events: none;
    z-index: 2;
  }
  .accent-rim {
    position: absolute; inset: 0;
    border-radius: 4px;
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 0%, transparent);
    transition: box-shadow 0.3s ease;
    pointer-events: none;
    z-index: 3;
  }
  .mission-card:hover .accent-rim {
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 30%, transparent);
  }
  .card-info {
    position: absolute;
    bottom: 0; left: 0; right: 0;
    padding: 1.25rem;
    background: linear-gradient(to top, rgba(3,3,8,0.97) 0%, rgba(3,3,8,0.85) 38%, rgba(3,3,8,0.4) 62%, transparent 100%);
    z-index: 4;
  }
  .mission-num {
    font-family: 'Inter', sans-serif;
    font-size: 0.52rem; letter-spacing: 0.25em;
    color: rgba(240,237,230,0.3);
    margin: 0 0 0.12rem; text-transform: uppercase;
  }
  .location {
    font-family: 'Inter', sans-serif;
    font-size: 0.58rem; letter-spacing: 0.18em;
    text-transform: uppercase; margin: 0 0 0.35rem;
  }
  .product-title {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.35rem; font-weight: 300;
    color: #F0EDE6; margin: 0 0 0.65rem; line-height: 1.1;
  }
  .card-footer { display: flex; align-items: center; }
  .cta {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem; letter-spacing: 0.18em;
    color: rgba(240,237,230,0.5);
    transition: color 0.2s; text-transform: uppercase;
  }
  .mission-card:hover .cta { color: rgba(240,237,230,0.85); }
  .badge {
    font-family: 'Inter', sans-serif;
    font-size: 0.58rem; letter-spacing: 0.18em;
    padding: 0.2rem 0.65rem; border-radius: 2px; text-transform: uppercase;
  }
  .badge-sold { border: 1px solid rgba(240,237,230,0.25); color: rgba(240,237,230,0.45); }
  .badge-soon { border: 1px solid rgba(240,237,230,0.15); color: rgba(240,237,230,0.35); }
</style>
