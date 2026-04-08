<script>
  import { page } from '$app/stores';
</script>

<svelte:head>
  <title>{$page.status} — Lost in the void</title>
</svelte:head>

<section class="void-error">
  <div class="void-error__content">
    <p class="void-error__code type-label text-lunar-tertiary">
      {$page.status}
    </p>
    <h1 class="type-hero void-error__title">
      Lost in the void
    </h1>
    <p class="type-body text-lunar-secondary void-error__message">
      {$page.error?.message ?? 'This coordinate does not exist.'}
    </p>
    <a href="/" class="void-error__link type-label">
      Return to orbit
    </a>
  </div>
</section>

<style>
  .void-error {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: calc(100vh - 4.5rem);
    padding: 2rem;
    text-align: center;
  }

  .void-error__content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.25rem;
  }

  .void-error__code {
    font-size: 0.7rem;
    letter-spacing: 0.3em;
  }

  .void-error__title {
    font-size: clamp(2rem, 6vw, 4rem);
    color: rgba(240, 237, 230, 0.9);
  }

  .void-error__message {
    font-size: 0.9rem;
    max-width: 380px;
    line-height: 1.7;
  }

  .void-error__link {
    display: inline-block;
    margin-top: 1.5rem;
    padding: 0.65rem 2rem;
    border: 1px solid rgba(240, 237, 230, 0.25);
    color: rgba(240, 237, 230, 0.7);
    text-decoration: none;
    font-size: 0.7rem;
    letter-spacing: 0.2em;
    transition:
      border-color 0.25s ease,
      color 0.25s ease;
  }

  .void-error__link:hover {
    border-color: rgba(240, 237, 230, 0.6);
    color: rgba(240, 237, 230, 1);
  }
</style>
