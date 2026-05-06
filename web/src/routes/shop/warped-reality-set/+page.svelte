<!-- web/src/routes/shop/warped-reality-set/+page.svelte -->
<!--
  Warped Reality SET drop landing page.
  Beanie + Tee + Sweatpants — sold individually but framed as a 3-piece set.
  Each card auto-cycles through its views/variants:
    · Tee — front ↔ back
    · Beanie — Earth Blue ↔ Green
    · Sweatpants — front only (static)
-->
<script lang="ts">
  import StarField from '$lib/components/StarField.svelte';
  import AutoCycleImage, { type CycleImage } from '$lib/components/AutoCycleImage.svelte';

  interface SetPiece {
    slug: string;
    name: string;
    label: string;       // sub-label under name
    images: CycleImage[]; // 1+ images; 2+ auto-cycle. Per-image scale to balance siblings.
    blurb: string;
  }

  const pieces: SetPiece[] = [
    {
      slug: 'warped-reality-tee',
      name: 'Warped Reality Tee',
      label: 'Front + Back · Black',
      images: ['/photos/_drop/tee-front.png', '/photos/_drop/tee-back.png'],
      blurb: 'Heavyweight oversized cut. Sky-blue collar print up front, ornate spires graphic at the back.',
    },
    {
      slug: 'warped-reality-beanie',
      name: 'Warped Reality Beanie',
      label: 'Earth Blue · Green',
      // Blue PNG has more surrounding whitespace; green is cropped tighter.
      // Per-image scale balances the two so they render the same visual size.
      images: [
        // Blue PNG has the beanie sitting high on the canvas — push it down.
        { src: '/photos/_drop/beanie-blue.png', scale: 1.4, offsetY: 8 },
        { src: '/photos/product-green-beanie.png', scale: 0.75 },
      ],
      blurb: 'Knit for those who drift between dimensions. Two colorways, one frequency.',
    },
    {
      slug: 'warped-reality-sweatpants',
      name: 'Warped Reality Sweatpants',
      label: 'Black',
      images: [{ src: '/photos/_drop/sweatpants-front.png', scale: 1.25 }],
      blurb: 'Wide-leg midweight fleece. Arched IMMORTAL VIBES graphic at the waist.',
    },
  ];
</script>

<svelte:head>
  <title>Warped Reality — Immortal Vibes</title>
  <meta name="description" content="The Warped Reality set drop — beanie, tee, sweatpants. Three pieces. One frequency." />
</svelte:head>

