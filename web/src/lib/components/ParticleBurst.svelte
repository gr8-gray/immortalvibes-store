<!-- web/src/lib/components/ParticleBurst.svelte -->
<script lang="ts" context="module">
  interface Particle {
    x: number;
    y: number;
    vx: number;
    vy: number;
    alpha: number;
    radius: number;
  }

  const PARTICLE_COUNT = 40;
  const GRAVITY = 0.1;
  const ALPHA_DECAY = 0.02;

  function createParticle(originX: number, originY: number): Particle {
    // Angle: upward cone — -90deg (straight up) ± 60deg → range [-150deg, -30deg]
    const angleDeg = -90 + (Math.random() - 0.5) * 120;
    const angleRad = (angleDeg * Math.PI) / 180;
    const speed = 2 + Math.random() * 4; // 2–6 px/frame
    return {
      x: originX,
      y: originY,
      vx: Math.cos(angleRad) * speed,
      vy: Math.sin(angleRad) * speed,
      alpha: 0.9 + Math.random() * 0.1,
      radius: 1 + Math.random() * 2,
    };
  }

  function updateParticle(p: Particle): Particle {
    return {
      ...p,
      x: p.x + p.vx,
      y: p.y + p.vy,
      vy: p.vy + GRAVITY,
      alpha: p.alpha - ALPHA_DECAY,
    };
  }
</script>

<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D;
  let particles: Particle[] = [];
  let rafId: number;
  let active = false;

  export function trigger(originX: number, originY: number) {
    particles = Array.from({ length: PARTICLE_COUNT }, () => createParticle(originX, originY));
    active = true;
    if (!rafId) animate();
  }

  function animate() {
    if (!ctx || !canvas) return;
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    particles = particles
      .map(updateParticle)
      .filter((p) => p.alpha > 0);

    for (const p of particles) {
      ctx.beginPath();
      ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2);
      // Star-gold tint for the burst
      ctx.fillStyle = `rgba(200, 146, 42, ${p.alpha})`;
      ctx.fill();
    }

    if (particles.length > 0) {
      rafId = requestAnimationFrame(animate);
    } else {
      active = false;
      rafId = 0;
      ctx.clearRect(0, 0, canvas.width, canvas.height);
    }
  }

  function resize() {
    if (!canvas) return;
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
  }

  onMount(() => {
    ctx = canvas.getContext('2d')!;
    resize();
    window.addEventListener('resize', resize);
  });

  onDestroy(() => {
    cancelAnimationFrame(rafId);
    window.removeEventListener('resize', resize);
  });
</script>

<canvas bind:this={canvas} class="particle-canvas" aria-hidden="true"></canvas>

<style>
  .particle-canvas {
    position: fixed;
    inset: 0;
    pointer-events: none;
    z-index: 1000;
  }
</style>
