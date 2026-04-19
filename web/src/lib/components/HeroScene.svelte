<!-- web/src/lib/components/HeroScene.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { page, navigating } from '$app/stores';

  $: showHero = $page.url.pathname === '/' && !$navigating;

  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D;
  let sceneEl: HTMLDivElement;
  let rafId: number;
  let w = 0, h = 0;
  const DPR = browser ? Math.min(window.devicePixelRatio || 1, 2) : 1;

  let targetX = 0, targetY = 0;
  let camX = 0, camY = 0;
  let lastMouseMove = 0;


  // ── Star layers ──────────────────────────────────────────────────────────────
  interface Star {
    x: number; y: number; r: number;
    baseA: number; vy: number; par: number;
    tint: string;
    twinkleSpeed: number; twinklePhase: number;
    layer: number;
  }

  const LAYERS = [
    { count: 220, minR: 0.10, maxR: 0.45, minA: 0.06, maxA: 0.55,
      drift: 0.04, par: 0.008,
      tints: ['rgba(240,237,230,','rgba(240,237,230,','rgba(255,248,225,','rgba(210,220,255,'] },
    { count: 110, minR: 0.28, maxR: 0.90, minA: 0.18, maxA: 0.80,
      drift: 0.09, par: 0.022,
      tints: ['rgba(240,237,230,','rgba(255,248,225,','rgba(210,220,255,','rgba(240,237,230,'] },
    { count: 55,  minR: 1.0,  maxR: 2.6,  minA: 0.55, maxA: 1.0,
      drift: 0.18, par: 0.055,
      tints: ['rgba(255,255,255,','rgba(240,237,230,','rgba(255,248,225,'] },
  ];

  let stars: Star[] = [];
  function rand(a: number, b: number): number { return a + Math.random() * (b - a); }

  function buildStars(): void {
    stars = [];
    let li = 0;
    for (const layer of LAYERS) {
      for (let i = 0; i < layer.count; i++) {
        const tint = layer.tints[Math.floor(Math.random() * layer.tints.length)];
        stars.push({
          x: Math.random(), y: Math.random(),
          r: rand(layer.minR, layer.maxR),
          baseA: rand(layer.minA, layer.maxA),
          vy: -layer.drift * 0.00012,
          par: layer.par, tint,
          twinkleSpeed: rand(0.003, 0.018),
          twinklePhase: rand(0, Math.PI * 2),
          layer: li,
        });
      }
      li++;
    }
  }

  // ── Shooting stars ───────────────────────────────────────────────────────────
  interface Shooter {
    x: number; y: number; vx: number; vy: number;
    alpha: number; decay: number;
  }
  let shooters: Shooter[] = [];
  let nextShooterAt = 0;

  function spawnShooter(): void {
    const sign = Math.random() < 0.5 ? 1 : -1;
    shooters.push({
      x: rand(0.1, 0.9), y: rand(0.02, 0.55),
      vx: rand(0.004, 0.009) * sign,
      vy: rand(0.001, 0.004),
      alpha: 1,
      decay: rand(0.006, 0.014),
    });
  }

  // ── 4-point diffraction spike ────────────────────────────────────────────────
  function drawSpike(px: number, py: number, r: number, a: number): void {
    const len = r * 14;
    for (const [dx, dy] of [[1,0],[0,1]] as [number,number][]) {
      for (const dir of [-1, 1]) {
        const x2 = px + dx * dir * len, y2 = py + dy * dir * len;
        const g = ctx.createLinearGradient(px, py, x2, y2);
        g.addColorStop(0,   `rgba(255,255,255,${(a * 0.55).toFixed(3)})`);
        g.addColorStop(0.4, `rgba(255,248,225,${(a * 0.18).toFixed(3)})`);
        g.addColorStop(1,   'rgba(255,248,225,0)');
        ctx.beginPath();
        ctx.moveTo(px, py);
        ctx.lineTo(x2, y2);
        ctx.strokeStyle = g;
        ctx.lineWidth = r * 0.35;
        ctx.stroke();
      }
    }
  }

  // ── Nebula dust ──────────────────────────────────────────────────────────────
  interface NebulaDust { x:number; y:number; rx:number; ry:number; r:number; g:number; b:number; a:number; phase:number; speed:number; }
  const NEBULA: NebulaDust[] = [
    { x:0.22, y:0.28, rx:0.22, ry:0.14, r:180, g:195, b:255, a:0.028, phase:0,   speed:0.00018 },
    { x:0.68, y:0.42, rx:0.18, ry:0.12, r:255, g:235, b:180, a:0.022, phase:1.8, speed:0.00022 },
    { x:0.50, y:0.18, rx:0.30, ry:0.18, r:200, g:210, b:255, a:0.018, phase:3.2, speed:0.00015 },
    { x:0.80, y:0.65, rx:0.16, ry:0.10, r:255, g:220, b:160, a:0.020, phase:0.9, speed:0.00020 },
    { x:0.12, y:0.72, rx:0.20, ry:0.13, r:160, g:180, b:255, a:0.016, phase:2.5, speed:0.00017 },
  ];

  function drawNebula(cx: number, cy: number): void {
    for (const n of NEBULA) {
      n.phase += n.speed;
      const alpha = n.a * (0.7 + 0.3 * Math.sin(n.phase));
      const ox = (n.x + cx * 0.018) * w, oy = (n.y + cy * 0.012) * h;
      const rx = n.rx * w, ry = n.ry * h;
      const grd = ctx.createRadialGradient(ox, oy, 0, ox, oy, Math.max(rx, ry));
      grd.addColorStop(0,    `rgba(${n.r},${n.g},${n.b},${alpha})`);
      grd.addColorStop(0.45, `rgba(${n.r},${n.g},${n.b},${alpha * 0.4})`);
      grd.addColorStop(1,    `rgba(${n.r},${n.g},${n.b},0)`);
      ctx.save();
      ctx.scale(1, ry / rx);
      ctx.beginPath();
      ctx.arc(ox, oy * (rx / ry), rx, 0, Math.PI * 2);
      ctx.fillStyle = grd;
      ctx.fill();
      ctx.restore();
    }
  }

  // ── Draw loop ────────────────────────────────────────────────────────────────
  function draw(): void {
    if (!ctx) return;
    ctx.clearRect(0, 0, w, h);

    camX += (targetX - camX) * 0.055;
    camY += (targetY - camY) * 0.055;

    const idleMs = Date.now() - lastMouseMove;
    let breathX = 0, breathY = 0;
    if (idleMs > 3000) {
      const ramp = Math.min(1, (idleMs - 3000) / 2000);
      const phase = Date.now() * 0.000085;
      breathX = Math.sin(phase) * 0.055 * ramp;
      breathY = Math.cos(phase * 0.65) * 0.038 * ramp;
    }
    const cx = camX + breathX;
    const cy = camY + breathY;

    // Stars + nebula drawn normally — scene container handles the spin
    drawNebula(cx, cy);

    for (const s of stars) {
      s.y += s.vy;
      if (s.y < -0.02) { s.y = 1.02; s.x = Math.random(); }

      s.twinklePhase += s.twinkleSpeed;
      const twink = 0.55 + 0.45 * Math.sin(s.twinklePhase);
      const a = s.baseA * twink;
      const px = (s.x + cx * s.par) * w;
      const py = (s.y + cy * s.par) * h;

      ctx.beginPath();
      ctx.arc(px, py, s.r, 0, Math.PI * 2);
      ctx.fillStyle = s.tint + a.toFixed(3) + ')';
      ctx.fill();

      if (s.layer === 2) {
        const g1 = ctx.createRadialGradient(px, py, 0, px, py, s.r * 4);
        g1.addColorStop(0, s.tint + (a * 0.55).toFixed(3) + ')');
        g1.addColorStop(1, s.tint + '0)');
        ctx.beginPath();
        ctx.arc(px, py, s.r * 4, 0, Math.PI * 2);
        ctx.fillStyle = g1;
        ctx.fill();

        if (s.r > 1.8) {
          const g2 = ctx.createRadialGradient(px, py, 0, px, py, s.r * 9);
          g2.addColorStop(0, s.tint + (a * 0.14).toFixed(3) + ')');
          g2.addColorStop(1, s.tint + '0)');
          ctx.beginPath();
          ctx.arc(px, py, s.r * 9, 0, Math.PI * 2);
          ctx.fillStyle = g2;
          ctx.fill();
          if (s.r > 2.0) drawSpike(px, py, s.r, a);
        }
      }
    }

    // Vignette drawn on canvas in screen-space (no rotation)
    const vgn = ctx.createRadialGradient(w*0.5, h*0.5, Math.min(w,h)*0.28, w*0.5, h*0.5, Math.max(w,h)*0.85);
    vgn.addColorStop(0, 'rgba(3,3,8,0)');
    vgn.addColorStop(1, 'rgba(3,3,8,0.72)');
    ctx.fillStyle = vgn;
    ctx.fillRect(0, 0, w, h);

    // Shooting stars
    const now = Date.now();
    if (now > nextShooterAt) { spawnShooter(); nextShooterAt = now + rand(2500, 8000); }
    for (let i = shooters.length - 1; i >= 0; i--) {
      const s = shooters[i];
      s.x += s.vx; s.y += s.vy; s.alpha -= s.decay;
      if (s.alpha <= 0) { shooters.splice(i, 1); continue; }
      const x1 = s.x * w, y1 = s.y * h;
      const tailX = s.vx * w * 55, tailY = s.vy * h * 55;
      const g = ctx.createLinearGradient(x1, y1, x1 - tailX, y1 - tailY);
      g.addColorStop(0,    `rgba(255,255,255,${s.alpha})`);
      g.addColorStop(0.25, `rgba(240,237,230,${(s.alpha * 0.5).toFixed(3)})`);
      g.addColorStop(1,    'rgba(240,237,230,0)');
      ctx.beginPath();
      ctx.moveTo(x1, y1);
      ctx.lineTo(x1 - tailX, y1 - tailY);
      ctx.strokeStyle = g; ctx.lineWidth = 1.2; ctx.stroke();
      ctx.beginPath();
      ctx.arc(x1, y1, 1.8, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(255,255,255,${s.alpha})`; ctx.fill();
    }

    // Parallax — scene drifts slowly with mouse, no rotation
    if (sceneEl) {
      sceneEl.style.transform = `translate(${cx * 10}px, ${cy * 7}px)`;
    }

    rafId = requestAnimationFrame(draw);
  }

  function resize(): void {
    if (!canvas) return;
    w = window.innerWidth; h = window.innerHeight;
    canvas.width  = Math.floor(w * DPR);
    canvas.height = Math.floor(h * DPR);
    canvas.style.width  = w + 'px';
    canvas.style.height = h + 'px';
    ctx.setTransform(DPR, 0, 0, DPR, 0, 0);
  }

  onMount(() => {
    if (!browser) return;
    ctx = canvas.getContext('2d')!;
    w = window.innerWidth; h = window.innerHeight;
    canvas.width  = Math.floor(w * DPR);
    canvas.height = Math.floor(h * DPR);
    canvas.style.width  = w + 'px';
    canvas.style.height = h + 'px';
    ctx.setTransform(DPR, 0, 0, DPR, 0, 0);

    buildStars();
    lastMouseMove = Date.now();
    nextShooterAt = Date.now() + rand(2000, 5000);
    window.addEventListener('resize', resize);

    function onMouseMove(e: MouseEvent) {
      targetX = (e.clientX / w - 0.5) * 2;
      targetY = (e.clientY / h - 0.5) * 2;
      lastMouseMove = Date.now();
    }
    function onMouseLeave() { targetX = 0; targetY = 0; }
    function onTouchMove(e: TouchEvent) {
      if (!e.touches[0]) return;
      targetX = (e.touches[0].clientX / w - 0.5) * 2;
      targetY = (e.touches[0].clientY / h - 0.5) * 2;
      lastMouseMove = Date.now();
    }

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseleave', onMouseLeave);
    window.addEventListener('touchmove', onTouchMove, { passive: true });
    rafId = requestAnimationFrame(draw);

    return () => {
      cancelAnimationFrame(rafId);
      window.removeEventListener('resize', resize);
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseleave', onMouseLeave);
      window.removeEventListener('touchmove', onTouchMove);
    };
  });
</script>

<!-- Scene wrapper — MW photo + canvas rotate together as one unit -->
<div bind:this={sceneEl} class="scene" aria-hidden="true">
  <div class="mw-photo"></div>
  <canvas bind:this={canvas} class="hero-canvas"></canvas>
</div>

{#if showHero}
<!-- UI overlay — stays fixed, never spins -->
<div class="hero">
  <h1 class="hero__headline">Rise Beyond<br>the Mortal Plane</h1>
  <a href="/shop" class="hero__cta">Enter the Missions →</a>
</div>

<p class="stamp" aria-hidden="true">Garments Built for Those Who Orbit Higher</p>
{/if}

<style>
  /* Scene container — both layers spin together */
  .scene {
    position: fixed;
    inset: 0;
    z-index: 0;
    transform-origin: center center;
    will-change: transform;
  }

  .mw-photo {
    position: absolute; inset: 0;
    background-image: url('/milky-way.webp');
    background-size: cover;
    background-position: center 0%;
    mix-blend-mode: screen;
    opacity: 0;
    animation: mwFadeIn 4s ease forwards;
  }
  @keyframes mwFadeIn { from{opacity:0} to{opacity:0.86} }

  .hero-canvas {
    position: absolute; inset: 0;
    pointer-events: none;
    animation: fadeIn 2.5s ease 0.4s both;
  }

  /* UI — locked to screen, never rotates */
  .hero {
    position: fixed;
    top: 42%; left: 50%;
    transform: translate(-50%, -50%);
    z-index: 5;
    display: flex; flex-direction: column;
    align-items: center; gap: 2.5rem;
    pointer-events: none;
  }

  .hero__headline {
    font-family: 'Cormorant Garamond', serif;
    font-weight: 200;
    font-size: clamp(2.25rem, 6.2vw, 5.25rem);
    line-height: 1.04; letter-spacing: 0.12em;
    text-align: center; text-transform: uppercase;
    color: rgba(240,237,230,0.96);
    text-shadow: 0 0 80px rgba(240,237,230,0.14), 0 0 160px rgba(240,237,230,0.05);
    max-width: 22ch;
    opacity: 0;
    animation: fadeUp 2.8s cubic-bezier(0.16,1,0.3,1) 0.8s forwards;
  }

  .hero__cta {
    display: inline-block;
    border: 1px solid rgba(240,237,230,0.3);
    border-bottom-color: rgba(200,146,42,0.5);
    color: rgba(240,237,230,0.9);
    font-family: 'Inter', sans-serif;
    font-size: 0.62rem; font-weight: 400;
    letter-spacing: 0.28em; padding: 1.05rem 2.6rem;
    text-decoration: none;
    background: rgba(0,0,0,0.35);
    backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);
    box-shadow: 0 0 30px rgba(240,237,230,0.08), 0 0 60px rgba(200,146,42,0.06);
    animation: ctaPulse 2.8s ease-in-out infinite, fadeIn 2s ease 1.4s backwards;
    pointer-events: all; cursor: none;
    transition: border-color 0.25s, box-shadow 0.25s, color 0.25s;
    text-transform: uppercase; white-space: nowrap;
  }
  .hero__cta:hover {
    border-color: rgba(240,237,230,0.8);
    border-bottom-color: rgba(200,146,42,1);
    color: rgba(240,237,230,1);
    box-shadow: 0 0 50px rgba(240,237,230,0.18), 0 0 100px rgba(200,146,42,0.14);
  }

  .stamp {
    position: fixed; bottom: 1.5rem; left: 50%;
    transform: translateX(-50%); z-index: 5;
    font-family: 'Inter', sans-serif;
    font-size: 0.50rem; letter-spacing: 0.32em;
    color: rgba(240,237,230,0.18); text-transform: uppercase;
    pointer-events: none; white-space: nowrap;
    opacity: 0; animation: fadeIn 2s ease 2s forwards;
  }

  @keyframes ctaPulse {
    0%,100% { border-bottom-color: rgba(200,146,42,0.18); }
    50%      { border-bottom-color: rgba(200,146,42,0.72); }
  }
  @keyframes fadeIn { from{opacity:0} to{opacity:1} }
  @keyframes fadeUp {
    from { opacity:0; transform:translateY(16px); }
    to   { opacity:1; transform:translateY(0); }
  }

  @media (prefers-reduced-motion: reduce) {
    .mw-photo { animation: none; opacity: 0.86; }
    .hero-canvas { animation: none; }
    .hero__headline, .hero__cta, .stamp { animation: none; opacity: 1; transform: none; }
  }

  @media (max-width: 640px) {
    .hero__headline { letter-spacing: 0.08em; }
    .hero { gap: 1.75rem; top: 50%; }
  }
</style>
