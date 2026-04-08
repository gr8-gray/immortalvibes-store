<!-- web/src/routes/shop/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  let canvas: HTMLCanvasElement;

  onMount(() => {
    if (!browser || !canvas) return;
    const ctx = canvas.getContext('2d')!;
    let rafId: number;
    let w = window.innerWidth;
    let h = window.innerHeight;
    canvas.width = w;
    canvas.height = h;

    // Stars
    const stars = Array.from({ length: 800 }, () => ({
      x: Math.random() * w,
      y: Math.random() * h,
      r: Math.random() < 0.04 ? 1.6 + Math.random() * 0.8 :
         Math.random() < 0.15 ? 0.8 + Math.random() * 0.6 : 0.3 + Math.random() * 0.4,
      a: 0.1 + Math.random() * 0.85,
      twinkleSpeed: 0.004 + Math.random() * 0.012,
      twinklePhase: Math.random() * Math.PI * 2,
    }));

    function draw() {
      ctx.clearRect(0, 0, w, h);

      // Nebula background
      const bg = ctx.createRadialGradient(w*0.5, h*0.45, 0, w*0.5, h*0.45, w*0.7);
      bg.addColorStop(0,   'rgba(35,18,8,1)');
      bg.addColorStop(0.3, 'rgba(22,10,18,1)');
      bg.addColorStop(0.7, 'rgba(8,5,14,1)');
      bg.addColorStop(1,   'rgba(3,2,8,1)');
      ctx.fillStyle = bg;
      ctx.fillRect(0, 0, w, h);

      // Warm nebula cloud — center-left
      const n1 = ctx.createRadialGradient(w*0.3, h*0.4, 0, w*0.3, h*0.4, w*0.45);
      n1.addColorStop(0,   'rgba(120,55,15,0.22)');
      n1.addColorStop(0.4, 'rgba(80,30,8,0.12)');
      n1.addColorStop(1,   'rgba(0,0,0,0)');
      ctx.fillStyle = n1; ctx.fillRect(0,0,w,h);

      // Purple nebula — upper right
      const n2 = ctx.createRadialGradient(w*0.75, h*0.25, 0, w*0.75, h*0.25, w*0.38);
      n2.addColorStop(0,   'rgba(70,20,80,0.18)');
      n2.addColorStop(0.5, 'rgba(40,10,55,0.08)');
      n2.addColorStop(1,   'rgba(0,0,0,0)');
      ctx.fillStyle = n2; ctx.fillRect(0,0,w,h);

      // Amber nebula — bottom center
      const n3 = ctx.createRadialGradient(w*0.5, h*0.8, 0, w*0.5, h*0.8, w*0.5);
      n3.addColorStop(0,   'rgba(100,45,10,0.15)');
      n3.addColorStop(0.5, 'rgba(60,20,5,0.07)');
      n3.addColorStop(1,   'rgba(0,0,0,0)');
      ctx.fillStyle = n3; ctx.fillRect(0,0,w,h);

      // Stars
      stars.forEach(s => {
        s.twinklePhase += s.twinkleSpeed;
        const a = s.a * (0.65 + 0.35 * Math.sin(s.twinklePhase));
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(230,225,245,${a})`;
        ctx.fill();
      });

      // Thin orbital arc lines (SVG-style, drawn in canvas)
      ctx.strokeStyle = 'rgba(180,140,80,0.08)';
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.ellipse(w*0.5, h*0.52, w*0.38, h*0.30, -0.15, 0, Math.PI * 2);
      ctx.stroke();

      rafId = requestAnimationFrame(draw);
    }

    draw();

    const onResize = () => {
      w = window.innerWidth; h = window.innerHeight;
      canvas.width = w; canvas.height = h;
    };
    window.addEventListener('resize', onResize);
    return () => { cancelAnimationFrame(rafId); window.removeEventListener('resize', onResize); };
  });
</script>

<svelte:head>
  <title>Immortal Vibes — Select Your Mission</title>
</svelte:head>

<!-- Nebula canvas background -->
<canvas
  bind:this={canvas}
  aria-hidden="true"
  style="position:fixed;inset:0;width:100vw;height:100vh;pointer-events:none;z-index:2;"
></canvas>

<!-- Space UI layer -->
<div class="space-ui">

  <!-- Header label -->
  <div class="destinations-label">MISSION SELECT</div>

  <!-- Planet grid -->
  <div class="planets">

    <!-- Mission 001 — Beanie — Blue-white LEO planet -->
    <div class="mission-slot slot-001">
      <a href="/shop/warped-reality-beanie" class="planet-link">
        <div class="planet planet-leo">
          <div class="planet-glow planet-glow-leo"></div>
        </div>
        <img src="/photos/blue-beanie.jpeg" alt="Warped Reality Beanie" class="product-float" />
        <div class="mission-label">
          <span class="mission-num">001</span>
          <span class="mission-name">Warped Reality Beanie</span>
          <span class="mission-loc">Low Earth Orbit</span>
        </div>
      </a>
    </div>

    <!-- Mission 002 — Hat — Lunar grey planet -->
    <div class="mission-slot slot-002">
      <a href="/shop/vanguard-trucker-hat" class="planet-link">
        <div class="planet planet-lunar">
          <div class="planet-glow planet-glow-lunar"></div>
        </div>
        <div class="product-float product-tbd">NO PHOTO YET</div>
        <div class="mission-label">
          <span class="mission-num">002</span>
          <span class="mission-name">Vanguard Trucker Hat</span>
          <span class="mission-loc">Lunar Surface</span>
        </div>
      </a>
    </div>

    <!-- Mission 003 — Tank — Warm orange nebula planet -->
    <div class="mission-slot slot-003">
      <a href="/shop/racerback-tanktop" class="planet-link">
        <div class="planet planet-nursery">
          <div class="planet-glow planet-glow-nursery"></div>
        </div>
        <img src="/photos/tank-front.png" alt="Racerback Tanktop" class="product-float" />
        <div class="mission-label">
          <span class="mission-num">003</span>
          <span class="mission-name">Racerback Tanktop</span>
          <span class="mission-loc">Stellar Nursery</span>
        </div>
      </a>
    </div>

  </div>

  <!-- Earth at bottom — where you came from -->
  <div class="earth-anchor">
    <div class="earth-orb"></div>
    <span class="earth-label">EARTH</span>
    <a href="/" class="return-btn">↓ RETURN TO EARTH</a>
  </div>

</div>

<style>
  :global(body) { overflow: hidden; }

  .space-ui {
    position: relative;
    z-index: 10;
    width: 100vw;
    height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: space-between;
    padding: 3rem 2rem 2.5rem;
    pointer-events: none;
  }

  .destinations-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.5em;
    color: rgba(200, 146, 42, 0.7);
    text-transform: uppercase;
  }

  /* ── Planet grid ── */
  .planets {
    display: flex;
    align-items: flex-end;
    justify-content: center;
    gap: clamp(3rem, 8vw, 7rem);
    flex: 1;
    padding: 2rem 0;
  }

  .mission-slot {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.2rem;
    pointer-events: all;
  }

  /* Stagger heights for Destiny arc feel */
  .slot-001 { transform: translateY(-2vh); }
  .slot-002 { transform: translateY(3vh); }
  .slot-003 { transform: translateY(-1vh); }

  .planet-link {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    text-decoration: none;
    cursor: pointer;
    transition: transform 0.4s ease;
  }

  .planet-link:hover { transform: scale(1.06) translateY(-6px); }
  .planet-link:hover .planet { box-shadow: var(--planet-hover-shadow); }

  /* ── Planet orbs ── */
  .planet {
    width: clamp(120px, 14vw, 175px);
    height: clamp(120px, 14vw, 175px);
    border-radius: 50%;
    position: relative;
    transition: box-shadow 0.4s ease;
  }

  .planet-leo {
    background: radial-gradient(circle at 38% 32%,
      #d8eeff 0%, #7ab8ec 18%, #3a80cc 38%, #1a52a8 58%, #0c2860 78%, #050f30 95%
    );
    box-shadow: 0 0 50px rgba(79,195,247,0.22), inset -22px -18px 45px rgba(0,0,20,0.55);
    --planet-hover-shadow: 0 0 80px rgba(79,195,247,0.45), inset -22px -18px 45px rgba(0,0,20,0.55);
  }

  .planet-lunar {
    background: radial-gradient(circle at 40% 35%,
      #e8e4de 0%, #b8b0a4 20%, #8a8278 40%, #5e5850 60%, #3a352e 80%, #1a1610 95%
    );
    box-shadow: 0 0 40px rgba(200,190,180,0.15), inset -20px -16px 40px rgba(0,0,0,0.6);
    --planet-hover-shadow: 0 0 65px rgba(200,190,180,0.35), inset -20px -16px 40px rgba(0,0,0,0.6);
  }

  .planet-nursery {
    background: radial-gradient(circle at 36% 30%,
      #ffd0a0 0%, #e8904a 18%, #c05a18 38%, #882808 58%, #4a1004 78%, #1c0402 95%
    );
    box-shadow: 0 0 50px rgba(220,120,40,0.22), inset -22px -18px 45px rgba(20,5,0,0.55);
    --planet-hover-shadow: 0 0 80px rgba(220,120,40,0.45), inset -22px -18px 45px rgba(20,5,0,0.55);
  }

  /* Atmosphere rim glow on each planet */
  .planet-glow {
    position: absolute;
    inset: -3px;
    border-radius: 50%;
    opacity: 0.5;
  }
  .planet-glow-leo     { box-shadow: inset 0 0 20px rgba(79,195,247,0.4); }
  .planet-glow-lunar   { box-shadow: inset 0 0 20px rgba(180,170,160,0.3); }
  .planet-glow-nursery { box-shadow: inset 0 0 20px rgba(255,150,60,0.4); }

  /* ── Product photos ── */
  .product-float {
    width: clamp(48px, 6vw, 72px);
    height: clamp(48px, 6vw, 72px);
    object-fit: contain;
    filter: drop-shadow(0 4px 16px rgba(0,0,0,0.7));
    transform: rotate(-5deg) translateY(-8px);
    transition: transform 0.3s ease;
  }

  .planet-link:hover .product-float { transform: rotate(-2deg) translateY(-14px) scale(1.1); }

  .product-tbd {
    display: flex;
    align-items: center;
    justify-content: center;
    width: clamp(48px, 6vw, 72px);
    height: clamp(48px, 6vw, 72px);
    font-family: 'Inter', sans-serif;
    font-size: 0.45rem;
    letter-spacing: 0.15em;
    color: rgba(240,237,230,0.2);
    text-align: center;
  }

  /* ── Mission label ── */
  .mission-label {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }

  .mission-num {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.3em;
    color: rgba(200,146,42,0.6);
  }

  .mission-name {
    font-family: 'Cormorant Garamond', serif;
    font-size: clamp(0.9rem, 1.5vw, 1.2rem);
    font-weight: 300;
    color: #F0EDE6;
    letter-spacing: 0.05em;
  }

  .mission-loc {
    font-family: 'Inter', sans-serif;
    font-size: 0.5rem;
    letter-spacing: 0.2em;
    color: rgba(240,237,230,0.3);
    text-transform: uppercase;
  }

  /* ── Earth at bottom ── */
  .earth-anchor {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.6rem;
    pointer-events: all;
  }

  .earth-orb {
    width: clamp(60px, 7vw, 90px);
    height: clamp(60px, 7vw, 90px);
    border-radius: 50%;
    background: radial-gradient(circle at 38% 32%,
      #a0d4f0 0%, #4494cc 18%, #2266aa 35%, #1a5e3a 48%,
      #124e2e 58%, #1a4a8c 68%, #0d2d5e 82%, #050f22 95%
    );
    box-shadow: 0 0 30px rgba(79,195,247,0.2), inset -10px -8px 25px rgba(0,0,20,0.5);
  }

  .earth-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.5rem;
    letter-spacing: 0.35em;
    color: rgba(240,237,230,0.3);
  }

  .return-btn {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.22em;
    color: rgba(240,237,230,0.45);
    text-decoration: none;
    border: 1px solid rgba(240,237,230,0.15);
    padding: 0.6rem 1.6rem;
    transition: color 0.2s, border-color 0.2s;
    background: rgba(0,0,0,0.3);
    backdrop-filter: blur(4px);
  }

  .return-btn:hover {
    color: #F0EDE6;
    border-color: rgba(240,237,230,0.45);
  }
</style>
