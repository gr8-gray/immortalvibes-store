<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { getOrder } from '$lib/api';
  import type { Order } from '$lib/api';

  let order: Order | null = null;
  let status: 'loading' | 'polling' | 'paid' | 'error' = 'loading';
  let errorMsg = '';
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  $: orderId = $page.params.id;

  async function fetchOrder() {
    try {
      order = await getOrder(orderId);
      if (order.status === 'paid' || order.status === 'fulfilled' || order.status === 'complete') {
        status = 'paid';
        if (pollTimer) clearInterval(pollTimer);
      } else {
        status = 'polling';
      }
    } catch (e: unknown) {
      errorMsg = e instanceof Error ? e.message : 'Failed to load order.';
      status = 'error';
      if (pollTimer) clearInterval(pollTimer);
    }
  }

  onMount(() => {
    fetchOrder();
    pollTimer = setInterval(fetchOrder, 3000);
  });

  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer);
  });

  function formatPrice(cents: number, currency: string): string {
    const symbol = currency.toLowerCase() === 'gbp' ? '£' : '$';
    return `${symbol}${(cents / 100).toFixed(2)}`;
  }
</script>

<svelte:head>
  <title>Order Confirmed — Immortal Vibes</title>
</svelte:head>

<div class="order-page">
  {#if status === 'loading'}
    <p class="status-msg">Loading your order…</p>

  {:else if status === 'polling'}
    <div class="polling-state">
      <div class="pulse-ring"></div>
      <p class="polling-msg">Confirming your payment…</p>
      <p class="polling-sub">This usually takes just a moment.</p>
    </div>

  {:else if status === 'error'}
    <p class="error-msg">{errorMsg}</p>

  {:else if status === 'paid' && order}
    <div class="confirmation">
      <p class="mission-label">ORDER CONFIRMED</p>
      <h1 class="confirm-title">Thank you.</h1>
      <p class="order-id-label">ORDER · {order.id.slice(0, 8).toUpperCase()}</p>

      <div class="order-total-row">
        <span class="total-label">TOTAL</span>
        <span class="total-amount">
          {formatPrice(order.total_amount, order.currency)}
        </span>
      </div>

      <p class="fulfillment-note">
        Your gear is being prepared for launch. A confirmation email has been sent to {order.email}.
      </p>

      <a href="/shop" class="continue-link">CONTINUE SHOPPING</a>
    </div>
  {/if}
</div>

<style>
  .order-page {
    max-width: 520px;
    margin: 0 auto;
    padding: 6rem 1.5rem;
    min-height: 60vh;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .status-msg,
  .error-msg {
    font-family: 'Inter', sans-serif;
    font-size: 0.8rem;
    color: rgba(240, 237, 230, 0.4);
    margin: 0;
  }

  .error-msg {
    color: rgba(200, 100, 80, 0.85);
  }

  .polling-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    text-align: center;
  }

  .pulse-ring {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    border: 2px solid rgba(200, 146, 42, 0.4);
    animation: pulse 1.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { transform: scale(1); opacity: 0.6; }
    50% { transform: scale(1.15); opacity: 1; }
  }

  .polling-msg {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.4rem;
    color: #F0EDE6;
    margin: 0;
    font-weight: 300;
  }

  .polling-sub {
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    color: rgba(240, 237, 230, 0.35);
    letter-spacing: 0.05em;
    margin: 0;
  }

  .confirmation {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    width: 100%;
  }

  .mission-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.3em;
    color: rgba(240, 237, 230, 0.35);
    margin: 0;
  }

  .confirm-title {
    font-family: 'Cormorant Garamond', serif;
    font-size: clamp(2.5rem, 6vw, 4.5rem);
    font-weight: 300;
    color: #F0EDE6;
    margin: 0;
    line-height: 1;
  }

  .order-id-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.3);
    margin: 0;
  }

  .order-items {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 1.25rem 0;
    border-top: 1px solid rgba(240, 237, 230, 0.07);
    border-bottom: 1px solid rgba(240, 237, 230, 0.07);
  }

  .order-row {
    display: flex;
    justify-content: space-between;
  }

  .order-variant {
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    color: rgba(240, 237, 230, 0.55);
  }

  .order-qty {
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    color: rgba(240, 237, 230, 0.35);
  }

  .order-total-row {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
  }

  .total-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.35);
  }

  .total-amount {
    font-family: 'Cormorant Garamond', serif;
    font-size: 1.5rem;
    color: #C8922A;
  }

  .fulfillment-note {
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    line-height: 1.7;
    color: rgba(240, 237, 230, 0.4);
    margin: 0;
  }

  .continue-link {
    display: inline-block;
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.5);
    text-decoration: none;
    border-bottom: 1px solid rgba(240, 237, 230, 0.15);
    padding-bottom: 2px;
    width: fit-content;
    transition: color 0.15s, border-color 0.15s;
  }

  .continue-link:hover {
    color: #F0EDE6;
    border-color: rgba(240, 237, 230, 0.4);
  }
</style>
