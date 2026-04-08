// web/src/lib/animations/reveal.ts
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

gsap.registerPlugin(ScrollTrigger);

/**
 * Fade-up + scale reveal triggered when `element` enters viewport.
 * Returns the ScrollTrigger so the caller can kill it on destroy.
 */
export function revealOnScroll(element: HTMLElement, delay: number = 0): ScrollTrigger {
  gsap.set(element, { opacity: 0, y: 30, scale: 0.95 });

  return ScrollTrigger.create({
    trigger: element,
    start: 'top 85%',
    once: true,
    onEnter: () => {
      gsap.to(element, {
        opacity: 1,
        y: 0,
        scale: 1,
        duration: 0.8,
        delay,
        ease: 'power2.out',
      });
    },
  });
}