<div class="set-page">
  <StarField />

  <header class="hero">
    <span class="kicker">SET DROP · 001</span>
    <h1 class="title">Warped Reality</h1>
    <p class="subtitle">Three pieces. One frequency.</p>
    <p class="lede">
      A capsule for the drift between dimensions. Each piece stands on its own —
      together they form the set. Sold individually below.
    </p>
  </header>

  <section class="pieces">
    {#each pieces as piece, i}
      <a href="/shop/{piece.slug}" class="piece" style="--i:{i}">
        <div class="piece-frame">
          <AutoCycleImage
            images={piece.images}
            alt={piece.name}
            interval={2600}
            fade={650}
            startOffsetMs={i * 850}
          />
        </div>
        <div class="piece-meta">
          <span class="piece-num">0{i + 1}</span>
          <h3 class="piece-name">{piece.name}</h3>
          <span class="piece-variant">{piece.label}</span>
          <p class="piece-blurb">{piece.blurb}</p>
          <span class="piece-cta">VIEW PIECE →</span>
        </div>
      </a>
    {/each}
  </section>

  <footer class="set-footer">
    <p class="bundle-note">Bundle pricing coming soon — supplier confirms next week.</p>
    <a href="/shop" class="back-link">← BACK TO MISSION SELECT</a>
  </footer>
</div>

<style>
  :global(body) { overflow-y: auto; }

  .set-page {
    position: relative;
    min-height: 100dvh;
    background: #000005;
    color: #F0EDE6;
    padding: 6rem 1.5rem 5rem;
    overflow: hidden;
  }

  /* StarField sits behind everything */
  .set-page :global(.star-field) {
    position: fixed; inset: 0; z-index: 0; pointer-events: none;
  }

  .hero, .pieces, .set-footer {
    position: relative;
    z-index: 2;
    max-width: 1200px;
    margin: 0 auto;
  }

  /* ── Hero ── */
  .hero {
    text-align: center;
    margin-bottom: 5rem;
  }

  .kicker {
    display: inline-block;
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.55em;
    color: rgba(200, 146, 42, 0.8);
    margin-bottom: 1.6rem;
  }

  .title {
    font-family: 'Cormorant Garamond', serif;
    font-weight: 300;
    font-size: clamp(2.6rem, 7vw, 5.4rem);
    letter-spacing: 0.02em;
    line-height: 1.05;
    margin: 0 0 0.8rem;
  }

  .subtitle {
    font-family: 'Cormorant Garamond', serif;
    font-style: italic;
    font-size: clamp(1rem, 1.6vw, 1.35rem);
    color: rgba(240, 237, 230, 0.62);
    margin: 0 0 1.4rem;
  }

  .lede {
    max-width: 38rem;
    margin: 0 auto;
    font-family: 'Inter', sans-serif;
    font-size: 0.92rem;
    line-height: 1.65;
    color: rgba(240, 237, 230, 0.55);
  }

  /* ── Pieces grid ── */
  .pieces {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(280px, 100%), 1fr));
    gap: 2.4rem;
    margin-bottom: 5rem;
  }

  .piece {
    text-decoration: none;
    color: inherit;
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
    transition: transform 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  }

  .piece:hover { transform: translateY(-6px); }

  .piece-frame {
    position: relative;
    aspect-ratio: 1 / 1;
    background: linear-gradient(160deg, rgba(20, 28, 48, 0.6), rgba(8, 10, 22, 0.9));
    border: 1px solid rgba(240, 237, 230, 0.08);
    overflow: hidden;
  }

  .piece-frame::before {
    content: '';
    position: absolute;
    inset: 0;
    background: radial-gradient(circle at 50% 35%, rgba(79,195,247,0.18) 0%, transparent 65%);
    pointer-events: none;
    z-index: 1;
  }

  /* AutoCycleImage's images sit above the radial bg, scale gently on hover */
  .piece-frame :global(.auto-cycle) { z-index: 2; }
  .piece-frame :global(.layer) {
    transition: transform 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  }
  .piece:hover .piece-frame :global(.layer) { transform: scale(1.04); }

  .piece-meta {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .piece-num {
    font-family: 'Inter', sans-serif;
    font-size: 0.5rem;
    letter-spacing: 0.4em;
    color: rgba(200, 146, 42, 0.65);
  }

  .piece-name {
    font-family: 'Cormorant Garamond', serif;
    font-weight: 300;
    font-size: 1.45rem;
    margin: 0.15rem 0 0.15rem;
    letter-spacing: 0.02em;
  }

  .piece-variant {
    font-family: 'Inter', sans-serif;
    font-style: italic;
    font-size: 0.7rem;
    letter-spacing: 0.16em;
    color: rgba(79, 195, 247, 0.85);
    margin-bottom: 0.5rem;
    text-transform: uppercase;
  }

  .piece-blurb {
    font-family: 'Inter', sans-serif;
    font-size: 0.82rem;
    line-height: 1.55;
    color: rgba(240, 237, 230, 0.5);
    margin: 0 0 0.6rem;
  }

  .piece-cta {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.32em;
    color: rgba(240, 237, 230, 0.72);
    transition: color 0.2s, letter-spacing 0.3s;
  }

  .piece:hover .piece-cta {
    color: #F0EDE6;
    letter-spacing: 0.38em;
  }

  /* ── Footer ── */
  .set-footer {
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 1.6rem;
  }

  .bundle-note {
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    letter-spacing: 0.18em;
    color: rgba(240, 237, 230, 0.4);
    margin: 0;
  }

  .back-link {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.32em;
    color: rgba(240, 237, 230, 0.55);
    text-decoration: none;
    border: 1px solid rgba(240, 237, 230, 0.15);
    padding: 0.7rem 1.6rem;
    align-self: center;
    transition: color 0.2s, border-color 0.2s;
  }

  .back-link:hover {
    color: #F0EDE6;
    border-color: rgba(240, 237, 230, 0.5);
  }
</style>
