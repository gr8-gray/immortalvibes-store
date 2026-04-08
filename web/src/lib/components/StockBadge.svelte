<!-- web/src/lib/components/StockBadge.svelte -->
<script lang="ts">
  export let status: 'available' | 'sold_out' | 'coming_soon' = 'available';

  let email = '';
  let submitted = false;

  async function submitEmail() {
    if (!email || submitted) return;
    // POST to Go API /waitlist endpoint (wired in Plan 5 — for now, optimistic UI only)
    submitted = true;
  }
</script>

{#if status === 'sold_out'}
  <div class="badge badge--sold-out">
    <span class="badge-dot"></span>
    SOLD OUT
  </div>
{:else if status === 'coming_soon'}
  <div class="coming-soon">
    <div class="badge badge--coming-soon">
      <span class="badge-dot badge-dot--pulse"></span>
      COMING SOON
    </div>
    {#if !submitted}
      <form class="email-capture" on:submit|preventDefault={submitEmail}>
        <input
          type="email"
          bind:value={email}
          placeholder="Your email for launch notification"
          class="email-input"
          required
        />
        <button type="submit" class="email-submit" data-magnetic>NOTIFY ME</button>
      </form>
    {:else}
      <p class="submitted-msg">You're on the list.</p>
    {/if}
  </div>
{/if}

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.2em;
    padding: 0.4rem 1rem;
    border-radius: 2px;
  }

  .badge--sold-out {
    border: 1px solid rgba(240, 237, 230, 0.25);
    color: rgba(240, 237, 230, 0.4);
  }

  .badge--coming-soon {
    border: 1px solid rgba(240, 237, 230, 0.2);
    color: rgba(240, 237, 230, 0.5);
    margin-bottom: 1rem;
  }

  .badge-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: rgba(240, 237, 230, 0.4);
    display: inline-block;
  }

  .badge-dot--pulse {
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.3; }
    50% { opacity: 1; }
  }

  .coming-soon {
    display: flex;
    flex-direction: column;
  }

  .email-capture {
    display: flex;
    gap: 0.5rem;
    align-items: stretch;
  }

  .email-input {
    flex: 1;
    background: rgba(240, 237, 230, 0.04);
    border: 1px solid rgba(240, 237, 230, 0.15);
    color: #F0EDE6;
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    padding: 0.6rem 0.8rem;
    outline: none;
    transition: border-color 0.2s;
  }

  .email-input::placeholder {
    color: rgba(240, 237, 230, 0.3);
  }

  .email-input:focus {
    border-color: rgba(240, 237, 230, 0.4);
  }

  .email-submit {
    background: transparent;
    border: 1px solid rgba(240, 237, 230, 0.3);
    color: rgba(240, 237, 230, 0.7);
    font-family: 'Inter', sans-serif;
    font-size: 0.65rem;
    letter-spacing: 0.15em;
    padding: 0 1rem;
    cursor: none;
    transition: border-color 0.2s, color 0.2s;
    white-space: nowrap;
  }

  .email-submit:hover {
    border-color: rgba(240, 237, 230, 0.6);
    color: #F0EDE6;
  }

  .submitted-msg {
    font-family: 'Inter', sans-serif;
    font-size: 0.75rem;
    color: rgba(240, 237, 230, 0.5);
    margin: 0;
    letter-spacing: 0.05em;
  }
</style>
