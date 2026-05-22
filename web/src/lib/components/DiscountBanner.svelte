<!-- web/src/lib/components/DiscountBanner.svelte -->
<!-- Fixed bottom strip — click opens email capture. Hides after subscribe or dismiss. -->
<script lang="ts">
  import { browser } from '$app/environment';

  export let onOpen: () => void;

  const STORAGE_KEY = 'iv_subscribed';

  let dismissed = false;
  let subscribed = false;

  $: if (browser) {
    subscribed = localStorage.getItem(STORAGE_KEY) === '1';
  }

  export function markSubscribed() {
    subscribed = true;
    if (browser) localStorage.setItem(STORAGE_KEY, '1');
  }

  function dismiss(e: MouseEvent) {
    e.stopPropagation();
    dismissed = true;
  }
</script>

{#if !dismissed && !subscribed}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="banner" on:click={onOpen} role="button" tabindex="0">
    <span class="banner-text">
      FIRST ORDER — <span class="gold">10% OFF</span> · ENTER THE FREQUENCY →
    </span>
    <button class="dismiss" on:click={dismiss} aria-label="Dismiss offer" tabindex="0">✕</button>
  </div>
{/if}

<style>
  .banner {
    position: fixed;
    top: 5rem;
    left: 0;
    right: 0;
    height: 34px;
    background: rgba(3, 3, 8, 0.92);
    border-bottom: 1px solid rgba(200, 146, 42, 0.25);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 80;
    cursor: pointer;
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    transition: border-color 0.25s, background 0.25s;
  }

  .banner:hover {
    background: rgba(8, 6, 20, 0.96);
    border-bottom-color: rgba(200, 146, 42, 0.55);
  }

  .banner-text {
    font-family: 'Inter', sans-serif;
    font-size: 0.52rem;
    letter-spacing: 0.32em;
    color: rgba(240, 237, 230, 0.75);
    text-transform: uppercase;
    pointer-events: none;
    user-select: none;
  }

  .gold {
    color: #C8922A;
  }

  .dismiss {
    position: absolute;
    right: 1.25rem;
    background: none;
    border: none;
    color: rgba(240, 237, 230, 0.3);
    font-size: 0.65rem;
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    line-height: 1;
    transition: color 0.15s;
  }

  .dismiss:hover {
    color: rgba(240, 237, 230, 0.7);
  }
</style>
