<script lang="ts">
  import '../app.css';
  import { browser } from '$app/environment';
  import { onNavigate } from '$app/navigation';
  import StarField from '$lib/components/StarField.svelte';
  import Nav from '$lib/components/Nav.svelte';
  import Footer from '$lib/components/Footer.svelte';
  import MagneticCursor from '$lib/components/MagneticCursor.svelte';

  let { children } = $props();

  // View Transition API for page transitions (progressive enhancement)
  if (browser) {
    onNavigate((navigation) => {
      if (!document.startViewTransition) return;
      return new Promise((resolve) => {
        document.startViewTransition(async () => {
          resolve();
          await navigation.complete;
        });
      });
    });
  }
</script>

<!-- Magnetic cursor — desktop only, hidden on touch -->
<MagneticCursor />

<!-- Star field fixed behind everything -->
<StarField />

<!-- Page chrome: nav sits above star field -->
<Nav />

<!-- Main content -->
<main class="layout-main">
  {@render children()}
</main>

<!-- Footer -->
<Footer />

<style>
  :global(body) {
    background-color: var(--void);
  }

  .layout-main {
    position: relative;
    z-index: 10;
    /* Offset for fixed nav height (~72px) */
    padding-top: 4.5rem;
    min-height: 100vh;
  }
</style>
