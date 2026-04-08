// web/src/lib/transitions/t4-return.ts
import gsap from 'gsap';

export interface T4Elements {
  overlay: HTMLElement;
  spaceStars: HTMLElement;
  heat: HTMLElement;
  craft: HTMLElement;
  trail: HTMLElement;
  atmo: HTMLElement;
  cityline: HTMLElement;
}

export function playT4(
  els: T4Elements,
  onMidpoint: () => void,
  onComplete: () => void
): gsap.core.Timeline {
  const tl = gsap.timeline({ onComplete });

  gsap.set(els.overlay, { display: 'block', opacity: 1 });
  gsap.set([els.spaceStars, els.heat, els.craft, els.trail, els.atmo, els.cityline], { opacity: 0 });
  gsap.set(els.cityline, { y: 60 });
  gsap.set(els.craft, { y: -80 });

  tl
    .to(els.spaceStars, { opacity: 1, duration: 0.12 })
    .to(els.spaceStars, { opacity: 0, duration: 0.2 }, 0.18)
    .to(els.craft, { opacity: 0.7, y: 0, duration: 0.3, ease: 'power2.in' }, 0.1)
    .to(els.heat, { opacity: 1, duration: 0.2, ease: 'power2.in' }, 0.2)
    .to(els.trail, { opacity: 0.6, duration: 0.2 }, 0.2)
    .to(els.atmo, { opacity: 1, duration: 0.25 }, 0.28)
    .call(onMidpoint, [], 0.4)
    .to(els.heat, { opacity: 0, duration: 0.2 }, 0.45)
    .to(els.trail, { opacity: 0, duration: 0.15 }, 0.45)
    .to(els.craft, { opacity: 0, y: 40, duration: 0.2, ease: 'power2.out' }, 0.48)
    .to(els.cityline, { opacity: 1, y: 0, duration: 0.2, ease: 'power2.out' }, 0.52)
    .to(els.overlay, { opacity: 0, duration: 0.2 }, 0.65)
    .to(els.cityline, { opacity: 0, duration: 0.1 }, 0.7)
    .call(() => gsap.set(els.overlay, { display: 'none' }));

  return tl;
}
