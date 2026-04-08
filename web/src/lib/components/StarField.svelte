<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import gsap from 'gsap';
  import ScrollTrigger from 'gsap/ScrollTrigger';

  if (browser) {
    gsap.registerPlugin(ScrollTrigger);
  }

  // Layer configuration: speed is the parallax scroll multiplier
  interface Layer {
    count: number;
    speed: number;    // 0.2 = slow (far), 0.8 = fast (close)
    minRadius: number;
    maxRadius: number;
    minAlpha: number;
    maxAlpha: number;
  }

  const LAYERS: Layer[] = [
    { count: 200, speed: 0.2, minRadius: 0.5, maxRadius: 0.8, minAlpha: 0.2, maxAlpha: 0.4 },
    { count: 120, speed: 0.5, minRadius: 0.6, maxRadius: 1.1, minAlpha: 0.3, maxAlpha: 0.6 },
    { count: 60,  speed: 0.8, minRadius: 0.9, maxRadius: 1.5, minAlpha: 0.4, maxAlpha: 0.7 }
  ];

  interface Star {
    x: number; // 0..1 normalized
    y: number; // 0..1 normalized
    radius: number;
    alpha: number;
  }

  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D;
  let rafId: number;
  let stars: Star[][] = [];
  let scrollY = 0;
  let scrollTriggerInstance: ScrollTrigger | null = null;

  function rand(min: number, max: number): number {
    return min + Math.random() * (max - min);
  }

  function buildStars(): void {
    stars = LAYERS.map((layer) =>
      Array.from({ length: layer.count }, () => ({
        x: Math.random(),
        y: Math.random(),
        radius: rand(layer.minRadius, layer.maxRadius),
        alpha: rand(layer.minAlpha, layer.maxAlpha)
      }))
    );
  }

  function resize(): void {
    if (!canvas) return;
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
  }

  function draw(): void {
    if (!ctx || !canvas) return;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    LAYERS.forEach((layer, i) => {
      const parallaxOffset = scrollY * layer.speed;

      stars[i].forEach((star) => {
        const x = star.x * canvas.width;
        // Apply parallax: shift Y up by offset, wrap with modulo
        const rawY = star.y * canvas.height - parallaxOffset;
        const y = ((rawY % canvas.height) + canvas.height) % canvas.height;

        ctx.beginPath();
        ctx.arc(x, y, star.radius, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(240, 237, 230, ${star.alpha})`;
        ctx.fill();
      });
    });

    rafId = requestAnimationFrame(draw);
  }

  onMount(() => {
    if (!browser) return;

    ctx = canvas.getContext('2d')!;
    buildStars();
    resize();

    window.addEventListener('resize', resize);

    // GSAP ScrollTrigger drives the parallax scrollY value
    scrollTriggerInstance = ScrollTrigger.create({
      start: 0,
      end: 'max',
      onUpdate: (self) => {
        scrollY = self.scroll();
      }
    });

    rafId = requestAnimationFrame(draw);
  });

  onDestroy(() => {
    if (!browser) return;
    cancelAnimationFrame(rafId);
    window.removeEventListener('resize', resize);
    scrollTriggerInstance?.kill();
  });
</script>

<canvas
  bind:this={canvas}
  aria-hidden="true"
  style="
    position: fixed;
    inset: 0;
    width: 100vw;
    height: 100vh;
    pointer-events: none;
    z-index: 0;
  "
/>
