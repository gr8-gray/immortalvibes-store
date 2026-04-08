// web/src/lib/transitions/t1-ascent.ts
import gsap from 'gsap';

export interface T1Elements {
  overlay: HTMLElement;
  flash: HTMLElement;
  blast: HTMLElement;
  shockwave: HTMLElement;
  streaks: HTMLElement[];
  atmo: HTMLElement;
  starsCanvas: HTMLCanvasElement;
  cityline: HTMLElement;
}

function drawVoidStars(canvas: HTMLCanvasElement): void {
  const ctx = canvas.getContext('2d');
  if (!ctx) return;
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  for (let i = 0; i < 500; i++) {
    const x = Math.random() * canvas.width;
    const y = Math.random() * canvas.height;
    const r = Math.random() < 0.05 ? 2 : Math.random() < 0.2 ? 1.5 : 1;
    const alpha = 0.4 + Math.random() * 0.6;
    ctx.beginPath();
    ctx.arc(x, y, r, 0, Math.PI * 2);
    ctx.fillStyle = `rgba(240,237,230,${alpha})`;
    ctx.fill();
  }

  const grd = ctx.createRadialGradient(
    canvas.width * 0.5, canvas.height * 0.5, 0,
    canvas.width * 0.5, canvas.height * 0.5, canvas.width * 0.4
  );
  grd.addColorStop(0, 'rgba(240,237,230,0.04)');
  grd.addColorStop(1, 'transparent');
  ctx.fillStyle = grd;
  ctx.fillRect(0, 0, canvas.width, canvas.height);
}

export function playT1Out(
  els: T1Elements,
  onMidpoint: () => void
): gsap.core.Timeline {
  const tl = gsap.timeline();

  gsap.set(els.overlay, { display: 'block', opacity: 1 });
  gsap.set([els.flash, els.blast, els.shockwave, ...els.streaks, els.atmo, els.starsCanvas], {
    opacity: 0,
  });
  gsap.set(els.starsCanvas, { display: 'none' });

  tl
    .to(els.flash, { opacity: 1, duration: 0.08, ease: 'power3.in' })
    .to(els.flash, { opacity: 0, duration: 0.07, ease: 'power2.out' })
    .fromTo(els.blast,
      { scale: 0, opacity: 0, transformOrigin: '50% 100%' },
      { scale: 1, opacity: 1, duration: 0.15, ease: 'power3.out' }, 0
    )
    .fromTo(els.shockwave,
      { scale: 0, opacity: 0.6, transformOrigin: '50% 100%' },
      { scale: 4, opacity: 0, duration: 0.3, ease: 'power2.out' }, 0
    )
    .to(els.cityline, { x: -4, duration: 0.04 }, 0.04)
    .to(els.cityline, { x: 4, duration: 0.04 }, 0.08)
    .to(els.cityline, { x: 0, duration: 0.04 }, 0.12)
    .to(els.atmo, { opacity: 1, duration: 0.1 }, 0.15)
    .to(els.blast, { opacity: 0, y: -80, duration: 0.3, ease: 'power2.in' }, 0.15);

  els.streaks.forEach((streak, i) => {
    tl.fromTo(streak,
      { scaleY: 0, opacity: 0, transformOrigin: '50% 100%' },
      { scaleY: 1, opacity: 0.5 + Math.random() * 0.4, duration: 0.25, ease: 'power2.out' },
      0.15 + i * 0.02
    );
  });

  tl
    .to(els.atmo, { opacity: 0, duration: 0.15 }, 0.38)
    .to(els.streaks, { opacity: 0, duration: 0.1 }, 0.38)
    .to(els.cityline, { opacity: 0, y: 60, duration: 0.2 }, 0.2)
    .call(onMidpoint, [], 0.45)
    .call(() => {
      drawVoidStars(els.starsCanvas);
      gsap.set(els.starsCanvas, { display: 'block', opacity: 0 });
    }, [], 0.45)
    .to(els.starsCanvas, { opacity: 1, duration: 0.15, ease: 'power1.in' }, 0.45)
    .to({}, { duration: 0.25 }, 0.6);

  return tl;
}

export function playT1In(els: T1Elements, onComplete: () => void): gsap.core.Timeline {
  const tl = gsap.timeline({ onComplete });

  tl
    .to(els.starsCanvas, { opacity: 0, duration: 0.2, ease: 'power1.out' })
    .to(els.overlay, { opacity: 0, duration: 0.15, ease: 'power1.out' }, 0.1)
    .call(() => {
      gsap.set(els.overlay, { display: 'none' });
    });

  return tl;
}
