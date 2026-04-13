<!-- web/src/routes/shop/[slug]/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { gsap } from 'gsap';
  import type { PageData } from './$types';
  import MissionScene from '$lib/components/MissionScene.svelte';
  import SizeSelector from '$lib/components/SizeSelector.svelte';
  import StockBadge from '$lib/components/StockBadge.svelte';
  import ParticleBurst from '$lib/components/ParticleBurst.svelte';
  import { revealOnScroll } from '$lib/animations/reveal';
  import { cart } from '$lib/stores/cart';
  import { openCart } from '$lib/stores/cartDrawer';
  import { addItemToCart } from '$lib/api';
  import type { CartItem } from '$lib/stores/cart';
  import { MISSION_ORDER } from '$lib/stores/transition';
  import { goto } from '$app/navigation';

  export let data: PageData;

  $: product     = data.product;
  $: variants    = product.variants ?? [];
  $: hasVariants = variants.length > 0;

  // Active variant & gallery state
  let activeVariantIdx = 0;
  let activeGalleryIdx = 0;

  // Reset state on navigation (next/prev mission)
  let _lastSlug = '';
  $: if (product.slug !== _lastSlug) {
    _lastSlug = product.slug;
    activeVariantIdx = 0;
    activeGalleryIdx = (product.variants?.[0]?.gallery?.length ?? 0) > 0 ? 0 : -1;
  }

  $: activeVariant  = hasVariants ? variants[activeVariantIdx] : null;
  $: galleryImages  = activeVariant?.gallery ?? [];

  // If no gallery images, fall back to standalone
  $: effectiveIdx = galleryImages.length > 0 ? activeGalleryIdx : -1;

  // Main displayed image — gallery first, standalone last
  $: mainImage = effectiveIdx >= 0
    ? galleryImages[effectiveIdx]
    : (activeVariant?.productImage ?? product.image_url);

  $: isProductShot = effectiveIdx < 0;

  function selectVariant(idx: number) {
    activeVariantIdx = idx;
    // Reset to first model shot for new variant (or standalone if no gallery)
    activeGalleryIdx = variants[idx]?.gallery?.length > 0 ? 0 : -1;
  }

  function selectGallery(idx: number) {
    activeGalleryIdx = idx;
  }

  function showStandalone() {
    activeGalleryIdx = -1;
  }

  // Mission nav
  $: currentIndex = MISSION_ORDER.indexOf(product.slug ?? '');
  $: prevSlug = currentIndex > 0 ? MISSION_ORDER[currentIndex - 1] : null;
  $: nextSlug = currentIndex < MISSION_ORDER.length - 1 ? MISSION_ORDER[currentIndex + 1] : null;

  function navigateMission(slug: string) {
    goto(`/shop/${slug}`, { noScroll: true });
  }

  $: missionNumber = product.mission_number;

  let selectedSize = '';
  let burst: ParticleBurst;
  let productImageEl: HTMLImageElement;
  let heroContent: HTMLDivElement;
  let cartError = '';
  let floatTween: gsap.core.Tween | null = null;

  const missionLabels: Record<string, string> = {
    '001': 'Low Earth Orbit',
    '002': 'Lunar Surface',
    '003': 'Stellar Nursery',
    '004': 'Deep Space',
  };

  function formatPrice(cents: number, currency: string): string {
    return `${currency === 'gbp' ? '£' : '$'}${(cents / 100).toFixed(2)}`;
  }

  function getDisplayPrice(): string {
    return product.currency === 'gbp'
      ? formatPrice(product.price_gbp, 'gbp')
      : formatPrice(product.price_usd, 'usd');
  }

  // Float animation — only on standalone product shot
  $: if (productImageEl) {
    floatTween?.kill();
    if (isProductShot) {
      floatTween = gsap.to(productImageEl, {
        y: -12, rotation: 1.5, duration: 4,
        ease: 'sine.inOut', repeat: -1, yoyo: true,
      });
    } else {
      gsap.set(productImageEl, { y: 0, rotation: 0 });
    }
  }

  async function handleAddToCart(e: MouseEvent) {
    if (!selectedSize) {
      cartError = 'Please select a size.';
      gsap.to('.size-selector', { x: [-6, 6, -4, 4, 0], duration: 0.35, ease: 'none' });
      return;
    }
    cartError = '';
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    burst.trigger(rect.left + rect.width / 2, rect.top);

    try {
      const goCart = await addItemToCart({
        price_id:   product.price_id,
        product_id: product.id,
        name:       `${product.name} / ${selectedSize}`,
        image_url:  product.image_url ?? '',
        currency:   'usd',
        amount:     product.price_usd,
        quantity:   1,
      });

      // Map Go line items → CartItem[] for the store
      const items: CartItem[] = goCart.line_items.map(li => ({
        variantId:  li.price_id,
        productId:  li.product_id,
        title:      li.name,
        quantity:   li.quantity,
        unitPrice:  li.amount,
        currency:   li.currency,
      }));

      cart.setCart(goCart.token, items);
      openCart();
    } catch (err) {
      cartError = err instanceof Error ? err.message : 'Failed to add to cart.';
    }
  }

  onMount(() => {
    if (heroContent) {
      heroContent.querySelectorAll<HTMLElement>('.reveal-child')
        .forEach((el, i) => revealOnScroll(el, i * 0.08));
    }
  });
