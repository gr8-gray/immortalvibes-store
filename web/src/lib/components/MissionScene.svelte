<!-- web/src/lib/components/MissionScene.svelte -->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  export let missionNumber: '001' | '002' | '003' | '004' = '001';

  // --- Mission environment configs ---
  const missions = {
    '001': {
      // Low Earth Orbit
      gradient: 'linear-gradient(to top, #060a14, #0a1020)',
      glowColor: 'rgba(30,111,217,0.3)',
      glowPosition: 'bottom',
      starCount: 220,
      starColorRange: [180, 220] as [number, number], // blue-white
    },
    '002': {
      // Lunar Surface
      gradient: 'linear-gradient(to top, #1a1814, #030308)',
      glowColor: 'rgba(240,237,230,0.04)',
      glowPosition: 'bottom',
      starCount: 280,
      starColorRange: [220, 255] as [number, number], // white-ish
    },
    '003': {
      // Stellar Nursery
      gradient: 'linear-gradient(to top, #3d1200, #030308)',
      glowColor: 'rgba(255,100,30,0.18)',
      glowPosition: 'center',
      starCount: 260,
      starColorRange: [200, 255] as [number, number], // warm white
    },
    '004': {
      // Deep Space
      gradient: 'linear-gradient(to top, #2a0d00, #030308)',
      glowColor: 'rgba(180,60,20,0.12)',
      glowPosition: 'center',
      starCount: 200,
      starColorRange: [200, 240] as [number, number],
    },
  } as const;

  $: config = missions[missionNumber] ?? missions['001'];

  let canvas: HTMLCanvasElement;
  let animationId: number;

  interface Star {
    x: number;
    y: number;
    radius: number;
    alpha: number;
    twinkleSpeed: number;
    twinklePhase: number;
  }

  let stars: Star[] = [];

  function initStars(width: number, height: number): Star[] {
    const out: Star[] = [];
    for (let i = 0; i < config.starCount; i++) {
      out.push({
        x: Math.random() * width,
        y: Math.random() * height,
        radius: Math.random() * 1.2 + 0.2,
        alpha: Math.random() * 0.6 + 0.3,
        twinkleSpeed: Math.random() * 0.015 + 0.005,
        twinklePhase: Math.random() * Math.PI * 2,
      });
    }
    return out;
  }

  function drawStars(ctx: CanvasRenderingContext2D, t: number) {
    ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
    for (const star of stars) {
      const alpha = star.alpha * (0.7 + 0.3 * Math.sin(t * star.twinkleSpeed + star.twinklePhase));
      ctx.beginPath();
      ctx.arc(star.x, star.y, star.radius, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(240,237,230,${alpha})`;
      ctx.fill();
    }
  }

  function startLoop(ctx: CanvasRenderingContext2D) {
    let t = 0;
    function frame() {
      t += 1;
      drawStars(ctx, t);
      animationId = requestAnimationFrame(frame);
    }
    animationId = requestAnimationFrame(frame);
  }

  function resize() {
    if (!canvas) return;
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    stars = initStars(canvas.width, canvas.height);
  }

  onMount(() => {
    const ctx = canvas.getContext('2d')!;
    resize();
    window.addEventListener('resize', resize);
    startLoop(ctx);
  });

  onDestroy(() => {
    cancelAnimationFrame(animationId);
    window.removeEventListener('resize', resize);
  });
</script>

<div
  class="mission-scene"
  style="background: {config.gradient};"
>
  <!-- Ambient glow layer -->
  {#if config.glowPosition === 'bottom'}
    <div
      class="glow glow--bottom"
      style="background: radial-gradient(ellipse 80% 40% at 50% 100%, {config.glowColor}, transparent);"
    ></div>
  {:else}
    <div
      class="glow glow--center"
      style="background: radial-gradient(ellipse 60% 50% at 50% 60%, {config.glowColor}, transparent);"
    ></div>
  {/if}

  <!-- Star canvas -->
  <canvas bind:this={canvas} class="star-canvas" aria-hidden="true"></canvas>

  <!-- Page content goes here -->
  <slot />
</div>

<style>
  .mission-scene {
    position: relative;
    min-height: 100vh;
    width: 100%;
    overflow: hidden;
  }

  .glow {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 1;
  }

  .star-canvas {
    position: absolute;
    inset: 0;
    z-index: 2;
    pointer-events: none;
  }

  /* Content placed in the slot sits above canvas */
  :global(.mission-scene > *:not(.glow):not(.star-canvas)) {
    position: relative;
    z-index: 3;
  }
</style>
