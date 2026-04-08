<svelte:head>
  <title>Immortal Vibes</title>
  <meta name="description" content="Apparel forged at the edge of orbit." />
</svelte:head>

<section class="hero">
  <div class="hero__content">
    <h1 class="type-hero hero__title">Immortal Vibes</h1>
    <p class="hero__sub type-label text-lunar-secondary">
      Apparel forged at the edge of orbit
    </p>
    <a href="/shop" class="hero__cta type-label">
      Explore Collection
    </a>
  </div>
</section>

<style>
  .hero {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: calc(100vh - 4.5rem);
    padding: 2rem;
    text-align: center;
  }

  .hero__content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
  }

  .hero__title {
    font-size: clamp(2rem, 8vw, 5rem);
    color: rgba(240, 237, 230, 1);
  }

  .hero__sub {
    font-size: 0.8rem;
    letter-spacing: 0.2em;
  }

  .hero__cta {
    display: inline-block;
    margin-top: 1rem;
    padding: 0.75rem 2.5rem;
    border: 1px solid rgba(240, 237, 230, 0.3);
    color: rgba(240, 237, 230, 0.8);
    text-decoration: none;
    letter-spacing: 0.2em;
    font-size: 0.75rem;
    transition:
      border-color 0.25s ease,
      color 0.25s ease,
      background 0.25s ease;
  }

  .hero__cta:hover {
    border-color: rgba(240, 237, 230, 0.7);
    color: rgba(240, 237, 230, 1);
    background: rgba(240, 237, 230, 0.05);
  }
</style>
