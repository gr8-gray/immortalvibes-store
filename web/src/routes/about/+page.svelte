<!-- web/src/routes/about/+page.svelte -->
<svelte:head>
  <title>About — Immortal Vibes</title>
  <meta name="description" content="This Reality is Yours to Master. The story behind Immortal Vibes." />
</svelte:head>

<div class="about">
  <div class="about-layout">

    <!-- Sticky photo — left column -->
    <div class="photo-col">
      <img
        src="/photos/brand/owner-bape-store.jpg"
        alt="Immortal Vibes"
        class="owner-photo"
      />
      <div class="photo-caption">Rise Beyond the Mortal Plane</div>
    </div>

    <!-- Scrollable copy — right column -->
    <div class="content-col">

      <p class="eyebrow">About</p>
      <h1 class="headline">This Reality is<br />Yours to Master</h1>

      <p class="lead">
        Most brands sell you a look, we provide the armor for your evolution.
        Immortal Vibes was born from a singular defiance: the refusal to let
        the world dictate your limits. We aren't here to fit into the basic
        streetwear mold. We're here to break the cycle.
      </p>

      <div class="divider"></div>

      <p class="body-text">
        The spark for this brand started in Florida, born from the heat of
        fighting against the odds and pushing through the noise of those trying
        to disrupt my energy. In 2025, that vision took physical form in
        Okinawa, Japan. From the streets of the South to the islands of the
        Pacific, the message remained the same: <em>prosper past the interference.</em>
      </p>

      <p class="body-text">
        At Immortal Vibes, we treat every drop as a distinct frequency.
      </p>

      <ul class="tenets">
        <li><span class="dot">·</span> Every colorway is its own entity.</li>
        <li><span class="dot">·</span> Every design is its own energy.</li>
        <li><span class="dot">·</span> Every piece is a path toward elevation.</li>
      </ul>

      <p class="body-text">
        Our artwork isn't just decoration — it's a visual representation of the
        breakthrough. It's the moment the old self shatters and the higher self
        takes the lead. This is the gear for the version of you that can't be
        broken, slowed, or drained.
      </p>

      <p class="body-text">
        When you wear Immortal Vibes, you aren't just wearing a brand; you are
        claiming your sovereignty and protecting your peace.
      </p>

      <p class="signoff">Rise Above the Mortals.</p>

      <img src="/logo-full.png" alt="Immortal Vibes" class="logo-mark" />

    </div>
  </div>
</div>

<style>
  .about {
    min-height: 100vh;
    color: #F0EDE6;
  }

  /* ── Two-column layout ── */
  .about-layout {
    display: grid;
    grid-template-columns: 1fr 1fr;
    min-height: calc(100vh - 4.5rem);
    align-items: start;
  }

  @media (max-width: 768px) {
    .about-layout {
      grid-template-columns: 1fr;
      min-height: unset;
    }
    .photo-col {
      position: relative;
      top: 0;
      height: 52vw;
      max-height: 280px;
    }
    .content-col {
      border-left: none;
      border-top: 1px solid rgba(240,237,230,0.06);
      padding: 2.5rem 1.5rem 4rem;
    }
    .headline { font-size: clamp(1.6rem, 7vw, 2.4rem); }
    .lead, .body-text { font-size: 0.875rem; }
  }

  /* ── Sticky photo ── */
  .photo-col {
    position: sticky;
    top: 4.5rem;
    height: calc(100vh - 4.5rem);
    overflow: hidden;
  }

  .owner-photo {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center 20%;
    display: block;
    filter: brightness(0.85);
  }

  .photo-caption {
    position: absolute;
    bottom: 1.5rem;
    left: 1.5rem;
    font-family: 'Inter', sans-serif;
    font-size: 0.50rem;
    letter-spacing: 0.25em;
    color: rgba(240,237,230,0.35);
    text-transform: uppercase;
  }

  /* ── Scrollable copy ── */
  .content-col {
    padding: 5rem 4.5rem 6rem;
    display: flex;
    flex-direction: column;
    gap: 2rem;
    border-left: 1px solid rgba(240,237,230,0.06);
  }

  @media (max-width: 1024px) {
    .content-col { padding: 4rem 2.5rem 5rem; }
  }

  .eyebrow {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.45em;
    color: rgba(200,146,42,0.7);
    text-transform: uppercase;
    margin: 0;
  }

  .headline {
    font-family: 'Cormorant Garamond', serif;
    font-size: clamp(2rem, 3.5vw, 3.5rem);
    font-weight: 300;
    line-height: 1.08;
    margin: 0;
    color: #F0EDE6;
    letter-spacing: 0.02em;
  }

  .lead {
    font-family: 'Inter', sans-serif;
    font-size: 0.9375rem;
    line-height: 1.8;
    color: rgba(240,237,230,0.65);
    margin: 0;
  }

  .divider {
    height: 1px;
    background: rgba(240,237,230,0.08);
  }

  .body-text {
    font-family: 'Inter', sans-serif;
    font-size: 0.9375rem;
    line-height: 1.85;
    color: rgba(240,237,230,0.6);
    margin: 0;
  }

  .body-text em {
    font-style: italic;
    color: rgba(240,237,230,0.9);
  }

  .tenets {
    list-style: none;
    padding: 0 0 0 1rem;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    border-left: 1px solid rgba(200,146,42,0.25);
  }

  .tenets li {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.1rem;
    font-weight: 300;
    color: rgba(240,237,230,0.85);
    letter-spacing: 0.03em;
  }

  .dot {
    color: rgba(200,146,42,0.7);
    margin-right: 0.75rem;
  }

  .signoff {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.4rem;
    font-weight: 300;
    letter-spacing: 0.12em;
    color: rgba(240,237,230,0.9);
    margin: 0;
    padding-top: 1rem;
    border-top: 1px solid rgba(240,237,230,0.08);
  }

  .logo-mark {
    width: clamp(80px, 10vw, 120px);
    opacity: 0.45;
    margin-top: 1rem;
  }
</style>
