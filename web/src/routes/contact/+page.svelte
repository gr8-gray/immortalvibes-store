<!-- web/src/routes/contact/+page.svelte -->
<svelte:head>
  <title>Contact — Immortal Vibes</title>
  <meta name="description" content="Get in touch with Immortal Vibes." />
</svelte:head>

<div class="contact">
  <div class="contact-inner">

    <p class="eyebrow">Contact</p>
    <h1 class="headline">Get in Touch</h1>
    <p class="sub">
      Questions about orders, wholesale inquiries, or just want to connect —
      reach out directly.
    </p>

    <div class="channels">
      <!-- Email placeholder — owner to supply -->
      <div class="channel">
        <span class="channel-label">Email</span>
        <a href="mailto:immortalvibesco@gmail.com" class="channel-value">immortalvibesco@gmail.com</a>
      </div>

      <!-- Social placeholder — owner to supply -->
      <div class="channel">
        <span class="channel-label">Instagram</span>
        <span class="channel-value coming">Coming soon</span>
      </div>
    </div>

    <div class="divider"></div>

    <p class="note">
      For order issues or shipping questions, please include your order number.
      We respond within 24–48 hours.
    </p>

  </div>
</div>

<style>
  .contact {
    min-height: 80vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
  }

  .contact-inner {
    max-width: 520px;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 2rem;
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
    font-size: clamp(2rem, 5vw, 3.5rem);
    font-weight: 300;
    color: #F0EDE6;
    margin: 0;
    letter-spacing: 0.03em;
    line-height: 1.05;
  }

  .sub {
    font-family: 'Inter', sans-serif;
    font-size: 0.875rem;
    line-height: 1.75;
    color: rgba(240,237,230,0.5);
    margin: 0;
  }

  .channels {
    display: flex;
    flex-direction: column;
    gap: 0;
    border: 1px solid rgba(240,237,230,0.08);
  }

  .channel {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.2rem 1.5rem;
    border-bottom: 1px solid rgba(240,237,230,0.06);
  }

  .channel:last-child { border-bottom: none; }

  .channel-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.28em;
    color: rgba(240,237,230,0.35);
    text-transform: uppercase;
  }

  .channel-value {
    font-family: 'Inter', sans-serif;
    font-size: 0.8rem;
    color: rgba(240,237,230,0.85);
    letter-spacing: 0.02em;
    text-decoration: none;
    transition: color 0.2s ease;
  }

  .channel-value.coming {
    font-size: 0.55rem;
    letter-spacing: 0.2em;
    color: rgba(240,237,230,0.2);
    text-transform: uppercase;
  }

  .divider {
    height: 1px;
    background: rgba(240,237,230,0.06);
  }

  .note {
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    line-height: 1.7;
    color: rgba(240,237,230,0.3);
    margin: 0;
  }
</style>
