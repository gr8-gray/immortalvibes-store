<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { loadStripe } from '@stripe/stripe-js';
  import { env } from '$env/dynamic/public';
  import { cart } from '$lib/stores/cart';
  import { createCheckout } from '$lib/api';
  import type { Stripe, StripeElements } from '@stripe/stripe-js';
  import type { ShippingAddress } from '$lib/api';

  let stripe: Stripe | null = null;
  let elements: StripeElements | null = null;
  let paymentElement: ReturnType<StripeElements['create']> | null = null;
  let mountNode: HTMLDivElement;

  let orderId = '';
  let email = '';
  let shippingName = '';
  let line1 = '';
  let line2 = '';
  let city = '';
  let addrState = '';
  let postalCode = '';
  let country = 'US';
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
    }

    if (!$cart.id) {
      errorMsg = 'No active cart. Add items before checking out.';
    }

    loading = false;
  });

  async function initPayment() {
    if (!stripe || !email || submitting) return;
    if (!shippingName || !line1 || !city || !addrState || !postalCode || !country) {
      errorMsg = 'Please complete your shipping address.';
      return;
    }
    submitting = true;
    errorMsg = '';
    const address: ShippingAddress = {
      shipping_name: shippingName,
      line1,
      line2: line2 || undefined,
      city,
      state: addrState,
      postal_code: postalCode,
      country,
    };
    try {
      const session = await createCheckout($cart.id!, email, address);
      const clientSecret = session.client_secret;
      orderId = session.order_id;
      elements = stripe!.elements({ clientSecret });
      paymentElement = elements.create('payment');
      paymentElement.mount(mountNode);
    } catch (e: unknown) {
      errorMsg = e instanceof Error ? e.message : 'Failed to initialise checkout.';
    } finally {
      submitting = false;
    }
  }

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

    <!-- Email step -->
    {#if !elements}
      <div class="email-step">
        <label class="field-label" for="email">EMAIL</label>
        <input
          id="email"
          type="email"
          bind:value={email}
          placeholder="your@email.com"
          class="email-input"
          required
        />

        <p class="field-label" style="margin-top:1.25rem">SHIPPING ADDRESS</p>

        <input
          type="text"
          bind:value={shippingName}
          placeholder="Full name"
          class="email-input"
          required
        />
        <input
          type="text"
          bind:value={line1}
          placeholder="Address line 1"
          class="email-input"
          required
        />
        <input
          type="text"
          bind:value={line2}
          placeholder="Apartment, suite, etc. (optional)"
          class="email-input"
        />
        <div class="addr-row">
          <input
            type="text"
            bind:value={city}
            placeholder="City"
            class="email-input"
            required
          />
          <input
            type="text"
            bind:value={addrState}
            placeholder="State"
            class="email-input addr-state"
            required
          />
        </div>
        <div class="addr-row">
          <input
            type="text"
            bind:value={postalCode}
            placeholder="ZIP / Postal code"
            class="email-input"
            required
          />
          <input
            type="text"
            bind:value={country}
            placeholder="Country"
            class="email-input addr-country"
            required
          />
        </div>

        <button class="pay-btn" on:click={initPayment} disabled={submitting || !email}>
          {submitting ? 'PREPARING…' : 'CONTINUE TO PAYMENT'}
        </button>
      </div>
    {/if}

    <!-- Stripe Payment Element -->
    <form on:submit={handleSubmit} class="payment-form">
      <div bind:this={mountNode} class="payment-element-mount"></div>

      {#if errorMsg}
        <p class="error-msg">{errorMsg}</p>
      {/if}

      {#if elements}
        <button type="submit" class="pay-btn" disabled={submitting}>
          {submitting ? 'PROCESSING…' : 'PAY NOW'}
        </button>
      {/if}
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

  .email-step {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .field-label {
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.2em;
    color: rgba(240, 237, 230, 0.4);
  }

  .email-input {
    width: 100%;
    background: rgba(240, 237, 230, 0.04);
    border: 1px solid rgba(240, 237, 230, 0.15);
    color: #F0EDE6;
    font-family: 'Inter', sans-serif;
    font-size: 0.85rem;
    padding: 0.8rem 1rem;
    outline: none;
    transition: border-color 0.2s;
  }

  .email-input:focus { border-color: rgba(240, 237, 230, 0.4); }
  .email-input::placeholder { color: rgba(240, 237, 230, 0.25); }

  .addr-row {
    display: flex;
    gap: 0.75rem;
  }

  .addr-row .email-input { flex: 1; }
  .addr-row .addr-state  { flex: 0 0 5rem; }
  .addr-row .addr-country { flex: 0 0 5rem; }

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
