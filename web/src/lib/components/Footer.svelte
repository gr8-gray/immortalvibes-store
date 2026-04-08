<footer class="footer">
  <div class="footer__inner">
    <p class="footer__copy">
      &copy; {new Date().getFullYear()} Immortal Vibes
    </p>
    <nav class="footer__links" aria-label="Footer navigation">
      <a href="/shop" class="footer__link">Shop</a>
      <a href="/about" class="footer__link">About</a>
      <a href="/contact" class="footer__link">Contact</a>
    </nav>
  </div>
</footer>

<style>
  .footer {
    position: relative;
    z-index: 10;
    border-top: 1px solid rgba(240, 237, 230, 0.08);
    padding: 2rem 0;
    margin-top: 6rem;
  }

  .footer__inner {
    max-width: 1280px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 1rem;
  }

  .footer__copy {
    font-family: 'Inter', sans-serif;
    font-weight: 300;
    font-size: 0.75rem;
    letter-spacing: 0.08em;
    color: rgba(240, 237, 230, 0.4);
  }

  .footer__links {
    display: flex;
    gap: 1.5rem;
  }

  .footer__link {
    font-family: 'Inter', sans-serif;
    font-weight: 300;
    font-size: 0.75rem;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    text-decoration: none;
    color: rgba(240, 237, 230, 0.4);
    transition: color 0.2s ease;
  }

  .footer__link:hover {
    color: rgba(240, 237, 230, 0.8);
  }
</style>
