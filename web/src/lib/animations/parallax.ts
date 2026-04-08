// web/src/lib/animations/parallax.ts
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

gsap.registerPlugin(ScrollTrigger);

/**
 * Apply a vertical parallax effect to `element`.
 * `speed` controls how many px the element moves per 100px of scroll.
 * Positive speed = moves up slower than the page (typical background parallax).
 */
export function applyParallax(element: HTMLElement, speed: number = 0.4): ScrollTrigger {
  return ScrollTrigger.create({
    trigger: element,
    start: 'top bottom',
    end: 'bottom top',
    scrub: true,
    onUpdate: (self) => {
      const progress = self.progress; // 0 → 1
      const yOffset = (progress - 0.5) * speed * window.innerHeight;
      gsap.set(element, { y: yOffset });
    },
  });
}
