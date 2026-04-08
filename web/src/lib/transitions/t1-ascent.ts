// web/src/lib/transitions/t1-ascent.ts
import gsap from 'gsap';

export interface T1Elements {
  overlay: HTMLElement;
  flash: HTMLElement;
  horizon: HTMLElement;
  streakCanvas: HTMLCanvasElement;
  atmoLeft: HTMLElement;
  atmoRight: HTMLElement;
}

function drawStreaks(
  ctx: CanvasRenderingContext2D,
  w: number,
  h: number,
  speed: number
): void {
  ctx.clearRect(0, 0, w, h);
  if (speed < 0.02) return;

  const cx = w / 2;
  const cy = h / 2;
  const maxLen = Math.sqrt(cx * cx + cy * cy) * 1.2;
  const count = 220;

  for (let i = 0; i < count; i++) {
    const angle = (i / count) * Math.PI * 2;
    const lenMult = 0.5 + ((i * 73) % 100) / 200;
    const len = speed * maxLen * lenMult;
    const tail = speed * 0.22;

    const x1 = cx + Math.cos(angle) * len * tail;
    const y1 = cy + Math.sin(angle) * len * tail;
    const x2 = cx + Math.cos(angle) * len;
    const y2 = cy + Math.sin(angle) * len;

    const grad = ctx.createLinearGradient(x1, y1, x2, y2);
    const alpha = Math.min(speed * 0.75, 0.65);
    grad.addColorStop(0, `rgba(240,237,230,${alpha})`);
    grad.addColorStop(1, 'rgba(240,237,230,0)');

    ctx.beginPath();
    ctx.moveTo(x1, y1);
    ctx.lineTo(x2, y2);
    ctx.strokeStyle = grad;
    ctx.lineWidth = 0.4 + (i % 4) * 0.2;
    ctx.stroke();
  }
}

export function playT1Out(
  els: T1Elements,
  onMidpoint: () => void
): gsap.core.Timeline {
  const tl = gsap.timeline();

  // ── Setup ──
  gsap.set(els.overlay, { display: 'block', opacity: 1 });
  gsap.set([els.flash, els.atmoLeft, els.atmoRight, els.streakCanvas], { opacity: 0 });
  gsap.set([els.atmoLeft, els.atmoRight], { y: 0 });
  gsap.set(els.horizon, { y: 0, opacity: 1 });

  const w = window.innerWidth;
  const h = window.innerHeight;
  els.streakCanvas.width = w;
  els.streakCanvas.height = h;
  const ctx = els.streakCanvas.getContext('2d')!;

  const state = { speed: 0 };
  let rafId = 0;

  function loop() {
    drawStreaks(ctx, w, h, state.speed);
    rafId = requestAnimationFrame(loop);
  }

  tl
    .to(els.horizon, { y: '100vh', duration: 0.4, ease: 'power3.in' }, 0)
    .call(() => {
      gsap.set(els.streakCanvas, { opacity: 1 });
      rafId = requestAnimationFrame(loop);
    }, [], 0)
    .to(state, { speed: 1, duration: 0.5, ease: 'power3.in' }, 0)
    .to([els.atmoLeft, els.atmoRight], { opacity: 0.75, duration: 0.15, ease: 'power2.out' }, 0.3)
    .to([els.atmoLeft, els.atmoRight], { y: '-100vh', duration: 0.4, ease: 'power3.in' }, 0.45)
    .call(onMidpoint, [], 0.5)
    .to(els.flash, { opacity: 1, duration: 0.08, ease: 'power4.in' }, 0.8)
    .to(els.flash, { opacity: 0, duration: 0.22, ease: 'power2.out' }, 0.88)
    .to(state, { speed: 0, duration: 0.3, ease: 'power3.out' }, 0.9)
    .call(() => {
      cancelAnimationFrame(rafId);
      gsap.set(els.streakCanvas, { opacity: 0 });
    }, [], 1.2)
    .to({}, { duration: 0.2 }, 1.2);

  return tl;
}

export function playT1In(els: T1Elements, onComplete: () => void): gsap.core.Timeline {
  const tl = gsap.timeline({ onComplete });

  tl
    .to(els.overlay, { opacity: 0, duration: 0.3, ease: 'power2.out' })
    .call(() => { gsap.set(els.overlay, { display: 'none' }); });

  return tl;
}