</script>

<svelte:head>
  <title>{product.name} — Immortal Vibes</title>
  <meta name="description" content={product.description} />
</svelte:head>

<ParticleBurst bind:this={burst} />

<MissionScene missionNumber={missionNumber}>
  <div class="product-page">

    <!-- Mission label -->
    <div class="mission-tag reveal-child">
      <span class="mission-number">{missionNumber}</span>
      <span class="mission-name">{missionLabels[missionNumber]}</span>
    </div>

    <!-- Mission prev/next nav -->
    {#if prevSlug || nextSlug}
      <div class="mission-nav">
        {#if prevSlug}
          <button class="mission-nav-btn" on:click={() => prevSlug && navigateMission(prevSlug)} aria-label="Previous mission">
            ← PREV MISSION
          </button>
        {:else}<span></span>{/if}
        {#if nextSlug}
          <button class="mission-nav-btn" on:click={() => nextSlug && navigateMission(nextSlug)} aria-label="Next mission">
            NEXT MISSION →
          </button>
        {/if}
      </div>
    {/if}

    <div class="product-layout">

      <!-- ── Left: image display ── -->
      <div class="image-col">
        <!-- Main image -->
        <div class="main-image-wrap" class:product-shot={isProductShot}>
          <img
            bind:this={productImageEl}
            class="main-image"
            src={mainImage}
            alt={product.name}
            style={isProductShot && activeVariant?.imageScale
              ? `transform: scale(${activeVariant.imageScale})`
              : ''}
          />
        </div>

        <!-- Thumbnail strip: model shots first, standalone last -->
        {#if hasVariants && (activeVariant?.productImage || galleryImages.length > 0)}
          <div class="thumb-strip">
            <!-- Model / gallery shots first -->
            {#each galleryImages as img, i}
              <button
                class="thumb"
                class:active={effectiveIdx === i}
                on:click={() => selectGallery(i)}
                aria-label="Gallery image {i + 1}"
              >
                <img src={img} alt="gallery {i + 1}" />
              </button>
            {/each}

            <!-- Standalone product shot last -->
            {#if activeVariant?.productImage}
              <button
                class="thumb"
                class:active={isProductShot}
                on:click={showStandalone}
                aria-label="Product standalone"
              >
                <img src={activeVariant.productImage} alt="standalone" />
              </button>
            {/if}
          </div>
        {/if}
      </div>

      <!-- ── Right: product details ── -->
      <div bind:this={heroContent} class="details-col">
        <p class="reveal-child product-category">MISSION {missionNumber} · IMMORTAL VIBES</p>

        <h1 class="reveal-child product-name">{product.name}</h1>

        <p class="reveal-child product-price">{getDisplayPrice()}</p>

        <!-- Color variant swatches -->
        {#if variants.length > 1}
          <div class="reveal-child variant-section">
            <p class="field-label">COLOR — {activeVariant?.colorName ?? ''}</p>
            <div class="swatches">
              {#each variants as v, i}
                <button
                  class="swatch"
                  class:active={activeVariantIdx === i}
                  style="--c:{v.hex}"
                  on:click={() => selectVariant(i)}
                  aria-label={v.colorName}
                  title={v.colorName}
                ></button>
              {/each}
            </div>
          </div>
        {/if}

        <p class="reveal-child product-description">{product.description}</p>

        {#if product.status === 'available'}
          <div class="reveal-child">
            <p class="field-label">SELECT SIZE</p>
            <SizeSelector sizes={product.sizes} bind:selected={selectedSize} />
            {#if cartError}<p class="cart-error">{cartError}</p>{/if}
          </div>

          <button class="reveal-child add-to-cart" on:click={handleAddToCart} data-magnetic>
            ADD TO CART
          </button>
        {:else}
          <div class="reveal-child"><StockBadge status={product.status} /></div>
        {/if}
      </div>
    </div>
  </div>
</MissionScene>

<style>
  :global(body) { overflow-y: auto; }

  .product-page {
    min-height: 100vh;
    padding: 2rem;
    display: flex;
    flex-direction: column;
  }

  .mission-tag {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1.5rem 0 3rem;
  }

  .mission-number,
  .mission-name {
    font-family: 'Inter', sans-serif;
    font-size: 0.60rem;
    letter-spacing: 0.30em;
    color: rgba(240, 237, 230, 0.35);
    text-transform: uppercase;
  }

  /* ── Layout ── */
  .product-layout {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6rem;
    align-items: start;
    flex: 1;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
    padding-bottom: 6rem;
  }

  @media (max-width: 768px) {
    .product-layout { grid-template-columns: 1fr; gap: 2rem; }
    .image-col { position: relative; top: 0; height: auto; }
    .main-image-wrap { min-height: 280px; }
    .main-image { max-height: 50vh; }
    .thumb { width: 52px; height: 52px; }
    .details-col { padding-top: 0; }
  }

  /* ── Image column ── */
  .image-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
    position: sticky;
    top: 4rem;
  }

  .main-image-wrap {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 420px;
  }

  .main-image {
    max-width: 100%;
    max-height: 65vh;
    object-fit: contain;
    filter: drop-shadow(0 24px 80px rgba(0,0,0,0.6));
    will-change: transform;
    transition: opacity 0.2s ease;
  }

  /* Standalone product shot gets extra drop shadow glow */
  .product-shot .main-image {
    filter: drop-shadow(0 24px 80px rgba(0,0,0,0.6))
            drop-shadow(0 0 40px rgba(255,255,255,0.04));
  }

  /* ── Thumbnail strip ── */
  .thumb-strip {
    display: flex;
    gap: 0.6rem;
    flex-wrap: wrap;
    justify-content: center;
    max-width: 100%;
  }

  .thumb {
    width: 56px;
    height: 56px;
    border: 1px solid rgba(240,237,230,0.1);
    background: rgba(0,0,0,0.4);
    padding: 0;
    cursor: pointer;
    overflow: hidden;
    transition: border-color 0.2s ease;
    flex-shrink: 0;
  }

  .thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center 20%;
    display: block;
  }

  .thumb.active {
    border-color: rgba(240,237,230,0.55);
  }

  .thumb:hover:not(.active) {
    border-color: rgba(240,237,230,0.28);
  }

  /* ── Details column ── */
  .details-col {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding-top: 1rem;
  }

  .product-category {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.28em;
    color: rgba(240,237,230,0.35);
    margin: 0;
  }

  .product-name {
    font-family: 'Cormorant Garamond', serif;
    font-size: clamp(2rem, 5vw, 4rem);
    font-weight: 300;
    color: #F0EDE6;
    margin: 0;
    line-height: 1.05;
  }

  .product-price {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.5rem;
    color: #C8922A;
    margin: 0;
    letter-spacing: 0.02em;
  }

  /* ── Color swatches ── */
  .variant-section { display: flex; flex-direction: column; gap: 0.75rem; }

  .swatches { display: flex; gap: 0.6rem; align-items: center; }

  .swatch {
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--c);
    border: 1px solid rgba(240,237,230,0.15);
    cursor: pointer;
    transition: transform 0.15s ease, border-color 0.15s ease;
    padding: 0;
  }

  .swatch:hover { transform: scale(1.15); }

  .swatch.active {
    border-color: rgba(240,237,230,0.7);
    box-shadow: 0 0 0 2px rgba(240,237,230,0.18);
    transform: scale(1.1);
  }

  /* ── Rest of details ── */
  .product-description {
    font-family: 'Inter', sans-serif;
    font-size: 0.875rem;
    line-height: 1.7;
    color: rgba(240,237,230,0.55);
    margin: 0;
    max-width: 38ch;
  }

  .field-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.22em;
    color: rgba(240,237,230,0.35);
    margin: 0 0 0.75rem;
  }

  .cart-error {
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    color: rgba(240,237,230,0.45);
    margin: 0.5rem 0 0;
  }

  .add-to-cart {
    display: inline-block;
    width: 100%;
    max-width: 320px;
    padding: 1.1rem 2rem;
    background: #F0EDE6;
    color: #030308;
    border: none;
    font-family: 'Inter', sans-serif;
    font-size: 0.60rem;
    letter-spacing: 0.22em;
    cursor: none;
    transition: background 0.2s, transform 0.1s;
  }

  .add-to-cart:hover { background: #ffffff; transform: translateY(-1px); }
  .add-to-cart:active { transform: translateY(0); }

  /* ── Mission nav ── */
  .mission-nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 0 2rem;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
  }

  .mission-nav-btn {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.22em;
    color: rgba(240,237,230,0.35);
    background: none;
    border: 1px solid rgba(240,237,230,0.1);
    padding: 0.6rem 1.2rem;
    cursor: none;
    transition: color 0.2s, border-color 0.2s;
    text-transform: uppercase;
  }

  .mission-nav-btn:hover {
    color: rgba(240,237,230,0.75);
    border-color: rgba(240,237,230,0.25);
  }
</style>
