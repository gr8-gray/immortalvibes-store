<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import gsap from 'gsap';
  import ScrollTrigger from 'gsap/ScrollTrigger';
  import { mouseParallax } from '$lib/stores/mouseParallax';

  let mouseX = 0;
  let mouseY = 0;

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
    { count: 1800, speed: 0.05, minRadius: 0.3, maxRadius: 0.5, minAlpha: 0.04, maxAlpha: 0.18 }, // haze — Milky Way cloud
    { count: 500,  speed: 0.2,  minRadius: 0.4, maxRadius: 0.9, minAlpha: 0.15, maxAlpha: 0.5  }, // deep stars
    { count: 200,  speed: 0.5,  minRadius: 0.6, maxRadius: 1.3, minAlpha: 0.3,  maxAlpha: 0.75 }, // mid stars
    { count: 80,   speed: 0.8,  minRadius: 1.0, maxRadius: 2.2, minAlpha: 0.55, maxAlpha: 1.0  }  // bright stars
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

  // Mouse parallax multipliers per layer (fraction of canvas dimension)
  const MOUSE_MULT = [0.002, 0.008, 0.018, 0.035];

  function draw(): void {
    if (!ctx || !canvas) return;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    LAYERS.forEach((layer, i) => {
      const scrollOffset = scrollY * layer.speed;
      const mox = mouseX * MOUSE_MULT[i] * canvas.width;
      const moy = mouseY * MOUSE_MULT[i] * canvas.height;

      stars[i].forEach((star) => {
        const x = star.x * canvas.width + mox;
        const rawY = star.y * canvas.height - scrollOffset + moy;
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

    scrollTriggerInstance = ScrollTrigger.create({
      start: 0,
      end: 'max',
      onUpdate: (self) => {
        scrollY = self.scroll();
      }
    });

    const unsubMouse = mouseParallax.subscribe(({ x, y }) => {
      mouseX = x;
      mouseY = y;
    });

    rafId = requestAnimationFrame(draw);

    return () => {
      cancelAnimationFrame(rafId);
      window.removeEventListener('resize', resize);
      scrollTriggerInstance?.kill();
      unsubMouse();
    };
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
></canvas>
