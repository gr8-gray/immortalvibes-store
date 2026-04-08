<script lang="ts">
  import { goto } from '$app/navigation';
  import { cart } from '$lib/stores/cart';
  import { isCartOpen, closeCart } from '$lib/stores/cartDrawer';

  function formatPrice(cents: number, currency: string): string {
    const symbol = currency.toLowerCase() === 'gbp' ? '£' : '$';
    return `${symbol}${(cents / 100).toFixed(2)}`;
  }

  $: subtotalCents = $cart.items.reduce(
    (sum, item) => sum + item.unitPrice * item.quantity,
    0
  );
  $: subtotalCurrency = $cart.items[0]?.currency ?? 'usd';

  function handleRemove(variantId: string) {
    cart.removeItem(variantId);
  }

  function handleCheckout() {
    closeCart();
    goto('/checkout');
  }

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) closeCart();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') closeCart();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- Backdrop -->
{#if $isCartOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div
    class="backdrop"
    on:click={handleBackdropClick}
    aria-hidden="true"
  ></div>
{/if}

<!-- Drawer -->
<aside
  class="drawer"
  class:open={$isCartOpen}
  aria-label="Shopping cart"
  role="dialog"
  aria-modal="true"
>
  <!-- Header -->
  <div class="drawer-header">
    <span class="drawer-title">CART</span>
    <button class="close-btn" on:click={closeCart} aria-label="Close cart">
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
        <line x1="1" y1="1" x2="13" y2="13" stroke="currentColor" stroke-width="1.5"/>
        <line x1="13" y1="1" x2="1" y2="13" stroke="currentColor" stroke-width="1.5"/>
      </svg>
    </button>
  </div>

  <!-- Items -->
  <div class="drawer-items">
    {#if $cart.items.length === 0}
      <p class="empty-msg">Your cart is empty.</p>
    {:else}
      {#each $cart.items as item (item.variantId)}
        <div class="cart-item">
          <div class="item-info">
            <p class="item-title">{item.title}</p>
            <p class="item-qty">Qty: {item.quantity}</p>
          </div>
          <div class="item-right">
            <span class="item-price">
              {formatPrice(item.unitPrice * item.quantity, item.currency)}
            </span>
            <button
              class="remove-btn"
              on:click={() => handleRemove(item.variantId)}
              aria-label="Remove {item.title}"
            >
              Remove
            </button>
          </div>
        </div>
      {/each}
    {/if}
  </div>

  <!-- Footer -->
  {#if $cart.items.length > 0}
    <div class="drawer-footer">
      <div class="subtotal-row">
        <span class="subtotal-label">SUBTOTAL</span>
        <span class="subtotal-amount">
          {formatPrice(subtotalCents, subtotalCurrency)}
        </span>
      </div>
      <button class="checkout-btn" on:click={handleCheckout}>
        PROCEED TO CHECKOUT
      </button>
    </div>
  {/if}
</aside>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(3, 3, 8, 0.6);
    z-index: 90;
    backdrop-filter: blur(2px);
  }

  .drawer {
    position: fixed;
    top: 0;
    right: 0;
    height: 100vh;
    width: min(420px, 100vw);
    background: #0d0d14;
    border-left: 1px solid rgba(240, 237, 230, 0.08);
    z-index: 100;
    display: flex;
    flex-direction: column;
    transform: translateX(100%);
    transition: transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .drawer.open {
    transform: translateX(0);
  }

  .drawer-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.75rem 1.5rem;
    border-bottom: 1px solid rgba(240, 237, 230, 0.08);
  }

  .drawer-title {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.25em;
    color: rgba(240, 237, 230, 0.5);
  }

  .close-btn {
    background: none;
    border: none;
    color: rgba(240, 237, 230, 0.5);
    cursor: pointer;
    padding: 0.25rem;
    transition: color 0.15s;
    line-height: 0;
  }

  .close-btn:hover {
    color: #F0EDE6;
  }

  .drawer-items {
    flex: 1;
    overflow-y: auto;
    padding: 1rem 1.5rem;
  }

  .empty-msg {
    font-family: 'Inter', sans-serif;
    font-size: 0.8rem;
    color: rgba(240, 237, 230, 0.35);
    margin-top: 2rem;
    text-align: center;
  }

  .cart-item {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    padding: 1.25rem 0;
    border-bottom: 1px solid rgba(240, 237, 230, 0.06);
  }

  .item-info {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }

  .item-title {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1rem;
    color: #F0EDE6;
    margin: 0;
    font-weight: 400;
  }

  .item-qty {
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    color: rgba(240, 237, 230, 0.35);
    letter-spacing: 0.1em;
    margin: 0;
  }

  .item-right {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.5rem;
  }

  .item-price {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1rem;
    color: #C8922A;
  }

  .remove-btn {
    background: none;
    border: none;
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.1em;
    color: rgba(240, 237, 230, 0.3);
    cursor: pointer;
    padding: 0;
    transition: color 0.15s;
    text-transform: uppercase;
  }

  .remove-btn:hover {
    color: rgba(240, 237, 230, 0.7);
  }

  .drawer-footer {
    padding: 1.5rem;
    border-top: 1px solid rgba(240, 237, 230, 0.08);
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .subtotal-row {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
  }

  .subtotal-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.4);
  }

  .subtotal-amount {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.35rem;
    color: #C8922A;
  }

  .checkout-btn {
    width: 100%;
    padding: 1.1rem 2rem;
    background: #F0EDE6;
    color: #030308;
    border: none;
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    letter-spacing: 0.2em;
    cursor: pointer;
    transition: background 0.2s, transform 0.1s;
  }

  .checkout-btn:hover {
    background: #ffffff;
    transform: translateY(-1px);
  }

  .checkout-btn:active {
    transform: translateY(0);
  }
</style>
