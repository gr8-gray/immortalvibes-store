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

  export let data: PageData;

  $: product = data.product;
  $: missionNumber = product.mission_number;

  let selectedSize = '';
  let burst: ParticleBurst;
  let productImage: HTMLImageElement;
  let heroContent: HTMLDivElement;
  let addCartBtn: HTMLButtonElement;
  let cartError = '';

  const missionLabels: Record<string, string> = {
    '001': 'Low Earth Orbit',
    '002': 'Lunar Surface',
    '003': 'Stellar Nursery',
    '004': 'Deep Space',
  };

  function formatPrice(cents: number, currency: string): string {
    const symbol = currency === 'gbp' ? '£' : '$';
    return `${symbol}${(cents / 100).toFixed(2)}`;
  }

  function getDisplayPrice(): string {
    if (product.currency === 'gbp') return formatPrice(product.price_gbp, 'gbp');
    return formatPrice(product.price_usd, 'usd');
  }

  function handleAddToCart(e: MouseEvent) {
    if (!selectedSize) {
      cartError = 'Please select a size.';
      // Shake the size selector
      gsap.to('.size-selector', {
        x: [-6, 6, -4, 4, 0],
        duration: 0.35,
        ease: 'none',
      });
      return;
    }
    cartError = '';

    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    burst.trigger(rect.left + rect.width / 2, rect.top);

    // Placeholder: cart store will be wired in Plan 5
    console.log('Add to cart:', product.id, selectedSize);
  }

  onMount(() => {
    // Floating product image animation
    if (productImage) {
      gsap.to(productImage, {
        y: -12,
        rotation: 1.5,
        duration: 4,
        ease: 'sine.inOut',
        repeat: -1,
        yoyo: true,
      });
    }

    // Hero content reveal
    if (heroContent) {
      const children = heroContent.querySelectorAll<HTMLElement>('.reveal-child');
      children.forEach((el, i) => revealOnScroll(el, i * 0.08));
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
    <!-- Mission label top-left -->
    <div class="mission-tag reveal-child">
      <span class="mission-number">{missionNumber}</span>
      <span class="mission-name">{missionLabels[missionNumber]}</span>
    </div>

    <div class="product-layout">
      <!-- Left: product image -->
      <div class="image-col">
        {#if product.image_url}
          <img
            bind:this={productImage}
            class="product-hero-image"
            src={product.image_url}
            alt={product.name}
          />
        {:else}
          <div class="image-placeholder"></div>
        {/if}
      </div>

      <!-- Right: product details -->
      <div bind:this={heroContent} class="details-col">
        <p class="reveal-child product-category">
          MISSION {missionNumber} · IMMORTAL VIBES
        </p>

        <h1 class="reveal-child product-name">{product.name}</h1>

        <p class="reveal-child product-price">{getDisplayPrice()}</p>

        <p class="reveal-child product-description">{product.description}</p>

        {#if product.status === 'available'}
          <div class="reveal-child">
            <p class="field-label">SELECT SIZE</p>
            <SizeSelector
              sizes={product.sizes}
              bind:selected={selectedSize}
            />
            {#if cartError}
              <p class="cart-error">{cartError}</p>
            {/if}
          </div>

          <button
            bind:this={addCartBtn}
            class="reveal-child add-to-cart"
            on:click={handleAddToCart}
            data-magnetic
          >
            ADD TO CART
          </button>
        {:else}
          <div class="reveal-child">
            <StockBadge status={product.status} />
          </div>
        {/if}
      </div>
    </div>
  </div>
</MissionScene>

<style>
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

  .mission-number {
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.3em;
    color: rgba(240, 237, 230, 0.35);
  }

  .mission-name {
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.35);
    text-transform: uppercase;
  }

  .product-layout {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6rem;
    align-items: center;
    flex: 1;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
    padding-bottom: 6rem;
  }

  @media (max-width: 768px) {
    .product-layout {
      grid-template-columns: 1fr;
      gap: 3rem;
    }
  }

  .image-col {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .product-hero-image {
    max-width: 100%;
    max-height: 70vh;
    object-fit: contain;
    filter: drop-shadow(0 24px 80px rgba(0, 0, 0, 0.6));
    will-change: transform;
  }

  .image-placeholder {
    width: 300px;
    height: 400px;
    border: 1px solid rgba(240, 237, 230, 0.06);
  }

  .details-col {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .product-category {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.25em;
    color: rgba(240, 237, 230, 0.35);
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

  .product-description {
    font-family: 'Inter', sans-serif;
    font-size: 0.875rem;
    line-height: 1.7;
    color: rgba(240, 237, 230, 0.55);
    margin: 0;
    max-width: 38ch;
  }

  .field-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.35);
    margin: 0 0 0.75rem;
  }

  .cart-error {
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    color: rgba(240, 237, 230, 0.45);
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
    font-size: 0.7rem;
    letter-spacing: 0.2em;
    cursor: none;
    transition: background 0.2s, transform 0.1s;
  }

  .add-to-cart:hover {
    background: #ffffff;
    transform: translateY(-1px);
  }

  .add-to-cart:active {
    transform: translateY(0);
  }
</style>
