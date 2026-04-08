// web/src/lib/transitions/t3-ring.ts
import gsap from 'gsap';

export interface T3Elements {
  overlay: HTMLElement;
  rings: HTMLElement[];
  core: HTMLElement;
  rays: HTMLElement[];
}

export function playT3(
  els: T3Elements,
  mainContent: HTMLElement,
  onMidpoint: () => void,
  onComplete: () => void
): gsap.core.Timeline {
  const tl = gsap.timeline({ onComplete });

  gsap.set(els.overlay, { display: 'block', opacity: 1 });
  gsap.set([...els.rings, els.core, ...els.rays], { opacity: 0, scale: 0 });

  tl
    .to(mainContent, {
      scale: 0.92,
      opacity: 0,
      duration: 0.25,
      ease: 'power2.in',
    })
    .to(els.rings, {
      scale: 1,
      opacity: (i: number) => 0.15 + i * 0.15,
      duration: 0.2,
      stagger: -0.04,
      ease: 'power2.out',
    }, 0.1)
    .to(els.rays, { scale: 1, opacity: 0.4, duration: 0.15, stagger: 0.02 }, 0.15)
    .to(els.core, { scale: 1, opacity: 1, duration: 0.1 }, 0.2)
    .call(onMidpoint, [], 0.3)
    .to([...els.rings, els.core, ...els.rays], {
      scale: 4,
      opacity: 0,
      duration: 0.25,
      stagger: 0.03,
      ease: 'power3.out',
    }, 0.35)
    .fromTo(mainContent,
      { scale: 1.08, opacity: 0 },
      { scale: 1, opacity: 1, duration: 0.22, ease: 'power2.out' },
      0.38
    )
    .to(els.overlay, { opacity: 0, duration: 0.1 }, 0.55)
    .call(() => gsap.set(els.overlay, { display: 'none' }));

  return tl;
}
