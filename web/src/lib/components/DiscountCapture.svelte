<!-- web/src/lib/components/DiscountCapture.svelte -->
<!-- Email capture modal — opens from banner click. Scene bleeds through backdrop. -->
<script lang="ts">
  import { env } from '$env/dynamic/public';

  export let onSubscribed: () => void;
  export let onClose: () => void;

  let email = '';
  let submitting = false;
  let errorMsg = '';
  let code = '';

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    if (!email || submitting) return;
    submitting = true;
    errorMsg = '';

    try {
      const apiBase = env.PUBLIC_API_URL ?? '';
      const res = await fetch(`${apiBase}/api/subscribe`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || 'Subscription failed.');
      }

      const data = await res.json();
      code = data.code ?? 'VIBE10';
      onSubscribed();
    } catch (err) {
      errorMsg = err instanceof Error ? err.message : 'Something went wrong.';
    } finally {
      submitting = false;
    }
  }

  function copyCode() {
    navigator.clipboard.writeText(code);
  }

  function handleBackdrop(e: MouseEvent) {
    if (e.target === e.currentTarget) onClose();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="overlay" on:click={handleBackdrop} role="dialog" aria-modal="true" aria-label="Claim your discount">
  <div class="panel">

    <button class="close" on:click={onClose} aria-label="Close">✕</button>

    {#if !code}
      <!-- Email capture state -->
      <span class="kicker">TRANSMISSION INCOMING</span>
      <h2 class="headline">Rise Beyond<br>10%</h2>
      <p class="sub">Enter your frequency to unlock 10% off your first order. Code delivered to your inbox.</p>

      <form on:submit={handleSubmit} class="form">
        <input
          type="email"
          bind:value={email}
          placeholder="your@email.com"
          class="email-input"
          required
          autocomplete="email"
        />
        <button type="submit" class="submit-btn" disabled={submitting || !email}>
          {submitting ? 'TRANSMITTING…' : 'CLAIM 10% OFF →'}
        </button>
      </form>

      {#if errorMsg}
        <p class="error">{errorMsg}</p>
      {/if}

      <p class="fine">One use per customer. New customers only. No spam, ever.</p>

    {:else}
      <!-- Code revealed state -->
      <span class="kicker">TRANSMISSION RECEIVED</span>
      <h2 class="headline">Your Code</h2>
      <p class="sub">Apply at checkout. Valid on your first order.</p>

      <button class="code-block" on:click={copyCode} title="Click to copy">
        <span class="code-text">{code}</span>
        <span class="copy-hint">CLICK TO COPY</span>
      </button>

      <p class="sent-note">Code also sent to {email}</p>

      <button class="continue-btn" on:click={onClose}>CONTINUE SHOPPING →</button>
    {/if}

  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    z-index: 200;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(3, 3, 8, 0.72);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    animation: fadeIn 0.3s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to   { opacity: 1; }
  }

  .panel {
    position: relative;
    width: min(480px, calc(100vw - 2rem));
    background: rgba(8, 8, 18, 0.94);
    border: 1px solid rgba(200, 146, 42, 0.22);
    padding: 3.5rem 3rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.25rem;
    text-align: center;
    animation: slideUp 0.35s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideUp {
    from { opacity: 0; transform: translateY(20px); }
    to   { opacity: 1; transform: translateY(0); }
  }

  .close {
    position: absolute;
    top: 1rem;
    right: 1.25rem;
    background: none;
    border: none;
    color: rgba(240, 237, 230, 0.3);
    font-size: 0.8rem;
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    transition: color 0.15s;
    line-height: 1;
  }

  .close:hover { color: rgba(240, 237, 230, 0.7); }

  .kicker {
    font-family: 'Inter', sans-serif;
    font-size: 0.48rem;
    letter-spacing: 0.45em;
    color: rgba(200, 146, 42, 0.7);
    text-transform: uppercase;
  }

  .headline {
    font-family: 'GodOfWar', 'Gods of War', serif;
    font-size: clamp(2rem, 6vw, 3.2rem);
    color: #F0EDE6;
    line-height: 1.1;
    margin: 0;
  }

  .sub {
    font-family: 'Inter', sans-serif;
    font-size: 0.8rem;
    line-height: 1.65;
    color: rgba(240, 237, 230, 0.5);
    max-width: 30ch;
    margin: 0;
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    width: 100%;
    margin-top: 0.5rem;
  }

  .email-input {
    width: 100%;
    background: rgba(240, 237, 230, 0.04);
    border: 1px solid rgba(240, 237, 230, 0.15);
    color: #F0EDE6;
    font-family: 'Inter', sans-serif;
    font-size: 0.85rem;
    padding: 0.9rem 1.1rem;
    outline: none;
    transition: border-color 0.2s;
    text-align: center;
  }

  .email-input:focus { border-color: rgba(200, 146, 42, 0.5); }
  .email-input::placeholder { color: rgba(240, 237, 230, 0.2); }

  .submit-btn {
    width: 100%;
    padding: 1rem 2rem;
    background: #C8922A;
    color: #030308;
    border: none;
    font-family: 'Inter', sans-serif;
    font-size: 0.62rem;
    letter-spacing: 0.22em;
    cursor: pointer;
    transition: background 0.2s, transform 0.1s, opacity 0.2s;
    text-transform: uppercase;
  }

  .submit-btn:hover:not(:disabled) {
    background: #d9a033;
    transform: translateY(-1px);
  }

  .submit-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .error {
    font-family: 'Inter', sans-serif;
    font-size: 0.7rem;
    color: rgba(200, 80, 80, 0.85);
    margin: 0;
  }

  .fine {
    font-family: 'Inter', sans-serif;
    font-size: 0.48rem;
    letter-spacing: 0.12em;
    color: rgba(240, 237, 230, 0.2);
    margin: 0;
  }

  /* ── Code revealed state ── */
  .code-block {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    background: rgba(200, 146, 42, 0.08);
    border: 1px solid rgba(200, 146, 42, 0.4);
    padding: 1.4rem 3rem;
    cursor: pointer;
    transition: background 0.2s, border-color 0.2s;
    width: 100%;
    margin-top: 0.5rem;
  }

  .code-block:hover {
    background: rgba(200, 146, 42, 0.14);
    border-color: rgba(200, 146, 42, 0.7);
  }

  .code-text {
    font-family: 'GodOfWar', 'Gods of War', serif;
    font-size: 2rem;
    color: #C8922A;
    letter-spacing: 0.12em;
  }

  .copy-hint {
    font-family: 'Inter', sans-serif;
    font-size: 0.45rem;
    letter-spacing: 0.3em;
    color: rgba(200, 146, 42, 0.5);
    text-transform: uppercase;
  }

  .sent-note {
    font-family: 'Inter', sans-serif;
    font-size: 0.62rem;
    color: rgba(240, 237, 230, 0.3);
    margin: 0;
    letter-spacing: 0.06em;
  }

  .continue-btn {
    background: none;
    border: 1px solid rgba(240, 237, 230, 0.2);
    color: rgba(240, 237, 230, 0.55);
    font-family: 'Inter', sans-serif;
    font-size: 0.55rem;
    letter-spacing: 0.22em;
    padding: 0.75rem 1.8rem;
    cursor: pointer;
    transition: border-color 0.2s, color 0.2s;
    text-transform: uppercase;
    margin-top: 0.5rem;
  }

  .continue-btn:hover {
    border-color: rgba(240, 237, 230, 0.5);
    color: #F0EDE6;
  }
</style>
