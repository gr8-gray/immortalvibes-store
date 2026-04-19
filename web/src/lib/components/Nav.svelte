<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import { cartCount } from '$lib/stores/cart';
  import { openCart } from '$lib/stores/cartDrawer';
  import { page } from '$app/stores';

  let scrolled = $state(false);
  const isHomepage = $derived($page.url.pathname === '/' || $page.url.pathname === '/shop');
  let menuOpen = $state(false);
  let scrollHandler: (() => void) | null = null;

  onMount(() => {
    if (!browser) return;
    scrollHandler = () => { scrolled = window.scrollY > 10; };
    window.addEventListener('scroll', scrollHandler, { passive: true });
  });

  onDestroy(() => {
    if (browser && scrollHandler) window.removeEventListener('scroll', scrollHandler);
  });

  function isActive(href: string): boolean {
    return $page.url.pathname === href || $page.url.pathname.startsWith(href + '/');
  }

  function toggleMenu() { menuOpen = !menuOpen; }
  function closeMenu() { menuOpen = false; }

  // Close on route change
  $effect(() => { $page.url.pathname; closeMenu(); });
</script>

<nav class="nav" class:nav--scrolled={scrolled} class:nav--frosted={isHomepage} aria-label="Main navigation">
  <div class="nav__inner">

    <!-- Logo -->
    <a href="/" class="nav__logo" aria-label="Immortal Vibes home" onclick={closeMenu}>
      <img src="/logo-bare.png" alt="Immortal Vibes" class="nav__logo-img" />
    </a>

    <!-- Desktop links -->
    <ul class="nav__links" role="list">
      <li><a href="/shop"    class="nav__link" class:nav__link--active={isActive('/shop')}>Shop</a></li>
      <li><a href="/about"   class="nav__link" class:nav__link--active={isActive('/about')}>About</a></li>
      <li><a href="/contact" class="nav__link" class:nav__link--active={isActive('/contact')}>Contact</a></li>
    </ul>

    <!-- Right side: cart + hamburger -->
    <div class="nav__right">
      <button onclick={openCart} class="nav__cart" aria-label="Cart ({$cartCount} items)">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
             fill="none" stroke="currentColor" stroke-width="1.5"
             stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z" />
          <line x1="3" y1="6" x2="21" y2="6" />
          <path d="M16 10a4 4 0 0 1-8 0" />
        </svg>
        {#if $cartCount > 0}
          <span class="nav__cart-badge" aria-hidden="true">{$cartCount}</span>
        {/if}
      </button>

      <!-- Hamburger — mobile only -->
      <button
        class="nav__hamburger"
        class:nav__hamburger--open={menuOpen}
        onclick={toggleMenu}
        aria-label={menuOpen ? 'Close menu' : 'Open menu'}
        aria-expanded={menuOpen}
      >
        <span></span>
        <span></span>
        <span></span>
      </button>
    </div>

  </div>

  <!-- Mobile menu dropdown -->
  {#if menuOpen}
    <div class="nav__mobile-menu" role="navigation">
      <a href="/shop"    class="nav__mobile-link" class:active={isActive('/shop')}    onclick={closeMenu}>Shop</a>
      <a href="/about"   class="nav__mobile-link" class:active={isActive('/about')}   onclick={closeMenu}>About</a>
      <a href="/contact" class="nav__mobile-link" class:active={isActive('/contact')} onclick={closeMenu}>Contact</a>
    </div>
  {/if}
</nav>

<style>
  .nav {
    position: fixed;
    top: 0; left: 0; right: 0;
    z-index: 100;
    background: transparent;
    transition: background 0.3s ease, backdrop-filter 0.3s ease, border-bottom-color 0.3s ease;
    border-bottom: 1px solid transparent;
  }

  .nav--scrolled {
    background: rgba(3,3,8,0.82);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom-color: rgba(240,237,230,0.08);
  }

  /* Homepage: frosted from the start, lighter than scroll state */
  .nav--frosted {
    background: rgba(3,3,8,0.45);
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    border-bottom-color: rgba(240,237,230,0.06);
  }

  .nav__inner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    max-width: 1280px;
    margin: 0 auto;
    padding: 1rem 2rem;
    gap: 2rem;
  }

  /* ── Logo ── */
  .nav__logo {
    display: flex;
    align-items: center;
    text-decoration: none;
    flex-shrink: 0;
    opacity: 0.9;
    transition: opacity 0.2s ease;
  }
  .nav__logo:hover { opacity: 1; }
  .nav__logo-img { height: 52px; width: 52px; object-fit: contain; }

  /* ── Desktop links ── */
  .nav__links {
    display: flex;
    align-items: center;
    gap: 2.5rem;
    list-style: none;
    margin: 0 auto;
    padding: 0;
  }

  .nav__link {
    font-family: 'Inter', sans-serif;
    font-weight: 300;
    font-size: 0.60rem;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    text-decoration: none;
    color: rgba(240,237,230,0.6);
    transition: color 0.2s ease;
  }
  .nav__link:hover, .nav__link--active { color: rgba(240,237,230,1); }

  /* ── Right cluster ── */
  .nav__right {
    display: flex;
    align-items: center;
    gap: 1.2rem;
    flex-shrink: 0;
  }

  /* ── Cart ── */
  .nav__cart {
    position: relative;
    display: flex;
    align-items: center;
    color: rgba(240,237,230,0.6);
    text-decoration: none;
    transition: color 0.2s ease;
    background: none;
    border: none;
    padding: 0;
    cursor: none;
  }
  .nav__cart:hover { color: rgba(240,237,230,1); }

  .nav__cart-badge {
    position: absolute;
    top: -6px; right: -8px;
    min-width: 16px; height: 16px;
    background: rgba(240,237,230,0.9);
    color: #030308;
    border-radius: 8px;
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 3px;
    line-height: 1;
  }

  /* ── Hamburger — hidden on desktop ── */
  .nav__hamburger {
    display: none;
    flex-direction: column;
    justify-content: center;
    gap: 5px;
    width: 32px;
    height: 32px;
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
  }

  .nav__hamburger span {
    display: block;
    width: 100%;
    height: 1px;
    background: rgba(240,237,230,0.75);
    transition: transform 0.25s ease, opacity 0.25s ease;
    transform-origin: center;
  }

  .nav__hamburger--open span:nth-child(1) { transform: translateY(6px) rotate(45deg); }
  .nav__hamburger--open span:nth-child(2) { opacity: 0; }
  .nav__hamburger--open span:nth-child(3) { transform: translateY(-6px) rotate(-45deg); }

  /* ── Mobile menu ── */
  .nav__mobile-menu {
    display: none;
    flex-direction: column;
    padding: 1rem 2rem 1.5rem;
    border-top: 1px solid rgba(240,237,230,0.06);
    background: rgba(3,3,8,0.95);
    backdrop-filter: blur(16px);
    gap: 0;
  }

  .nav__mobile-link {
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    letter-spacing: 0.28em;
    text-transform: uppercase;
    text-decoration: none;
    color: rgba(240,237,230,0.6);
    padding: 1rem 0;
    border-bottom: 1px solid rgba(240,237,230,0.06);
    transition: color 0.2s ease;
  }

  .nav__mobile-link:last-child { border-bottom: none; }
  .nav__mobile-link:hover,
  .nav__mobile-link.active { color: rgba(240,237,230,1); }

  /* ── Mobile breakpoint ── */
  @media (max-width: 640px) {
    .nav__inner { padding: 0.85rem 1.25rem; gap: 1rem; }
    .nav__logo-img { height: 42px; width: 42px; }
    .nav__links { display: none; }
    .nav__hamburger { display: flex; }
    .nav__mobile-menu { display: flex; }
  }
</style>
