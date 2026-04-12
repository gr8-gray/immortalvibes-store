<!-- web/src/routes/+layout.svelte -->
<script lang="ts">
  import '../app.css';
  import { browser } from '$app/environment';
  import { beforeNavigate, goto } from '$app/navigation';
  import { page } from '$app/stores';
  import StarField from '$lib/components/StarField.svelte';
  import Nav from '$lib/components/Nav.svelte';
  import Footer from '$lib/components/Footer.svelte';
  import MagneticCursor from '$lib/components/MagneticCursor.svelte';
  import CartDrawer from '$lib/components/CartDrawer.svelte';
  import TransitionOverlay from '$lib/components/TransitionOverlay.svelte';
  import { resolveTransition, slugFromPath, MISSION_ACCENT, transitionStore } from '$lib/stores/transition';

  // Hide chrome on immersive pages
  const immersive = $derived($page.url.pathname === '/' || $page.url.pathname === '/shop');
  const isMobile = browser && /Android|iPhone|iPad|iPod/i.test(navigator.userAgent);

  let { children } = $props();

  let overlayComponent: TransitionOverlay;
  let mainEl: HTMLElement;
  let navigating = false;

  if (browser) {
    beforeNavigate(({ to, cancel }) => {
      if (navigating || !to) return;

      const fromPath = window.location.pathname;
      const toPath = to.url.pathname;
      const transType = resolveTransition(fromPath, toPath);

      if (!transType) return;

      cancel();
      navigating = true;

      const slug = slugFromPath(toPath);
      const accent = MISSION_ACCENT[slug] ?? '#4FC3F7';

      let cx = 0.5;
      let cy = 0.5;
      const unsub = transitionStore.subscribe((s) => { cx = s.clickX; cy = s.clickY; });
      unsub();

      overlayComponent.triggerOut(
        transType,
        { clickX: cx, clickY: cy, accentColor: accent, mainContent: mainEl },
        () => {
          goto(toPath, { noScroll: true }).then(() => {
            overlayComponent.triggerIn(transType, () => {
              navigating = false;
            });
          });
        }
      );
    });
  }
</script>

{#if !isMobile}<MagneticCursor />{/if}
<StarField />
<TransitionOverlay bind:this={overlayComponent} />
{#if !immersive}<Nav />{/if}

<main bind:this={mainEl} class="layout-main">
  {@render children()}
</main>

{#if !immersive}<Footer />{/if}
{#if !immersive}<CartDrawer />{/if}

<style>
  :global(body) {
    background-color: var(--void);
  }

  .layout-main {
    position: relative;
    z-index: 10;
    padding-top: 4.5rem;
    min-height: 100vh;
  }
</style>
