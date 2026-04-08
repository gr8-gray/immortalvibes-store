<svelte:head>
  <title>Shop — Immortal Vibes</title>
  <meta name="description" content="Browse the Immortal Vibes collection." />
</svelte:head>

<section class="page-section">
  <header class="page-header">
    <h1 class="type-display page-title">Collection</h1>
    <p class="type-body text-lunar-secondary page-subtitle">
      Products load here in Plan 4.
    </p>
  </header>

  <div class="product-grid" aria-label="Product grid placeholder">
    <!-- Product cards populated in Plan 4 -->
  </div>
</section>

<style>
  .page-section {
    max-width: 1280px;
    margin: 0 auto;
    padding: 4rem 2rem;
  }

  .page-header {
    margin-bottom: 3rem;
  }

  .page-title {
    font-size: clamp(1.5rem, 4vw, 3rem);
    color: rgba(240, 237, 230, 1);
    margin-bottom: 0.75rem;
  }

  .page-subtitle {
    font-size: 0.9rem;
  }

  .product-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 2rem;
  }
</style>
