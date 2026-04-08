<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { loadStripe } from '@stripe/stripe-js';
  import { env } from '$env/dynamic/public';
  import { cart } from '$lib/stores/cart';
  import { createCheckout } from '$lib/api';
  import type { Stripe, StripeElements } from '@stripe/stripe-js';

  let stripe: Stripe | null = null;
  let elements: StripeElements | null = null;
  let paymentElement: ReturnType<StripeElements['create']> | null = null;
  let mountNode: HTMLDivElement;

  let orderId = '';
  let loading = true;
  let submitting = false;
  let errorMsg = '';
  let cartSnapshot = $cart;

  onMount(async () => {
    if (!browser) return;

    const pubKey = env.PUBLIC_STRIPE_PUBLISHABLE_KEY ?? '';
    stripe = await loadStripe(pubKey);
    if (!stripe) {
      errorMsg = 'Failed to load payment processor.';
      loading = false;
      return;
    }

    // Create checkout session
    try {
      const cartId = $cart.id;
      if (!cartId) {
        errorMsg = 'No active cart. Add items before checking out.';
        loading = false;
        return;
      }

      const session = await createCheckout(cartId, 'usd');
      // session.url is used as clientSecret in our integration
      // The worker returns { clientSecret, orderId } — map accordingly
      const clientSecret = (session as unknown as { clientSecret: string }).clientSecret;
      orderId = (session as unknown as { orderId: string }).orderId ?? session.id;

      elements = stripe.elements({ clientSecret });
      paymentElement = elements.create('payment');
      paymentElement.mount(mountNode);
    } catch (e: unknown) {
      errorMsg = e instanceof Error ? e.message : 'Failed to initialise checkout.';
    } finally {
      loading = false;
    }
  });

  onDestroy(() => {
    paymentElement?.destroy();
  });

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    if (!stripe || !elements || submitting) return;
    submitting = true;
    errorMsg = '';

    const { error } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: `${window.location.origin}/order/${orderId}`
      }
    });

    if (error) {
      errorMsg = error.message ?? 'Payment failed.';
      submitting = false;
    }
    // On success Stripe redirects to return_url automatically
  }
</script>

<svelte:head>
  <title>Checkout — Immortal Vibes</title>
</svelte:head>

<div class="checkout-page">
  <h1 class="page-title">CHECKOUT</h1>

  {#if loading}
    <p class="status-msg">Preparing your order…</p>
  {:else if errorMsg && !elements}
    <p class="error-msg">{errorMsg}</p>
    <button class="back-btn" on:click={() => goto('/shop')}>Back to Shop</button>
  {:else}
    <!-- Order summary -->
    <div class="order-summary">
      <p class="summary-label">ORDER SUMMARY</p>
      {#each cartSnapshot.items as item}
        <div class="summary-row">
          <span class="summary-item">{item.title} × {item.quantity}</span>
          <span class="summary-price">
            ${((item.unitPrice * item.quantity) / 100).toFixed(2)}
          </span>
        </div>
      {/each}
    </div>

    <!-- Stripe Payment Element -->
    <form on:submit={handleSubmit} class="payment-form">
      <div bind:this={mountNode} class="payment-element-mount"></div>

      {#if errorMsg}
        <p class="error-msg">{errorMsg}</p>
      {/if}

      <button type="submit" class="pay-btn" disabled={submitting}>
        {submitting ? 'PROCESSING…' : 'PAY NOW'}
      </button>
    </form>
  {/if}
</div>

<style>
  .checkout-page {
    max-width: 520px;
    margin: 0 auto;
    padding: 4rem 1.5rem 6rem;
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .page-title {
    font-family: 'Cormorant Garamond', serif;
    font-size: clamp(1.8rem, 4vw, 3rem);
    font-weight: 300;
    color: #F0EDE6;
    margin: 0;
    letter-spacing: 0.05em;
  }

  .status-msg {
    font-family: 'Inter', sans-serif;
    font-size: 0.8rem;
    color: rgba(240, 237, 230, 0.4);
    margin: 0;
  }

  .error-msg {
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    color: rgba(200, 100, 80, 0.85);
    margin: 0;
  }

  .order-summary {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding-bottom: 1.5rem;
    border-bottom: 1px solid rgba(240, 237, 230, 0.08);
  }

  .summary-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.6rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.35);
    margin: 0 0 0.25rem;
  }

  .summary-row {
    display: flex;
    justify-content: space-between;
  }

  .summary-item {
    font-family: 'Inter', sans-serif;
    font-size: 0.8rem;
    color: rgba(240, 237, 230, 0.65);
  }

  .summary-price {
    font-family: 'Cormorant Garamond', serif;
    font-size: 0.95rem;
    color: #C8922A;
  }

  .payment-form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .payment-element-mount {
    min-height: 200px;
  }

  .pay-btn {
    width: 100%;
    padding: 1.1rem 2rem;
    background: #F0EDE6;
    color: #030308;
    border: none;
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    letter-spacing: 0.2em;
    cursor: pointer;
    transition: background 0.2s, transform 0.1s, opacity 0.2s;
  }

  .pay-btn:hover:not(:disabled) {
    background: #ffffff;
    transform: translateY(-1px);
  }

  .pay-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .back-btn {
    background: none;
    border: 1px solid rgba(240, 237, 230, 0.2);
    color: rgba(240, 237, 230, 0.6);
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.15em;
    padding: 0.75rem 1.5rem;
    cursor: pointer;
    transition: border-color 0.15s, color 0.15s;
  }

  .back-btn:hover {
    border-color: rgba(240, 237, 230, 0.5);
    color: #F0EDE6;
  }
</style>
