# 2026-06-21: A paid order that couldn't ship

**What happened:** a real customer paid, and the order couldn't be fulfilled. Fixing it uncovered five bugs
in series, each hiding the next:

1. Stripe webhooks returned 403. The global proxy-auth middleware ran before the webhook route's exemption.
2. The shipping step ran in a fire-and-forget goroutine, and Fly scale-to-zero stopped the machine mid-task.
3. Order lookups crashed on unshipped orders, because `tracking`/`carrier`/`label_url` are NULL until shipped.
4. Shippo had no billing method, and the from-email was wrong.
5. USPS labels require `address_from` to carry both an email and a phone number.

Then the deeper problem: **the order didn't record what was bought.** The cart was deleted on payment, there
was no line-item storage, and the Stripe PaymentIntent held only an amount. The owner got a shipping label and
no packing list. Stock was also never decremented: the decrement function existed but was never called.

**Fixes:** `0e97384` (webhook exemption by path, `COALESCE` on order scans, Shippo from-email,
`min_machines_running = 1`, synchronous idempotent shipping) and `d4829d7` (`orders.line_items` persisted
at checkout, packing manifest in the owner email, per-line stock decrement once per order,
`TestWebhookFulfillmentE2E`).

**Rules produced:** STOP 18, STOP 19, IV-3, IV-4, IV-5, IV-6.

**Lesson:** unit tests were green the whole time. Only an end-to-end test asserting "the owner can actually
pack and ship this order" would have caught it.
