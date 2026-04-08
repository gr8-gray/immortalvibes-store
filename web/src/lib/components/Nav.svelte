<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import { cartCount } from '$lib/stores/cart';
  import { page } from '$app/stores';

  let scrolled = false;
  let scrollHandler: (() => void) | null = null;

  onMount(() => {
    if (!browser) return;

    scrollHandler = () => {
      scrolled = window.scrollY > 10;
    };
    window.addEventListener('scroll', scrollHandler, { passive: true });
  });

  onDestroy(() => {
    if (browser && scrollHandler) {
      window.removeEventListener('scroll', scrollHandler);
    }
  });

  function isActive(href: string): boolean {
    return $page.url.pathname === href || $page.url.pathname.startsWith(href + '/');
  }
</script>

<nav
  class="nav"
  class:nav--scrolled={scrolled}
  aria-label="Main navigation"
>
  <div class="nav__inner">
    <!-- Logo -->
    <a href="/" class="nav__logo" aria-label="Immortal Vibes home">
      <span class="nav__logo-symbol" aria-hidden="true">⌥</span>
      <span class="nav__logo-text">Immortal Vibes</span>
    </a>

    <!-- Links -->
    <ul class="nav__links" role="list">
      <li>
        <a
          href="/shop"
          class="nav__link"
          class:nav__link--active={isActive('/shop')}
        >Shop</a>
      </li>
      <li>
        <a
          href="/about"
          class="nav__link"
          class:nav__link--active={isActive('/about')}
        >About</a>
      </li>
      <li>
        <a
          href="/contact"
          class="nav__link"
          class:nav__link--active={isActive('/contact')}
        >Contact</a>
      </li>
    </ul>

    <!-- Cart -->
    <a href="/cart" class="nav__cart" aria-label="Cart ({$cartCount} items)">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z" />
        <line x1="3" y1="6" x2="21" y2="6" />
        <path d="M16 10a4 4 0 0 1-8 0" />
      </svg>
      {#if $cartCount > 0}
        <span class="nav__cart-badge" aria-hidden="true">{$cartCount}</span>
      {/if}
    </a>
  </div>
</nav>

<style>
  .nav {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;
    background: transparent;
    transition:
      background 0.3s ease,
      backdrop-filter 0.3s ease,
      border-bottom-color 0.3s ease;
    border-bottom: 1px solid transparent;
  }

  .nav--scrolled {
    background: rgba(3, 3, 8, 0.7);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom-color: rgba(240, 237, 230, 0.08);
  }

  .nav__inner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    max-width: 1280px;
    margin: 0 auto;
    padding: 1.25rem 2rem;
    gap: 2rem;
  }

  .nav__logo {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    text-decoration: none;
    color: rgba(240, 237, 230, 1);
    flex-shrink: 0;
  }

  .nav__logo-symbol {
    font-size: 1.25rem;
    line-height: 1;
    opacity: 0.8;
  }

  .nav__logo-text {
    font-family: 'Cormorant Garamond', serif;
    font-weight: 300;
    font-size: 1.1rem;
    letter-spacing: 0.25em;
    text-transform: uppercase;
  }

  .nav__links {
    display: flex;
    align-items: center;
    gap: 2.5rem;
    list-style: none;
    margin: 0 auto;
  }

  .nav__link {
    font-family: 'Inter', sans-serif;
    font-weight: 300;
    font-size: 0.8rem;
    letter-spacing: 0.15em;
    text-transform: uppercase;
    text-decoration: none;
    color: rgba(240, 237, 230, 0.7);
    transition: color 0.2s ease;
  }

  .nav__link:hover,
  .nav__link--active {
    color: rgba(240, 237, 230, 1);
  }

  .nav__cart {
    position: relative;
    display: flex;
    align-items: center;
    color: rgba(240, 237, 230, 0.7);
    text-decoration: none;
    transition: color 0.2s ease;
    flex-shrink: 0;
  }

  .nav__cart:hover {
    color: rgba(240, 237, 230, 1);
  }

  .nav__cart-badge {
    position: absolute;
    top: -6px;
    right: -8px;
    min-width: 16px;
    height: 16px;
    background: rgba(240, 237, 230, 0.9);
    color: #030308;
    border-radius: 8px;
    font-family: 'Inter', sans-serif;
    font-weight: 400;
    font-size: 0.65rem;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 3px;
    line-height: 1;
  }
</style>
