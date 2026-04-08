// web/src/lib/transitions/t2-hyperspace.ts
import gsap from 'gsap';

export interface T2Elements {
  overlay: HTMLElement;
  streaks: HTMLElement[];
  flash: HTMLElement;
  mist: HTMLElement;
}

export function playT2(
  els: T2Elements,
  clickX: number,
  clickY: number,
  accentColor: string,
  onMidpoint: () => void,
  onComplete: () => void
): gsap.core.Timeline {
  const tl = gsap.timeline({ onComplete });

  const originX = `${clickX * 100}%`;
  const originY = `${clickY * 100}%`;

  gsap.set(els.overlay, { display: 'block', opacity: 1 });
  gsap.set([...els.streaks, els.flash, els.mist], { opacity: 0 });

  // Build accent color for mist — strip opacity to 0.12
  const mistColor = accentColor.startsWith('rgba')
    ? accentColor.replace(/[\d.]+\)$/, '0.12)')
    : `${accentColor}1e`;

  gsap.set(els.mist, {
    background: `radial-gradient(ellipse at ${originX} ${originY}, ${mistColor} 0%, transparent 60%)`,
  });

  els.streaks.forEach((streak, i) => {
    const angle = (i / els.streaks.length) * 360;
    const len = 30 + (i % 5) * 8;
    gsap.set(streak, {
      transformOrigin: '0% 50%',
      rotation: angle,
      left: originX,
      top: originY,
      width: `${len}%`,
    });
  });

  tl
    .to(els.mist, { opacity: 1, duration: 0.1 })
    .to(els.streaks, {
      opacity: 0.85,
      scaleX: 1,
      duration: 0.18,
      stagger: 0.008,
      ease: 'power2.in',
    }, 0.05)
    .to(els.flash, { opacity: 1, duration: 0.06, ease: 'power3.in' }, 0.28)
    .call(onMidpoint, [], 0.3)
    .to(els.flash, { opacity: 0, duration: 0.08, ease: 'power2.out' }, 0.3)
    .to(els.streaks, { opacity: 0, scaleX: 0, duration: 0.15, ease: 'power2.out' }, 0.3)
    .to(els.mist, { opacity: 0, duration: 0.1 }, 0.36)
    .to(els.overlay, { opacity: 0, duration: 0.1 }, 0.42)
    .call(() => gsap.set(els.overlay, { display: 'none' }));

  return tl;
}
