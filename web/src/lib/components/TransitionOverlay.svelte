<!-- web/src/lib/components/TransitionOverlay.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import gsap from 'gsap';
  import { playT1Out, playT1In, type T1Elements } from '$lib/transitions/t1-ascent';
  import { playT2, type T2Elements } from '$lib/transitions/t2-hyperspace';
  import { playT3, type T3Elements } from '$lib/transitions/t3-ring';
  import { playT4, type T4Elements } from '$lib/transitions/t4-return';

  export function triggerOut(
    type: 'T1' | 'T2' | 'T3' | 'T4',
    opts: { clickX?: number; clickY?: number; accentColor?: string; mainContent?: HTMLElement },
    onMidpoint: () => void
  ): Promise<void> {
    return new Promise((resolve) => {
      if (!browser) { resolve(); return; }
      if (type === 'T1') {
        playT1Out(getT1Els(), onMidpoint);
        resolve();
      } else if (type === 'T2') {
        playT2(getT2Els(), opts.clickX ?? 0.5, opts.clickY ?? 0.5, opts.accentColor ?? '#4FC3F7', onMidpoint, resolve);
      } else if (type === 'T3') {
        playT3(getT3Els(), opts.mainContent!, onMidpoint, resolve);
      } else if (type === 'T4') {
        playT4(getT4Els(), onMidpoint, resolve);
      }
    });
  }

  export function triggerIn(type: 'T1' | 'T2' | 'T3' | 'T4', onComplete: () => void): void {
    if (!browser) { onComplete(); return; }
    if (type === 'T1') playT1In(getT1Els(), onComplete);
    else onComplete();
  }

  // ── DOM refs ──
  let overlayEl: HTMLElement;

  // T1
  let t1Flash: HTMLElement;
  let t1Horizon: HTMLElement;
  let t1StreakCanvas: HTMLCanvasElement;
  let t1AtmoLeft: HTMLElement;
  let t1AtmoRight: HTMLElement;

  // T2
  let t2StreakEls: HTMLElement[] = [];
  let t2Flash: HTMLElement;
  let t2Mist: HTMLElement;

  // T3
  let t3RingEls: HTMLElement[] = [];
  let t3Core: HTMLElement;
  let t3RayEls: HTMLElement[] = [];

  // T4
  let t4SpaceStars: HTMLElement;
  let t4Heat: HTMLElement;
  let t4Craft: HTMLElement;
  let t4Trail: HTMLElement;
  let t4Atmo: HTMLElement;
  let t4City: HTMLElement;

  function getT1Els(): T1Elements {
    return { overlay: overlayEl, flash: t1Flash, horizon: t1Horizon, streakCanvas: t1StreakCanvas, atmoLeft: t1AtmoLeft, atmoRight: t1AtmoRight };
  }
  function getT2Els(): T2Elements {
    return { overlay: overlayEl, streaks: t2StreakEls, flash: t2Flash, mist: t2Mist };
  }
  function getT3Els(): T3Elements {
    return { overlay: overlayEl, rings: t3RingEls, core: t3Core, rays: t3RayEls };
  }
  function getT4Els(): T4Elements {
    return { overlay: overlayEl, spaceStars: t4SpaceStars, heat: t4Heat, craft: t4Craft, trail: t4Trail, atmo: t4Atmo, cityline: t4City };
  }

  onMount(() => {
    if (!browser) return;
    gsap.set(overlayEl, { display: 'none' });
  });
</script>

