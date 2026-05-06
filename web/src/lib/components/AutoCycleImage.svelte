<!--
  AutoCycleImage — crossfades through an array of image URLs.
  Used on /shop/warped-reality-set cards (tee front↔back, beanie Blue↔Green).
  Single-image arrays render as a static image with no timer.

  Approach: every image in `images` is rendered once and stays mounted; only its
  opacity transitions on the active index change. Avoids the "swap flash" you get
  from mounting/unmounting layers around the moment the active image changes.
-->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  // Each image entry can be either a path string or an object with optional
  // per-image scale (compensates for tight/loose crops) and offsetX/Y in % of
  // frame width/height (compensates for off-center subjects in the source PNG).
  export type CycleImage =
    | string
    | { src: string; scale?: number; offsetX?: number; offsetY?: number };

  export let images: CycleImage[] = [];
  export let alt = '';
  export let interval = 2600;       // ms between swaps
  export let fade = 700;            // ms crossfade duration
  export let startOffsetMs = 0;     // stagger neighboring cards so they don't pulse in lockstep
  export let maxFill = 0.92;        // % of card the image is allowed to occupy

  $: items = images.map((it) =>
    typeof it === 'string'
      ? { src: it, scale: 1, offsetX: 0, offsetY: 0 }
      : {
          src:     it.src,
          scale:   it.scale   ?? 1,
          offsetX: it.offsetX ?? 0,
          offsetY: it.offsetY ?? 0,
        },
  );

  let activeIdx = 0;
  let cycleTimer: ReturnType<typeof setInterval> | null = null;
  let kickoffTimer: ReturnType<typeof setTimeout> | null = null;

  function step() {
    if (items.length < 2) return;
    activeIdx = (activeIdx + 1) % items.length;
  }

  onMount(() => {
    if (items.length < 2) return;
    kickoffTimer = setTimeout(() => {
      step();
      cycleTimer = setInterval(step, interval);
    }, startOffsetMs);
  });

  onDestroy(() => {
    if (cycleTimer) clearInterval(cycleTimer);
    if (kickoffTimer) clearTimeout(kickoffTimer);
  });

  $: fillPct = `${Math.round(maxFill * 100)}%`;
</script>

<div class="auto-cycle" style="--fade:{fade}ms; --fill:{fillPct}">
  {#each items as item, i (item.src)}
    <img
      class="layer"
      class:active={i === activeIdx}
      style="transform: translate({item.offsetX}%, {item.offsetY}%) scale({item.scale})"
      src={item.src}
      {alt}
      loading={i === 0 ? 'eager' : 'lazy'}
    />
  {/each}
</div>

<style>
  .auto-cycle {
    position: relative;
    width: 100%;
    height: 100%;
  }

  .layer {
    position: absolute;
    inset: 0;
    margin: auto;
    max-width:  var(--fill);
    max-height: var(--fill);
    width: auto;
    height: auto;
    object-fit: contain;
    opacity: 0;
    transition: opacity var(--fade) ease;
    will-change: opacity;
  }

  .layer.active { opacity: 1; }
</style>
