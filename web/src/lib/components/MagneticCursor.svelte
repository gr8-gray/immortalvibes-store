<!-- web/src/lib/components/MagneticCursor.svelte -->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import { gsap } from 'gsap';

  let cursorDot: HTMLDivElement;
  let cursorRing: HTMLDivElement;

  // Raw mouse position
  let mouseX = -100;
  let mouseY = -100;

  // Rendered positions (lerped)
  let dotX = -100;
  let dotY = -100;

  let rafId: number;
  let isHovering = false;

  const MAGNETIC_RADIUS = 100;
  const MAGNETIC_PULL = 0.3; // 30% toward element center

  function getMagneticElements(): NodeListOf<Element> {
    return document.querySelectorAll('[data-magnetic]');
  }

  function computeMagneticOffset(mx: number, my: number): { x: number; y: number } {
    for (const el of getMagneticElements()) {
      const rect = el.getBoundingClientRect();
      const cx = rect.left + rect.width / 2;
      const cy = rect.top + rect.height / 2;
      const dist = Math.hypot(mx - cx, my - cy);

      if (dist < MAGNETIC_RADIUS) {
        const pull = 1 - dist / MAGNETIC_RADIUS; // 0 → 1 as cursor approaches
        return {
          x: mx + (cx - mx) * MAGNETIC_PULL * pull,
          y: my + (cy - my) * MAGNETIC_PULL * pull,
        };
      }
    }
    return { x: mx, y: my };
  }

  function loop() {
    const target = computeMagneticOffset(mouseX, mouseY);

    // Lerp dot toward target
    dotX += (target.x - dotX) * 0.12;
    dotY += (target.y - dotY) * 0.12;

    gsap.set(cursorDot, { x: dotX, y: dotY });
    gsap.set(cursorRing, { x: dotX, y: dotY });

    rafId = requestAnimationFrame(loop);
  }

  function onMouseMove(e: MouseEvent) {
    mouseX = e.clientX;
    mouseY = e.clientY;
  }

  function onMouseDown() {
    gsap.to(cursorDot, { scale: 0.6, duration: 0.1 });
    gsap.to(cursorRing, { scale: 1.4, opacity: 0.4, duration: 0.15 });
  }

  function onMouseUp() {
    gsap.to(cursorDot, { scale: 1, duration: 0.2, ease: 'back.out(2)' });
    gsap.to(cursorRing, { scale: 1, opacity: 1, duration: 0.2 });
  }

  function onMouseEnterMagnetic() {
    isHovering = true;
    gsap.to(cursorDot, { scale: 1.6, duration: 0.2 });
  }

  function onMouseLeaveMagnetic() {
    isHovering = false;
    gsap.to(cursorDot, { scale: 1, duration: 0.2 });
  }

  onMount(() => {
    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mousedown', onMouseDown);
    window.addEventListener('mouseup', onMouseUp);

    // Attach hover listeners to all magnetic elements present at mount
    const attachHoverListeners = () => {
      for (const el of getMagneticElements()) {
        el.addEventListener('mouseenter', onMouseEnterMagnetic);
        el.addEventListener('mouseleave', onMouseLeaveMagnetic);
      }
    };

    attachHoverListeners();

    // Re-attach when DOM changes (e.g. route navigation adds new [data-magnetic] elements)
    const observer = new MutationObserver(attachHoverListeners);
    observer.observe(document.body, { childList: true, subtree: true });

    rafId = requestAnimationFrame(loop);

    return () => {
      observer.disconnect();
    };
  });

  onDestroy(() => {
    if (!browser) return;
    cancelAnimationFrame(rafId);
    window.removeEventListener('mousemove', onMouseMove);
    window.removeEventListener('mousedown', onMouseDown);
    window.removeEventListener('mouseup', onMouseUp);
  });
</script>

<svelte:head>
  <style>
    * { cursor: none !important; }
  </style>
</svelte:head>

<div bind:this={cursorRing} class="cursor-ring" aria-hidden="true"></div>
<div bind:this={cursorDot} class="cursor-dot" aria-hidden="true"></div>

<style>
  .cursor-dot,
  .cursor-ring {
    position: fixed;
    top: 0;
    left: 0;
    pointer-events: none;
    z-index: 9999;
    border-radius: 50%;
    transform: translate(-50%, -50%);
    will-change: transform;
  }

  .cursor-dot {
    width: 6px;
    height: 6px;
    background: #F0EDE6;
  }

  .cursor-ring {
    width: 32px;
    height: 32px;
    border: 1px solid rgba(240, 237, 230, 0.4);
    background: transparent;
    transition: opacity 0.2s;
  }
</style>
