<!-- web/src/lib/components/MissionPlanet.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import * as THREE from 'three';

  export let planetType: 'earth' | 'leo' | 'lunar' | 'nebula' = 'earth';
  export let photoUrl: string   = '';
  export let productUrl: string   = '';   // transparent PNG — rendered as in-scene sprite
  export let productScale: number   = 1.0;   // per-product size multiplier
  export let productOffsetY: number  = 0.0;          // vertical shift (+ = up)
  export let productOpacity: number  = 0.72;         // sprite opacity
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

  // ── Product sprite — apply radial alpha fade via canvas ───────────────────
  function loadProductSprite(url: string, scene: THREE.Scene) {
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
      // Base 1.5 world units, scaled per product
      const spriteH = 1.5 * productScale;
      const spriteW = spriteH * aspect;

      const spriteMat = new THREE.SpriteMaterial({
        map:         fadedTex,
        transparent: true,
        opacity:     productOpacity,
        blending:    productBlending === 'additive' ? THREE.AdditiveBlending : THREE.NormalBlending,
        depthWrite:  false,
        depthTest:   false,
      });
      const sprite = new THREE.Sprite(spriteMat);
      sprite.scale.set(spriteW, spriteH, 1);
      sprite.position.set(0, productOffsetY, 1.12);
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

    // Product sprite — rendered inside Three.js so veil wraps around it
    if (productUrl) loadProductSprite(productUrl, scene);

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