<div bind:this={overlayEl} class="t-overlay" aria-hidden="true">

  <!-- T1 LAYERS -->
  <div class="t1-layer">
    <div bind:this={t1Flash} class="t1-flash"></div>
    <div bind:this={t1Horizon} class="t1-horizon"></div>
    <canvas bind:this={t1StreakCanvas} class="t1-streak-canvas"></canvas>
    <div bind:this={t1AtmoLeft} class="t1-atmo-left"></div>
    <div bind:this={t1AtmoRight} class="t1-atmo-right"></div>
  </div>

  <!-- T2 LAYERS -->
  <div class="t2-layer">
    <div bind:this={t2Mist} class="t2-mist"></div>
    {#each Array(16) as _, i}
      <div bind:this={t2StreakEls[i]} class="t2-streak"></div>
    {/each}
    <div bind:this={t2Flash} class="t2-flash"></div>
  </div>

  <!-- T3 LAYERS -->
  <div class="t3-layer">
    {#each Array(8) as _, i}
      <div bind:this={t3RayEls[i]} class="t3-ray" style="transform: rotate({i * 45}deg);"></div>
    {/each}
    {#each Array(5) as _, i}
      <div bind:this={t3RingEls[i]} class="t3-ring" style="width: {40 + i * 16}vmin; height: {40 + i * 16}vmin;"></div>
    {/each}
    <div bind:this={t3Core} class="t3-core"></div>
  </div>

  <!-- T4 LAYERS -->
  <div class="t4-layer">
    <div bind:this={t4SpaceStars} class="t4-space-stars">
      {#each Array(40) as _, i}
        <div
          class="t4-star"
          style="top: {(i * 73) % 60}%; left: {(i * 137) % 100}%; width: {i % 5 === 0 ? 2 : 1}px; height: {i % 5 === 0 ? 2 : 1}px; opacity: {0.4 + (i % 6) * 0.08};"
        ></div>
      {/each}
    </div>
    <div bind:this={t4Heat} class="t4-heat"></div>
    <div bind:this={t4Trail} class="t4-trail"></div>
    <div bind:this={t4Craft} class="t4-craft"></div>
    <div bind:this={t4Atmo} class="t4-atmo"></div>
    <div bind:this={t4City} class="t4-city">
      <svg viewBox="0 0 1440 120" preserveAspectRatio="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M0,120 L0,80 L40,80 L40,50 L60,50 L60,30 L80,30 L80,50 L100,50 L100,60 L140,60 L140,40 L160,40 L160,20 L180,20 L180,40 L200,40 L200,55 L240,55 L240,35 L260,35 L260,55 L280,55 L280,70 L320,70 L320,45 L350,45 L350,25 L370,25 L370,45 L390,45 L390,60 L420,60 L420,80 L460,80 L460,55 L490,55 L490,70 L520,70 L520,45 L540,45 L540,30 L560,30 L560,45 L580,45 L580,65 L620,65 L620,80 L660,80 L660,50 L680,50 L680,35 L700,35 L700,50 L720,50 L720,60 L760,60 L760,40 L800,40 L800,55 L840,55 L840,75 L880,75 L880,50 L910,50 L910,30 L930,30 L930,50 L960,50 L960,65 L1000,65 L1000,80 L1040,80 L1040,55 L1070,55 L1070,40 L1090,40 L1090,55 L1110,55 L1110,70 L1150,70 L1150,45 L1180,45 L1180,60 L1220,60 L1220,80 L1260,80 L1260,50 L1290,50 L1290,35 L1310,35 L1310,50 L1340,50 L1340,65 L1380,65 L1380,80 L1440,80 L1440,120 Z" fill="rgba(20,40,80,0.9)"/>
      </svg>
    </div>
  </div>

</div>

<style>
  .t-overlay {
    position: fixed;
    inset: 0;
    z-index: 9000;
    pointer-events: none;
    background: #000005;
    display: none;
  }

  /* T1 */
  .t1-layer { position: absolute; inset: 0; }

  .t1-flash {
    position: absolute;
    inset: 0;
    background: #F0EDE6;
    opacity: 0;
  }

  .t1-horizon {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 14vh;
    background: radial-gradient(
      ellipse at 50% 100%,
      rgba(8, 22, 65, 0.7) 0%,
      rgba(15, 45, 110, 0.5) 20%,
      rgba(25, 65, 140, 0.3) 40%,
      rgba(79, 195, 247, 0.08) 75%,
      transparent 90%
    );
    border-radius: 50% 50% 0 0 / 80% 80% 0 0;
  }

  .t1-streak-canvas {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
  }

  .t1-atmo-left {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 18vw;
    background: linear-gradient(
      to right,
      rgba(79, 195, 247, 0.45) 0%,
      rgba(30, 100, 200, 0.25) 40%,
      transparent 100%
    );
    opacity: 0;
  }

  .t1-atmo-right {
    position: absolute;
    top: 0;
    bottom: 0;
    right: 0;
    width: 18vw;
    background: linear-gradient(
      to left,
      rgba(79, 195, 247, 0.45) 0%,
      rgba(30, 100, 200, 0.25) 40%,
      transparent 100%
    );
    opacity: 0;
  }

  /* T2 */
  .t2-layer { position: absolute; inset: 0; }
  .t2-mist { position: absolute; inset: 0; opacity: 0; }
  .t2-streak {
    position: absolute;
    height: 1px;
    background: linear-gradient(to right, transparent 0%, rgba(160,210,255,0.9) 30%, rgba(255,255,255,0.95) 60%, transparent 100%);
    opacity: 0;
    transform-origin: 0% 50%;
  }
  .t2-flash { position: absolute; inset: 0; background: #ffffff; opacity: 0; }

  /* T3 */
  .t3-layer { position: absolute; inset: 0; }
  .t3-ring {
    position: absolute;
    top: 50%; left: 50%;
    transform: translate(-50%, -50%) scale(0);
    border-radius: 50%;
    border: 1px solid #C8922A;
    box-shadow: 0 0 12px rgba(200,146,42,0.25);
    opacity: 0;
  }
  .t3-core {
    position: absolute;
    top: 50%; left: 50%;
    transform: translate(-50%, -50%) scale(0);
    width: 12px; height: 12px;
    border-radius: 50%;
    background: radial-gradient(ellipse, rgba(200,146,42,1), rgba(200,146,42,0.3) 70%);
    box-shadow: 0 0 30px rgba(200,146,42,0.8);
    opacity: 0;
  }
  .t3-ray {
    position: absolute;
    top: 50%; left: 50%;
    height: 1px; width: 50vw;
    transform-origin: 0% 50%;
    background: linear-gradient(to right, rgba(200,146,42,0.5), transparent);
    opacity: 0;
  }

  /* T4 */
  .t4-layer { position: absolute; inset: 0; }
  .t4-space-stars { position: absolute; inset: 0; opacity: 0; }
  .t4-star { position: absolute; border-radius: 50%; background: #F0EDE6; }
  .t4-heat {
    position: absolute;
    top: 5vh; left: 50%; transform: translateX(-50%);
    width: 120px; height: 80px;
    background: radial-gradient(ellipse at top center, rgba(255,220,100,0.9) 0%, rgba(255,120,30,0.6) 35%, rgba(200,50,0,0.3) 65%, transparent 80%);
    border-radius: 50%;
    opacity: 0;
    filter: blur(4px);
  }
  .t4-craft {
    position: absolute;
    left: 50%; transform: translateX(-50%);
    width: 10px; height: 10px;
    border-radius: 50% 50% 2px 2px;
    background: rgba(240,237,230,0.8);
    box-shadow: 0 0 8px rgba(255,180,80,0.5);
    opacity: 0;
    top: 8vh;
  }
  .t4-trail {
    position: absolute;
    left: 50%; transform: translateX(-50%);
    width: 3px; height: 60px; top: 10vh;
    background: linear-gradient(to bottom, rgba(255,180,50,0.4), rgba(255,80,0,0.2), transparent);
    opacity: 0;
    border-radius: 0 0 2px 2px;
  }
  .t4-atmo {
    position: absolute;
    bottom: 0; left: 0; right: 0; height: 40vh;
    background: linear-gradient(to top, rgba(79,195,247,0.35), rgba(79,195,247,0.08), transparent);
    opacity: 0;
  }
  .t4-city {
    position: absolute;
    bottom: 0; left: 0; right: 0; height: 120px;
    overflow: hidden; opacity: 0;
  }
  .t4-city svg { width: 100%; height: 100%; }

  @media (prefers-reduced-motion: reduce) {
    .t-overlay { transition: opacity 0.2s !important; }
  }
</style>
