<!-- web/src/lib/components/MissionPlanet.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import * as THREE from 'three';

  export interface ClusterItem {
    url: string;
    scale?: number;          // size multiplier (default productScale)
    offsetX?: number;        // horizontal offset (Three.js units)
    offsetY?: number;        // vertical offset
    opacity?: number;        // 0–1 (default productOpacity)
    blending?: 'normal' | 'additive';
  }

  export let planetType: 'earth' | 'leo' | 'lunar' | 'nebula' = 'earth';
  export let photoUrl: string   = '';
  export let productUrl: string   = '';   // single sprite path (legacy)
  export let products: ClusterItem[] = [];  // multi-sprite cluster (overrides productUrl when non-empty)
  // Auto-rotating carousel. Strings or objects with { url, offsetY?, scale? }
  // for per-slot positioning when source PNGs aren't perfectly centered/sized.
  export type CarouselSlot = string | { url: string; offsetY?: number; scale?: number };
  export let carousel: CarouselSlot[] = [];
  export let carouselInterval: number = 3500; // ms between transitions
  export let carouselFade: number     = 850;  // ms crossfade duration
  export let productScale: number   = 1.0;   // default per-sprite size
  export let productOffsetY: number  = 0.0;          // default vertical shift
  export let productOpacity: number  = 0.72;         // default opacity
  export let productBlending: 'normal' | 'additive' = 'normal';
  export let glowColor: string  = '#4FC3F7';
  export let rotationSpeed: number = 0.0015;
  export let axialTilt: number     = 0.2;

  let canvas: HTMLCanvasElement;
  let rafId: number;
  let renderer: THREE.WebGLRenderer;

  const SUN = new THREE.Vector3(0.596, 0.298, 0.745);

  // ── Shared 3D noise (no UV seams) ─────────────────────────────────────────
  const NOISE3_GLSL = /* glsl */`
    float h3(vec3 p){p=fract(p*vec3(127.1,311.7,74.7));p+=dot(p,p.yzx+19.19);return fract((p.x+p.y)*p.z);}
    float n3(vec3 p){
      vec3 i=floor(p),f=fract(p);f=f*f*(3.0-2.0*f);
      return mix(mix(mix(h3(i),h3(i+vec3(1,0,0)),f.x),mix(h3(i+vec3(0,1,0)),h3(i+vec3(1,1,0)),f.x),f.y),
                 mix(mix(h3(i+vec3(0,0,1)),h3(i+vec3(1,0,1)),f.x),mix(h3(i+vec3(0,1,1)),h3(i+vec3(1,1,1)),f.x),f.y),f.z);
    }
    float fbm(vec3 p){float v=0.0,a=0.5;
      v+=a*n3(p);p*=2.1;a*=0.5;v+=a*n3(p);p*=2.1;a*=0.5;
      v+=a*n3(p);p*=2.1;a*=0.5;v+=a*n3(p);p*=2.1;a*=0.5;
      v+=a*n3(p);p*=2.1;a*=0.5;v+=a*n3(p);return v;}
  `;

  const SHARED_VERT = /* glsl */`
    varying vec3 vNormal,vPos;
    void main(){vNormal=normalize(normalMatrix*normal);vPos=position;
      gl_Position=projectionMatrix*modelViewMatrix*vec4(position,1.0);}
  `;

  // ── LEO — sci-fi alien ocean ──────────────────────────────────────────────
  const leoFrag = /* glsl */`
    varying vec3 vNormal,vPos; uniform vec3 sunDir; ${NOISE3_GLSL}
    void main(){
      float cloud=fbm(vPos*3.2),cloud2=fbm(vPos*6.5+vec3(4.1,2.3,7.8)),land=fbm(vPos*2.1+vec3(9.2,1.7,3.4));
      vec3 ocean=mix(vec3(0.0,0.12,0.35),vec3(0.0,0.30,0.65),cloud);
      vec3 terrain=mix(vec3(0.02,0.18,0.22),vec3(0.04,0.28,0.28),cloud2);
      float isLand=smoothstep(0.54,0.62,land);
      vec3 surface=mix(ocean,terrain,isLand*0.7);
      float cloudMask=smoothstep(0.48,0.68,cloud+cloud2*0.4);
      surface=mix(surface,vec3(0.88,0.93,1.0),cloudMask*0.85);
      float diff=dot(vNormal,sunDir),lit=smoothstep(-0.12,0.18,diff);
      vec3 col=surface*(lit*0.88+0.04);
      vec3 refl=reflect(-sunDir,vNormal);
      float spec=pow(max(0.0,dot(refl,vec3(0,0,1))),28.0);
      col+=vec3(0.6,0.85,1.0)*spec*(1.0-isLand)*lit*0.6;
      float limb=pow(max(0.0,1.0-abs(dot(vNormal,vec3(0,0,1)))),2.0);
      col+=vec3(0.1,0.55,1.0)*limb*0.55+vec3(0.05,0.3,0.7)*limb*limb*0.3;
      gl_FragColor=vec4(clamp(col,0.0,1.0),1.0);}
  `;

  // ── Lunar ─────────────────────────────────────────────────────────────────
  const lunarFrag = /* glsl */`
    varying vec3 vNormal,vPos; uniform vec3 sunDir; ${NOISE3_GLSL}
    void main(){
      float n1=fbm(vPos*6.0),n2=fbm(vPos*14.0+vec3(5.2,3.1,8.7)),n3v=fbm(vPos*3.0+vec3(1.4,9.2,2.6));
      vec3 base=mix(vec3(0.52,0.50,0.48),vec3(0.84,0.83,0.80),n1);
      float maria=smoothstep(0.40,0.52,n3v);
      base=mix(base*0.42,base,maria);base+=(n2-0.5)*0.055;
      float cA=abs(fbm(vPos*20.0)-0.5),cB=abs(fbm(vPos*10.0+vec3(3.3))-0.5);
      base+=smoothstep(0.055,0.0,cA)*0.22+smoothstep(0.07,0.01,cB)*0.10;
      base=clamp(base,0.0,1.0);
      float diff=dot(vNormal,sunDir),lit=smoothstep(-0.06,0.14,diff);
      vec3 col=base*(lit*0.93+0.025);
      float rimLit=pow(max(0.0,1.0-abs(dot(vNormal,vec3(0,0,1)))),3.0);
      col+=vec3(0.55,0.65,0.9)*rimLit*lit*0.15;
      gl_FragColor=vec4(clamp(col,0.0,1.0),1.0);}
  `;

  // ── Nebula ────────────────────────────────────────────────────────────────
  const nebulaFrag = /* glsl */`
    varying vec3 vNormal,vPos; uniform vec3 sunDir; uniform float time; ${NOISE3_GLSL}
    void main(){
      float c=cos(time*0.012),s=sin(time*0.012);
      vec3 rp=vec3(c*vPos.x+s*vPos.z,vPos.y,-s*vPos.x+c*vPos.z);
      vec3 warp=rp+vec3(fbm(rp*3.0)*0.22,fbm(rp*2.6+vec3(5.2,1.3,0.0))*0.10,0.0);
      float t1=fbm(warp*4.8),t2=fbm(warp*10.0+vec3(3.1,8.7,2.2));
      float band=fract(vPos.y*2.6+t1*0.5);
      vec3 c0=vec3(0.10,0.03,0.40),c1=vec3(0.28,0.08,0.72),c2=vec3(0.06,0.02,0.28),c3=vec3(0.48,0.18,0.92),bandCol;
      if(band<0.25)bandCol=mix(c0,c1,band*4.0);
      else if(band<0.50)bandCol=mix(c1,c2,(band-0.25)*4.0);
      else if(band<0.75)bandCol=mix(c2,c3,(band-0.50)*4.0);
      else bandCol=mix(c3,c0,(band-0.75)*4.0);
      bandCol+=(t2-0.5)*0.09;
      vec3 sc=normalize(vec3(0.6,0.1,0.8));
      float sdst=acos(clamp(dot(normalize(vPos),sc),-1.0,1.0)),storm=smoothstep(0.28,0.10,sdst);
      bandCol=mix(bandCol,vec3(0.55,0.20,0.95),storm*0.55)+vec3(storm*0.06);
      bandCol=clamp(bandCol,0.0,1.0);
      float diff=max(0.0,dot(vNormal,sunDir));
      vec3 col=bandCol*(diff*0.78+0.22);
      float limb=pow(max(0.0,1.0-dot(vNormal,vec3(0,0,1))),2.2);
      col+=vec3(0.35,0.10,0.85)*limb*0.50;
      gl_FragColor=vec4(clamp(col,0.0,1.0),1.0);}
  `;

  // ── Texture builder — bake radial fade once per image ─────────────────────
  function buildFadedTexture(img: HTMLImageElement): THREE.CanvasTexture {
    const cw = img.naturalWidth  || img.width;
    const ch = img.naturalHeight || img.height;
    const c = document.createElement('canvas');
    c.width = cw; c.height = ch;
    const ctx = c.getContext('2d')!;
    ctx.drawImage(img, 0, 0);
    const cx = cw / 2, cy = ch / 2;
    const rx = cx * 0.88, ry = cy * 0.88;
    const grad = ctx.createRadialGradient(cx, cy, Math.min(rx, ry) * 0.15, cx, cy, Math.max(rx, ry));
    grad.addColorStop(0.0, 'rgba(0,0,0,1)');
    grad.addColorStop(0.6, 'rgba(0,0,0,0.9)');
    grad.addColorStop(0.85, 'rgba(0,0,0,0.4)');
    grad.addColorStop(1.0, 'rgba(0,0,0,0)');
    ctx.globalCompositeOperation = 'destination-in';
    ctx.fillStyle = grad;
    ctx.fillRect(0, 0, cw, ch);
    const t = new THREE.CanvasTexture(c);
    t.colorSpace = THREE.SRGBColorSpace;
    return t;
  }

  // ── Carousel loader — preload N textures, run a crossfade timer ──────────
  // Returns a cleanup function (clear interval, dispose textures).
  function loadCarousel(rawSlots: CarouselSlot[], scene: THREE.Scene): () => void {
    const loader = new THREE.TextureLoader();
    const items = rawSlots.map((it) =>
      typeof it === 'string'
        ? { url: it, offsetY: 0, scale: 1 }
        : { url: it.url, offsetY: it.offsetY ?? 0, scale: it.scale ?? 1 },
    );
    const slots: { tex: THREE.CanvasTexture; w: number; h: number; offsetY: number; scale: number }[] = [];
    let intervalId: number | null = null;

    // Two sprites — one shown, one hidden, swap roles each tick.
    let spriteA: THREE.Sprite | null = null;
    let spriteB: THREE.Sprite | null = null;
    let activeIsA = true;
    let nextIndex = 0;

    function makeSprite(): THREE.Sprite {
      const mat = new THREE.SpriteMaterial({
        transparent: true,
        opacity:     0,
        blending:    productBlending === 'additive' ? THREE.AdditiveBlending : THREE.NormalBlending,
        depthWrite:  false,
        depthTest:   false,
      });
      const s = new THREE.Sprite(mat);
      s.position.set(0, productOffsetY, 1.12);
      scene.add(s);
      return s;
    }

    function applySlot(sprite: THREE.Sprite, slot: typeof slots[0]) {
      const aspect = slot.w / slot.h;
      const spriteH = 1.5 * productScale * slot.scale;
      const spriteW = spriteH * aspect;
      sprite.material.map = slot.tex;
      sprite.material.needsUpdate = true;
      sprite.scale.set(spriteW, spriteH, 1);
      sprite.position.y = productOffsetY + slot.offsetY;
    }

    function fade(sprite: THREE.Sprite, from: number, to: number, durationMs: number) {
      const start = performance.now();
      function step(now: number) {
        const t = Math.min(1, (now - start) / durationMs);
        const eased = t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
        sprite.material.opacity = from + (to - from) * eased;
        if (t < 1) requestAnimationFrame(step);
      }
      requestAnimationFrame(step);
    }

    function tick() {
      if (slots.length === 0) return;
      const incoming = activeIsA ? spriteB : spriteA;
      const outgoing = activeIsA ? spriteA : spriteB;
      if (!incoming || !outgoing) return;
      applySlot(incoming, slots[nextIndex]);
      nextIndex = (nextIndex + 1) % slots.length;
      fade(incoming, 0, productOpacity, carouselFade);
      fade(outgoing, productOpacity, 0, carouselFade);
      activeIsA = !activeIsA;
    }

    // Preload all textures, then start.
    let loaded = 0;
    items.forEach((it, i) => {
      loader.load(it.url, (tex) => {
        const img = tex.image as HTMLImageElement;
        slots[i] = {
          tex: buildFadedTexture(img),
          w: img.naturalWidth || img.width,
          h: img.naturalHeight || img.height,
          offsetY: it.offsetY,
          scale: it.scale,
        };
        loaded++;
        if (loaded === items.length) start();
      });
    });

    function start() {
      spriteA = makeSprite();
      spriteB = makeSprite();
      // Show first slot immediately, fade in.
      applySlot(spriteA, slots[0]);
      fade(spriteA, 0, productOpacity, carouselFade);
      nextIndex = slots.length > 1 ? 1 : 0;
      if (slots.length > 1) {
        intervalId = window.setInterval(tick, carouselInterval);
      }
    }

    return () => {
      if (intervalId !== null) clearInterval(intervalId);
      slots.forEach(s => s.tex.dispose());
    };
  }

  // ── Product sprite — apply radial alpha fade via canvas ───────────────────
  function loadProductSprite(item: ClusterItem, scene: THREE.Scene) {
    const url      = item.url;
    const scale    = item.scale    ?? productScale;
    const offX     = item.offsetX  ?? 0;
    const offY     = item.offsetY  ?? productOffsetY;
    const op       = item.opacity  ?? productOpacity;
    const blend    = item.blending ?? productBlending;

    const loader = new THREE.TextureLoader();
    loader.load(url, (tex) => {
      const img = tex.image as HTMLImageElement;
      const cw = img.naturalWidth  || img.width;
      const ch = img.naturalHeight || img.height;

      // Bake a soft radial vignette into the texture
      const c = document.createElement('canvas');
      c.width = cw; c.height = ch;
      const ctx = c.getContext('2d')!;
      ctx.drawImage(img, 0, 0);

      // Radial alpha fade — centre full, edges dissolve
      const cx = cw / 2, cy = ch / 2;
      const rx = cx * 0.88, ry = cy * 0.88;
      const grad = ctx.createRadialGradient(cx, cy, Math.min(rx, ry) * 0.15, cx, cy, Math.max(rx, ry));
      grad.addColorStop(0.0, 'rgba(0,0,0,1)');
      grad.addColorStop(0.6, 'rgba(0,0,0,0.9)');
      grad.addColorStop(0.85, 'rgba(0,0,0,0.4)');
      grad.addColorStop(1.0, 'rgba(0,0,0,0)');
      ctx.globalCompositeOperation = 'destination-in';
      ctx.fillStyle = grad;
      ctx.fillRect(0, 0, cw, ch);

      const fadedTex = new THREE.CanvasTexture(c);
      fadedTex.colorSpace = THREE.SRGBColorSpace;

      const aspect = cw / ch;
      const spriteH = 1.5 * scale;
      const spriteW = spriteH * aspect;

      const spriteMat = new THREE.SpriteMaterial({
        map:         fadedTex,
        transparent: true,
        opacity:     op,
        blending:    blend === 'additive' ? THREE.AdditiveBlending : THREE.NormalBlending,
        depthWrite:  false,
        depthTest:   false,
      });
      const sprite = new THREE.Sprite(spriteMat);
      sprite.scale.set(spriteW, spriteH, 1);
      sprite.position.set(offX, offY, 1.12);
      scene.add(sprite);
    });
  }

  onMount(() => {
    if (!browser) return;

    const W = canvas.clientWidth  || canvas.offsetWidth  || 400;
    const H = canvas.clientHeight || canvas.offsetHeight || 400;

    const scene  = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(42, W/H, 0.1, 100);
    camera.position.z = 2.6;

    renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
    renderer.setSize(W, H);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.outputColorSpace = THREE.SRGBColorSpace;
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1.15;

    const geo = new THREE.SphereGeometry(1, 80, 80);
    let mat: THREE.Material;
    let timeUniform: { value: number } | null = null;

    if (planetType === 'leo') {
      mat = new THREE.ShaderMaterial({
        vertexShader: SHARED_VERT, fragmentShader: leoFrag,
        uniforms: { sunDir: { value: SUN } },
      });
    } else if (planetType === 'lunar') {
      mat = new THREE.ShaderMaterial({
        vertexShader: SHARED_VERT, fragmentShader: lunarFrag,
        uniforms: { sunDir: { value: SUN } },
      });
    } else if (planetType === 'nebula') {
      timeUniform = { value: 0 };
      mat = new THREE.ShaderMaterial({
        vertexShader: SHARED_VERT, fragmentShader: nebulaFrag,
        uniforms: { sunDir: { value: SUN }, time: timeUniform },
      });
    } else {
      const tex = new THREE.TextureLoader().load(photoUrl);
      tex.colorSpace = THREE.SRGBColorSpace;
      mat = new THREE.MeshStandardMaterial({ map: tex, roughness: 0.72, metalness: 0.04 });
    }

    const planet = new THREE.Mesh(geo, mat);
    planet.rotation.x = axialTilt;
    scene.add(planet);

    // Atmosphere rim — BackSide wraps around planet edge AND over the sprite
    const atmoOpacity = planetType === 'nebula' ? 0.20 : planetType === 'leo' ? 0.18 : 0.12;
    const atmoMat = new THREE.MeshBasicMaterial({
      color: new THREE.Color(glowColor), transparent: true,
      opacity: atmoOpacity, side: THREE.BackSide,
    });
    scene.add(new THREE.Mesh(new THREE.SphereGeometry(1.08, 32, 32), atmoMat));

    // Thin front-side atmosphere veil over product — integrates sprite with planet
    const veilMat = new THREE.MeshBasicMaterial({
      color: new THREE.Color(glowColor), transparent: true,
      opacity: 0.06, side: THREE.FrontSide, depthWrite: false,
    });
    scene.add(new THREE.Mesh(new THREE.SphereGeometry(1.09, 32, 32), veilMat));

    if (planetType === 'earth') {
      const sun = new THREE.DirectionalLight(0xfff4e0, 1.65);
      sun.position.set(4, 2, 5); scene.add(sun);
      const fill = new THREE.DirectionalLight(0x6699ff, 0.22);
      fill.position.set(-5, -1, -3); scene.add(fill);
      scene.add(new THREE.AmbientLight(0x0a0a14, 0.75));
    }

    // Product sprite(s) — rendered inside Three.js so veil wraps around them.
    // Precedence: carousel > products > productUrl.
    let cleanupCarousel: (() => void) | null = null;
    if (carousel && carousel.length > 0) {
      cleanupCarousel = loadCarousel(carousel, scene);
    } else if (products && products.length > 0) {
      products.forEach((item) => loadProductSprite(item, scene));
    } else if (productUrl) {
      loadProductSprite({ url: productUrl }, scene);
    }

    function animate() {
      rafId = requestAnimationFrame(animate);
      planet.rotation.y += rotationSpeed;
      if (timeUniform) timeUniform.value += 0.016;
      renderer.render(scene, camera);
    }
    animate();

    const obs = new ResizeObserver(() => {
      const w = canvas.clientWidth, h = canvas.clientHeight;
      camera.aspect = w/h; camera.updateProjectionMatrix();
      renderer.setSize(w, h);
    });
    obs.observe(canvas);

    return () => {
      obs.disconnect(); cancelAnimationFrame(rafId);
      if (cleanupCarousel) cleanupCarousel();
      renderer.dispose(); geo.dispose(); mat.dispose();
      atmoMat.dispose(); veilMat.dispose();
    };
  });
</script>

<canvas bind:this={canvas} class="planet-canvas" aria-hidden="true"></canvas>

<style>
  .planet-canvas {
    display: block;
    width: 100%;
    height: 100%;
    border-radius: 50%;
  }
</style>
