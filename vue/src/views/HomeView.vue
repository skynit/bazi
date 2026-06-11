<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

type WuxingKey = 'mu' | 'huo' | 'tu' | 'jin' | 'shui'

const router = useRouter()
const authStore = useAuthStore()

const savedChartId = ref<number | null>(null)
const currentYongshen = ref<WuxingKey>('mu')
const mounted = ref(false)
const sphereCanvas = ref<HTMLCanvasElement | null>(null)

const heavenlyStems = ['甲', '乙', '丙', '丁', '戊', '己', '庚', '辛', '壬', '癸']
const earthlyBranches = ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']
const wuxingCycle: Array<{ key: WuxingKey; label: string; glyph: string }> = [
  { key: 'mu', label: '木', glyph: '生发' },
  { key: 'huo', label: '火', glyph: '明照' },
  { key: 'tu', label: '土', glyph: '承载' },
  { key: 'jin', label: '金', glyph: '肃杀' },
  { key: 'shui', label: '水', glyph: '润下' }
]

type WuxingTheme = {
  seed: number
  accentRgb: string
  softRgb: string
  deepRgb: string
  accentHex: string
  accentDark: string
  buttonText: string
  glow1: string
  glow2: string
  ringColor: string
  ringShadow: string
}

const wuxingKeys: WuxingKey[] = ['mu', 'huo', 'tu', 'jin', 'shui']

function seededRandom(seed: number) {
  let value = seed % 2147483647
  if (value <= 0) value += 2147483646
  return () => {
    value = value * 16807 % 2147483647
    return (value - 1) / 2147483646
  }
}

function textureUrlFromCanvas(
  width: number,
  height: number,
  draw: (ctx: CanvasRenderingContext2D) => void
) {
  if (typeof document === 'undefined') return 'none'

  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height

  const ctx = canvas.getContext('2d', { alpha: true })
  if (!ctx) return 'none'

  ctx.clearRect(0, 0, width, height)
  draw(ctx)

  return `url("${canvas.toDataURL('image/png')}")`
}

function parseRgb(rgb: string): [number, number, number] {
  const [r, g, b] = rgb.split(',').map((value) => Number(value.trim()))
  return [r || 0, g || 0, b || 0]
}

function rgba(rgb: [number, number, number], alpha: number) {
  return `rgba(${rgb[0]},${rgb[1]},${rgb[2]},${alpha.toFixed(3)})`
}

function mixRgb(a: [number, number, number], b: [number, number, number], amount: number): [number, number, number] {
  const inverse = 1 - amount
  return [
    Math.round(a[0] * inverse + b[0] * amount),
    Math.round(a[1] * inverse + b[1] * amount),
    Math.round(a[2] * inverse + b[2] * amount)
  ]
}

function clamp01(value: number) {
  return Math.max(0, Math.min(1, value))
}

function compileShader(gl: WebGLRenderingContext, type: number, source: string) {
  const shader = gl.createShader(type)
  if (!shader) return null
  gl.shaderSource(shader, source)
  gl.compileShader(shader)
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    gl.deleteShader(shader)
    return null
  }
  return shader
}

function initSphereWebGL(canvas: HTMLCanvasElement, theme: WuxingTheme) {
  const context = canvas.getContext('webgl', { alpha: true, premultipliedAlpha: false })
  if (!context) return null
  const gl: WebGLRenderingContext = context

  const vsSource = `
    attribute vec2 a_position;
    void main() {
      gl_Position = vec4(a_position, 0.0, 1.0);
    }
  `

  const fsSource = `
    precision mediump float;
    uniform float u_time;
    uniform vec3 u_accent;
    uniform vec3 u_soft;
    uniform vec2 u_resolution;

    void main() {
      vec2 uv = gl_FragCoord.xy / u_resolution;

      vec2 center = vec2(0.5, 1.02);
      vec2 radius = vec2(0.72, 0.94);
      vec2 d = (uv - center) / radius;
      float dist = length(d);

      if (dist > 1.02) discard;

      float fresnel = pow(1.0 - abs(d.x) * 0.9, 4.0);

      float sweepPos = fract(u_time * 0.055);
      float sweepDist = abs(uv.x - sweepPos) * 4.0;
      float sweep = exp(-sweepDist * sweepDist) * smoothstep(0.0, 0.3, uv.y) * smoothstep(1.0, 0.6, uv.y);

      float hotspotX = (uv.x - 0.19) / 0.34;
      float hotspotY = (uv.y - 0.28) / 0.52;
      float hotspot = pow(1.0 - length(vec2(hotspotX, hotspotY)), 1.4);

      vec3 color = u_accent * 0.15 * fresnel
                 + u_accent * sweep * 0.4
                 + vec3(1.0) * hotspot * 0.08;

      float alpha = (fresnel * 0.5 + sweep * 0.6 + hotspot * 0.1) * smoothstep(1.02, 0.98, dist);

      gl_FragColor = vec4(color, alpha);
    }
  `

  const vs = compileShader(gl, gl.VERTEX_SHADER, vsSource)
  const fs = compileShader(gl, gl.FRAGMENT_SHADER, fsSource)
  if (!vs || !fs) return null

  const program = gl.createProgram()
  if (!program) return null
  gl.attachShader(program, vs)
  gl.attachShader(program, fs)
  gl.linkProgram(program)
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    gl.deleteProgram(program)
    return null
  }

  const posBuffer = gl.createBuffer()
  gl.bindBuffer(gl.ARRAY_BUFFER, posBuffer)
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 1, -1, -1, 1, 1, 1]), gl.STATIC_DRAW)

  const uTime = gl.getUniformLocation(program, 'u_time')
  const uAccent = gl.getUniformLocation(program, 'u_accent')
  const uSoft = gl.getUniformLocation(program, 'u_soft')
  const uResolution = gl.getUniformLocation(program, 'u_resolution')
  const aPosition = gl.getAttribLocation(program, 'a_position')

  const accent = parseRgb(theme.accentRgb)
  const soft = parseRgb(theme.softRgb)

  function resize() {
    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    const rect = canvas.getBoundingClientRect()
    const w = Math.round(rect.width * dpr)
    const h = Math.round(rect.height * dpr)
    if (canvas.width !== w || canvas.height !== h) {
      canvas.width = w
      canvas.height = h
    }
  }

  let animId = 0
  const startTime = performance.now()

  function render() {
    resize()
    gl.viewport(0, 0, canvas.width, canvas.height)
    gl.clearColor(0, 0, 0, 0)
    gl.clear(gl.COLOR_BUFFER_BIT)
    gl.enable(gl.BLEND)
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

    gl.useProgram(program)
    gl.uniform1f(uTime, (performance.now() - startTime) / 1000)
    gl.uniform3f(uAccent, accent[0] / 255, accent[1] / 255, accent[2] / 255)
    gl.uniform3f(uSoft, soft[0] / 255, soft[1] / 255, soft[2] / 255)
    gl.uniform2f(uResolution, canvas.width, canvas.height)

    gl.bindBuffer(gl.ARRAY_BUFFER, posBuffer)
    gl.enableVertexAttribArray(aPosition)
    gl.vertexAttribPointer(aPosition, 2, gl.FLOAT, false, 0, 0)
    gl.drawArrays(gl.TRIANGLE_STRIP, 0, 4)

    animId = requestAnimationFrame(render)
  }

  function updateTheme(newTheme: WuxingTheme) {
    const a = parseRgb(newTheme.accentRgb)
    const s = parseRgb(newTheme.softRgb)
    ;(accent as number[])[0] = a[0]; (accent as number[])[1] = a[1]; (accent as number[])[2] = a[2]
    ;(soft as number[])[0] = s[0]; (soft as number[])[1] = s[1]; (soft as number[])[2] = s[2]
  }

  function cleanup() {
    cancelAnimationFrame(animId)
    gl.deleteProgram(program)
    gl.deleteBuffer(posBuffer)
    if (vs) gl.deleteShader(vs)
    if (fs) gl.deleteShader(fs)
  }

  return { render, updateTheme, cleanup }
}

function createTheme(theme: WuxingTheme): WuxingTheme {
  return theme
}

function createDomeTexture(theme: WuxingTheme) {
  const random = seededRandom(theme.seed + 1907)
  const accent = parseRgb(theme.accentRgb)
  const soft = parseRgb(theme.softRgb)
  const deep = parseRgb(theme.deepRgb)
  const width = 1500
  const height = 920

  return textureUrlFromCanvas(width, height, (ctx) => {
    const cx = width * 0.5
    const cy = height * 1.02
    const rx = width * 0.72
    const ry = height * 0.94

    const fieldAt = (x: number, y: number) => {
      const dx = (x - cx) / rx
      const dy = (y - cy) / ry
      const radius = Math.sqrt(dx * dx + dy * dy)
      const inside = Math.pow(clamp01((1.018 - radius) / 0.09), 0.5)
      const body = Math.pow(clamp01(1 - radius / 1.02), 0.44)
      const shell = Math.pow(clamp01(1 - Math.abs(radius - 0.74) / 0.34), 1.05)
      const rim = Math.pow(clamp01(1 - Math.abs(radius - 0.992) / 0.052), 1.75)
      const leftLight = 0.28 + Math.pow(clamp01(1 - x / (width * 0.92)), 1.18) * 0.72
      const upperLight = Math.pow(clamp01(1 - y / (height * 0.78)), 1.05)
      const rightFade = 0.28 + Math.pow(clamp01(1 - (x - width * 0.56) / (width * 0.6)), 1.12) * 0.72
      const bottomFade = Math.pow(clamp01(1 - (y - height * 0.88) / (height * 0.12)), 0.6)
      const topFade = Math.pow(clamp01(y / (height * 0.18)), 0.32)
      const hotspotDx = (x - width * 0.19) / (width * 0.34)
      const hotspotDy = (y - height * 0.28) / (height * 0.52)
      const hotspot = Math.pow(clamp01(1 - Math.sqrt(hotspotDx * hotspotDx + hotspotDy * hotspotDy)), 1.4)
      const fresnel = Math.pow(1 - Math.abs(dx) * 0.9, 4)
      const grain = inside * rightFade * bottomFade * topFade
      return { body, shell, rim, leftLight, upperLight, hotspot, grain, fresnel }
    }

    for (let i = 0; i < 12000; i += 1) {
      const theta = random() * Math.PI
      const r = Math.sqrt(random())
      const x = cx + Math.cos(theta) * rx * r
      const y = cy - Math.sin(theta) * ry * r
      if (x < 0 || x > width || y < 0 || y > height) continue
      const field = fieldAt(x, y)
      if (!field.grain) continue

      const density = field.grain * clamp01(
        0.46 +
        field.body * 0.4 +
        field.shell * 0.58 +
        field.rim * 0.9 +
        field.leftLight * 0.34 +
        field.upperLight * 0.22 +
        field.hotspot * 0.58
      )
      if (random() > density) continue

      const brightness = clamp01(
        0.24 +
        field.body * 0.22 +
        field.shell * 0.34 +
        field.rim * 0.44 +
        field.leftLight * 0.62 +
        field.upperLight * 0.12 +
        field.hotspot * 0.72 +
        field.fresnel * 0.5
      )
      let alpha = Math.min(1, (0.46 + random() * 0.94) * brightness)
      const color = mixRgb(mixRgb(accent, soft, random() * 0.4), deep, random() * 0.08)
      let radius = random() > 0.64 ? 0.92 + random() * 1.72 : 0.52 + random() * 0.62

      if (random() > 0.995) {
        radius *= 2.5
        alpha = 1
      }

      ctx.fillStyle = rgba(color, alpha)
      if (radius < 0.86) {
        ctx.fillRect(x, y, 1, 1)
      } else {
        ctx.beginPath()
        ctx.arc(x, y, radius, 0, Math.PI * 2)
        ctx.fill()
      }
    }

    for (let i = 0; i < 8000; i += 1) {
      const theta = random() * Math.PI
      const r = Math.sqrt(random())
      const x = cx + Math.cos(theta) * rx * r
      const y = cy - Math.sin(theta) * ry * r
      if (x < 0 || x > width || y < 0 || y > height) continue
      const field = fieldAt(x, y)
      if (!field.grain) continue

      const density = field.grain * clamp01(0.38 + field.body * 0.32 + field.shell * 0.44 + field.rim * 0.66 + field.leftLight * 0.14)
      if (random() > density) continue

      const alpha = Math.min(0.72, (0.1 + random() * 0.42) * (0.3 + field.shell * 0.28 + field.rim * 0.16))
      const radius = random() > 0.86 ? 0.86 + random() * 1.78 : 0.36 + random() * 0.54

      ctx.fillStyle = `rgba(0,0,0,${alpha.toFixed(3)})`
      if (radius < 0.9) {
        ctx.fillRect(x, y, 1, 1)
      } else {
        ctx.beginPath()
        ctx.arc(x, y, radius, 0, Math.PI * 2)
        ctx.fill()
      }
    }

    for (let i = 0; i < 13500; i += 1) {
      const angle = Math.PI * (0.05 + random() * 0.9)
      const jitter = (random() - 0.5) * 0.026
      const x = cx + Math.cos(angle) * rx * (0.965 + jitter)
      const y = cy - Math.sin(angle) * ry * (0.965 + jitter)
      if (x < 0 || x > width || y < 0 || y > height) continue

      const side = clamp01(1 - x / (width * 0.72))
      if (random() > 0.28 + side * 0.72) continue

      const alpha = (0.16 + random() * 0.62) * (0.28 + side * 0.9)
      const radius = random() > 0.7 ? 0.82 + random() * 1.35 : 0.38 + random() * 0.5
      const color = mixRgb(accent, soft, random() * 0.5)

      ctx.fillStyle = rgba(color, Math.min(0.95, alpha))
      ctx.beginPath()
      ctx.arc(x, y, radius, 0, Math.PI * 2)
      ctx.fill()
    }

    ctx.globalCompositeOperation = 'source-atop'
    const shade = ctx.createLinearGradient(0, 0, width, 0)
    shade.addColorStop(0, 'rgba(0,0,0,0)')
    shade.addColorStop(0.48, 'rgba(0,0,0,0.02)')
    shade.addColorStop(0.76, 'rgba(0,0,0,0.24)')
    shade.addColorStop(1, 'rgba(0,0,0,0.62)')
    ctx.fillStyle = shade
    ctx.fillRect(0, 0, width, height)
    ctx.globalCompositeOperation = 'source-over'
  })
}

function createSweepTexture(theme: WuxingTheme) {
  const random = seededRandom(theme.seed + 4567)
  const accent = parseRgb(theme.accentRgb)
  const soft = parseRgb(theme.softRgb)
  const width = 680
  const height = 920

  return textureUrlFromCanvas(width, height, (ctx) => {
    for (let i = 0; i < 26000; i += 1) {
      const x = random() * width
      const y = random() * height
      const dx = Math.abs(x - width * 0.5) / (width * 0.44)
      const dy = Math.abs(y - height * 0.45) / (height * 0.72)
      const glow = Math.pow(clamp01(1 - dx), 1.65) * (0.38 + Math.pow(clamp01(1 - dy), 1.2) * 0.62)
      if (glow <= 0) continue

      const radius = random() > 0.8 ? 0.72 + random() * 1.2 : 0.34 + random() * 0.46
      const alpha = Math.min(0.82, (0.14 + random() * 0.62) * glow)
      const color = random() > 0.18 ? accent : mixRgb(accent, soft, 0.65)

      ctx.fillStyle = rgba(color, alpha)
      ctx.beginPath()
      ctx.arc(x, y, radius, 0, Math.PI * 2)
      ctx.fill()
    }

    const glow = ctx.createRadialGradient(width * 0.5, height * 0.45, 0, width * 0.5, height * 0.45, width * 0.5)
    glow.addColorStop(0, rgba(accent, 0.09))
    glow.addColorStop(0.46, rgba(soft, 0.035))
    glow.addColorStop(1, 'rgba(0,0,0,0)')
    ctx.fillStyle = glow
    ctx.fillRect(0, 0, width, height)
  })
}

const textureCache = new Map<WuxingKey, { dome: string; sweep: string }>()

function getThemeTextures(key: WuxingKey, theme: WuxingTheme) {
  const cached = textureCache.get(key)
  if (cached) return cached

  const textures = {
    dome: createDomeTexture(theme),
    sweep: createSweepTexture(theme)
  }
  textureCache.set(key, textures)

  return textures
}

function scheduleTextureWarmup() {
  if (typeof window === 'undefined') return

  const queue = wuxingKeys.filter((key) => key !== currentYongshen.value)
  const schedule = (callback: () => void) => {
    const requestIdle = window.requestIdleCallback
    if (requestIdle) {
      requestIdle(callback, { timeout: 1200 })
      return
    }
    setTimeout(callback, 220)
  }

  const warmNext = () => {
    const key = queue.shift()
    if (!key) return

    getThemeTextures(key, wuxingThemes[key])
    schedule(warmNext)
  }

  setTimeout(() => schedule(warmNext), 700)
}

const wuxingThemes: Record<WuxingKey, WuxingTheme> = {
  mu: createTheme({
    seed: 1101,
    accentRgb: '34, 211, 153',
    softRgb: '20, 184, 166',
    deepRgb: '4, 72, 58',
    accentHex: '#34d399',
    accentDark: '#059669',
    buttonText: '#00140e',
    glow1: 'radial-gradient(ellipse, rgba(34,211,153,.52), rgba(20,184,166,.18) 44%, transparent 72%)',
    glow2: 'radial-gradient(ellipse, rgba(45,212,191,.32), transparent 70%)',
    ringColor: '#5fffd8',
    ringShadow: '#00ffc3'
  }),
  huo: createTheme({
    seed: 2203,
    accentRgb: '251, 113, 133',
    softRgb: '244, 63, 94',
    deepRgb: '76, 5, 25',
    accentHex: '#fb7185',
    accentDark: '#e11d48',
    buttonText: '#190005',
    glow1: 'radial-gradient(ellipse, rgba(251,113,133,.5), rgba(244,63,94,.18) 44%, transparent 72%)',
    glow2: 'radial-gradient(ellipse, rgba(168,85,247,.24), transparent 70%)',
    ringColor: '#ff9aaa',
    ringShadow: '#ff4d6d'
  }),
  tu: createTheme({
    seed: 3307,
    accentRgb: '252, 211, 77',
    softRgb: '245, 158, 11',
    deepRgb: '69, 26, 3',
    accentHex: '#fcd34d',
    accentDark: '#d97706',
    buttonText: '#1a1000',
    glow1: 'radial-gradient(ellipse, rgba(252,211,77,.5), rgba(245,158,11,.16) 44%, transparent 72%)',
    glow2: 'radial-gradient(ellipse, rgba(249,115,22,.24), transparent 70%)',
    ringColor: '#ffe29a',
    ringShadow: '#f59e0b'
  }),
  jin: createTheme({
    seed: 4409,
    accentRgb: '226, 232, 240',
    softRgb: '148, 163, 184',
    deepRgb: '39, 39, 42',
    accentHex: '#e2e8f0',
    accentDark: '#94a3b8',
    buttonText: '#030404',
    glow1: 'radial-gradient(ellipse, rgba(226,232,240,.42), rgba(148,163,184,.14) 44%, transparent 72%)',
    glow2: 'radial-gradient(ellipse, rgba(203,213,225,.2), transparent 70%)',
    ringColor: '#d8fff4',
    ringShadow: '#b4fff0'
  }),
  shui: createTheme({
    seed: 5513,
    accentRgb: '34, 211, 238',
    softRgb: '59, 130, 246',
    deepRgb: '8, 47, 73',
    accentHex: '#22d3ee',
    accentDark: '#2563eb',
    buttonText: '#001116',
    glow1: 'radial-gradient(ellipse, rgba(34,211,238,.48), rgba(59,130,246,.18) 44%, transparent 72%)',
    glow2: 'radial-gradient(ellipse, rgba(14,165,233,.26), transparent 70%)',
    ringColor: '#7dfff5',
    ringShadow: '#37e8dc'
  })
}

// CSS 自定义变量，驱动穹顶各层
const domeThemeStyles = computed(() => {
  const key = currentYongshen.value
  const theme = wuxingThemes[key]
  const textures = getThemeTextures(key, theme)
  return {
    '--jade-accent-rgb': theme.accentRgb,
    '--jade-soft-rgb': theme.softRgb,
    '--jade-deep-rgb': theme.deepRgb,
    '--jade-accent': theme.accentHex,
    '--jade-accent-dark': theme.accentDark,
    '--jade-button-text': theme.buttonText,
    '--jade-dome-texture': textures.dome,
    '--jade-sweep-texture': textures.sweep,
    '--dome-glow-1': theme.glow1,
    '--dome-glow-2': theme.glow2,
    '--ring-color': theme.ringColor,
    '--ring-shadow': theme.ringShadow
  }
})

let webglHandle: { render: () => void; updateTheme: (t: WuxingTheme) => void; cleanup: () => void } | null = null

onMounted(async () => {
  if (authStore.isLoggedIn() && !authStore.user) {
    await authStore.fetchMe().catch(() => {})
  }
  const saved = localStorage.getItem('bazi_last_birth')
  if (saved) {
    try { savedChartId.value = JSON.parse(saved).chartId || null } catch {}
  }
  setTimeout(() => mounted.value = true, 100)
  scheduleTextureWarmup()

  if (sphereCanvas.value) {
    webglHandle = initSphereWebGL(sphereCanvas.value, wuxingThemes[currentYongshen.value])
    if (webglHandle) webglHandle.render()
  }
})

onUnmounted(() => {
  if (webglHandle) {
    webglHandle.cleanup()
    webglHandle = null
  }
})

watch(currentYongshen, (key) => {
  if (webglHandle) webglHandle.updateTheme(wuxingThemes[key])
})

function startChart() { router.push('/chart/new') }
function continueChart() { if (savedChartId.value) router.push(`/chart/${savedChartId.value}`) }
function switchYongshen(key: WuxingKey) { currentYongshen.value = key }
</script>

<template>
  <div class="home-page" :class="{ visible: mounted }" :style="domeThemeStyles">
    <!-- ==================== Jadense 穹顶结构 ==================== -->
    <div class="hero-dome">
      <!-- Glow -->
      <div class="dome-glow dome-glow-1"></div>
      <div class="dome-glow dome-glow-2"></div>

      <!-- Main Dome -->
      <canvas ref="sphereCanvas" class="dome-webgl"></canvas>
      <div class="dome-core"></div>
      <div class="dome-sweep"></div>
      <div class="dome-sweep dome-sweep-2"></div>
      <div class="dome-highlight"></div>

      <!-- Main Arc -->
      <svg class="dome-rings" viewBox="0 0 2000 1000" preserveAspectRatio="none">
        <path d="M-300 920 A1300 1300 0 0 1 2300 920" class="ring-main" />
        <path d="M-150 940 A1150 1150 0 0 1 2150 940" class="ring-sub" />
      </svg>

      <!-- Scale Arc -->
      <svg class="arc-scale-layer" viewBox="0 0 2000 1000" preserveAspectRatio="none">
        <path d="M-200 900 A1200 1200 0 0 1 2200 900" class="arc-scale" />
      </svg>

      <!-- HUD Scale -->
      <svg class="hud-scale" viewBox="0 0 2000 1000" preserveAspectRatio="none">
        <g v-for="i in 80" :key="i">
          <line
            :x1="150 + i * 22"
            y1="845"
            :x2="150 + i * 22"
            :y2="i % 5 === 0 ? 810 : 828"
            class="scale-line"
          />
        </g>
      </svg>

      <!-- Energy Core -->
      <div class="energy-core"></div>

      <!-- Horizon -->
      <div class="horizon-line"></div>

      <!-- Energy Beam -->
      <div class="energy-beam"></div>

      <!-- Astrolabe -->
      <svg viewBox="0 0 800 800" class="astrolabe-overlay">
        <g class="animate-[spin_240s_linear_infinite]" style="transform-origin:400px 400px">
          <circle cx="400" cy="400" r="385" fill="none" stroke="rgba(255,255,255,.18)" stroke-width="1" />
          <circle cx="400" cy="400" r="300" fill="none" stroke="rgba(255,255,255,.12)" stroke-width="1" />
          <circle cx="400" cy="400" r="214" fill="none" stroke="rgba(255,255,255,.1)" stroke-width="1" />
          <g v-for="(branch, index) in earthlyBranches" :key="branch" :transform="`rotate(${index * 30} 400 400)`">
            <line x1="400" y1="775" x2="400" y2="785" stroke="rgba(255,255,255,.18)" />
            <text x="400" y="758" text-anchor="middle" fill="rgba(255,255,255,.22)" font-size="12">
              {{ branch }}
            </text>
          </g>
          <g v-for="(stem, index) in heavenlyStems" :key="stem" :transform="`rotate(${index * 36 + 18} 400 400)`">
            <line x1="400" y1="690" x2="400" y2="705" stroke="rgba(255,255,255,.12)" />
            <text x="400" y="675" text-anchor="middle" fill="rgba(255,255,255,.18)" font-size="11">
              {{ stem }}
            </text>
          </g>
        </g>
      </svg>

      <!-- HUD Decor -->
      <div class="hud-left"></div>
      <div class="hud-right"></div>

      <!-- Noise -->
      <div class="dome-noise"></div>
    </div>

    <!-- 前景内容区域 -->
    <main class="hero-main">
      <div class="title-block">
        <div class="eyebrow">
          <span class="eyebrow-line"></span>
          Bazi · Ziwei Astral System
          <span class="eyebrow-line"></span>
        </div>
        <h1 class="hero-title">
          以天干地支<br />
          <span class="title-accent">推演命局流转</span>
        </h1>
        <p class="hero-sub">See Further · Think Deeper · Act Faster</p>
      </div>

      <div class="cta-group">
        <button @click="startChart" class="btn-primary-base">
          开始排盘
        </button>
        <button v-if="savedChartId" @click="continueChart" class="btn-secondary">
          继续上次
        </button>
      </div>

      <div class="element-rail" aria-label="五行切换">
        <button
          v-for="item in wuxingCycle"
          :key="item.key"
          type="button"
          :class="{ active: currentYongshen === item.key }"
          @click="switchYongshen(item.key)"
        >
          {{ item.label }}
        </button>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* =============================================
   页面容器
   ============================================= */
.home-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  overflow: hidden;
  padding: clamp(64px, 9vh, 92px) 24px 72px;
  background:
    radial-gradient(ellipse at 20% 0%, rgba(var(--jade-accent-rgb), .03), transparent 38%),
    radial-gradient(ellipse at 70% 0%, rgba(255,255,255,.018), transparent 28%),
    linear-gradient(180deg, #020303 0%, #000101 58%, #000 100%);
  color: #f4f4f5;
  opacity: 0;
  transform: translateY(-2px);
  transition: opacity 1.2s cubic-bezier(0.16, 1, 0.3, 1), transform 1.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.home-page.visible {
  opacity: 1;
  transform: translateY(0);
}

.home-page::after {
  content: "";
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 22% 4%, transparent 0%, rgba(0,0,0,.02) 40%, rgba(0,0,0,.18) 76%, rgba(0,0,0,.5) 100%),
    linear-gradient(90deg, rgba(0,0,0,0) 0%, rgba(0,0,0,.04) 58%, rgba(0,0,0,.5) 100%),
    linear-gradient(180deg, rgba(0,0,0,0) 0%, rgba(0,0,0,.03) 44%, rgba(0,0,0,.62) 100%);
  pointer-events: none;
  z-index: 5;
}

.home-page::before {
  content: "";
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
  background:
    radial-gradient(ellipse at 14% 0%, rgba(var(--jade-accent-rgb), .028), transparent 46%),
    linear-gradient(90deg, rgba(255,255,255,.01), transparent 34%, rgba(0,0,0,.3) 100%);
  opacity: .38;
}

/* =============================================
   半圆点阵舞台
   ============================================= */
.hero-dome {
  position: absolute;
  width: min(1480px, 118vw);
  height: min(740px, 59vw);
  left: 50%;
  top: clamp(42px, 7vh, 86px);
  transform: translateX(-50%);
  overflow: hidden;
  border-radius: 50% 50% 0 0 / 100% 100% 0 0;
  pointer-events: none;
  z-index: 1;
  isolation: isolate;
  contain: paint;
  background:
    radial-gradient(ellipse at 30% 18%, rgba(255,255,255,.12), transparent 34%),
    radial-gradient(ellipse at 50% 102%, rgba(var(--jade-accent-rgb), .28), rgba(var(--jade-accent-rgb), .08) 36%, transparent 70%),
    linear-gradient(115deg, rgba(var(--jade-deep-rgb), .52), rgba(4,7,7,.9) 58%, rgba(0,0,0,.98));
  box-shadow:
    inset 0 1px 0 rgba(255,255,255,.16),
    inset 0 -1px 0 rgba(var(--jade-accent-rgb), .1),
    0 32px 120px rgba(0,0,0,.46);
}

.hero-dome::before,
.hero-dome::after {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
}

.hero-dome::before {
  z-index: 0;
  border: 1px solid rgba(var(--jade-accent-rgb), .18);
  border-bottom: 0;
  background:
    linear-gradient(90deg, rgba(255,255,255,.1), transparent 18%, transparent 78%, rgba(0,0,0,.3)),
    radial-gradient(ellipse at 50% 98%, transparent 0 56%, rgba(255,255,255,.05) 57%, transparent 58%);
  mix-blend-mode: screen;
  opacity: .7;
}

.hero-dome::after {
  z-index: 8;
  inset: 18px 30px 0;
  border: 1px solid rgba(255,255,255,.07);
  border-bottom: 0;
  opacity: .8;
}

/* =============================================
   双层 Glow
   ============================================= */
.dome-glow {
  display: none;
}

.dome-glow-1 {
  display: none;
}

.dome-glow-2 {
  display: none;
}

/* =============================================
   核心椭圆渐变
   ============================================= */
.dome-core {
  position: absolute;
  inset: 0;
  z-index: 1;
  background: var(--jade-dome-texture);
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  transition: background-image .7s cubic-bezier(0.16, 1, 0.3, 1);
  mix-blend-mode: screen;
  opacity: .92;
}

.dome-core::before {
  display: none;
}

.dome-core::after {
  display: none;
}

.dome-sweep {
  position: absolute;
  top: 0;
  bottom: 0;
  left: -18%;
  width: 36%;
  z-index: 2;
  pointer-events: none;
  contain: paint;
  opacity: 0;
  background-image: var(--jade-sweep-texture);
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  transform: translate3d(-12%, 0, 0);
  will-change: transform, opacity;
  animation: domeSweep 18s linear infinite;
}

.dome-sweep::before,
.dome-sweep::after {
  display: none;
}

.dome-sweep-2 {
  filter: blur(24px);
  animation-duration: 22s;
}

.dome-highlight {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(
      ellipse at 50% 75%,
      rgba(255,255,255,.18),
      transparent 45%
    );
  mix-blend-mode: screen;
  pointer-events: none;
  z-index: 3;
}

.dome-webgl {
  position: absolute;
  inset: 0;
  z-index: 4;
  pointer-events: none;
  mix-blend-mode: screen;
}

/* =============================================
   极细发光弧线
   ============================================= */
.dome-rings {
  position: absolute;
  inset: 0;
  z-index: 5;
  opacity: .7;
}

.ring-main {
  fill: none;
  stroke: var(--ring-color);
  stroke-width: .8;
  filter: drop-shadow(0 0 6px var(--ring-shadow));
  transition: stroke 1s ease;
  opacity: .14;
}

.ring-sub {
  fill: none;
  stroke: var(--ring-color);
  stroke-width: 0.8;
  opacity: 0.055;
}

.arc-scale-layer {
  position: absolute;
  inset: 0;
  z-index: 5;
}

.arc-scale {
  fill: none;
  stroke: var(--ring-color);
  stroke-width: 1;
  stroke-dasharray: 4 14;
  opacity: 0.055;
}

.hud-scale {
  position: absolute;
  inset: 0;
  z-index: 6;
}

.scale-line {
  stroke: var(--ring-color);
  stroke-width: 1;
  opacity: 0.052;
}

.energy-core {
  display: none;
}

.horizon-line {
  position: absolute;
  left: 50%;
  bottom: 0;
  width: min(1180px, 88vw);
  height: 1px;
  transform: translateX(-50%);
  z-index: 9;
  background: linear-gradient(90deg, transparent, var(--ring-color) 18%, rgba(255,255,255,.5) 50%, var(--ring-color) 82%, transparent);
  opacity: .16;
  box-shadow:
    0 0 12px var(--ring-shadow),
    0 0 30px var(--ring-shadow);
}

.energy-beam {
  display: none;
}

.astrolabe-overlay {
  position: fixed;
  width: min(760px, 74vw);
  height: min(760px, 74vw);
  left: 50%;
  top: 67%;
  transform: translate(-50%, -50%);
  opacity: .038;
}

.hud-left,
.hud-right {
  position: absolute;
  width: 180px;
  height: 180px;
  border-top: 1px solid var(--ring-color);
  opacity: 0.04;
  transition: border-color 1s ease;
}

.hud-left {
  left: 12%;
  top: 58%;
  transform: rotate(-45deg);
}

.hud-right {
  right: 12%;
  top: 58%;
  transform: rotate(45deg);
}

.dome-noise {
  display: none;
}

.hero-main {
  position: relative;
  z-index: 20;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  width: min(920px, 100%);
  padding: 0 20px;
  gap: 24px;
  margin-top: clamp(10px, 2.4vh, 30px);
  transform: translateZ(0);
}

.title-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 22px;
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 10px;
  font-family: var(--font-mono), monospace;
  letter-spacing: 0.45em;
  text-transform: uppercase;
  color: rgba(235,255,248,.38);
}

.eyebrow-line {
  display: block;
  width: 28px;
  height: 1px;
  background: rgba(var(--jade-accent-rgb), .18);
}

.hero-title {
  font-family: var(--font-serif), 'Songti SC', serif;
  font-size: clamp(3.4rem, 6.25vw, 6.2rem);
  font-weight: 400;
  letter-spacing: 0.2em;
  color: rgba(255,255,255,.99);
  margin: 0;
  line-height: 1.26;
  padding-left: 0.2em;
  text-shadow:
    0 2px 10px rgba(0,0,0,.66),
    0 20px 60px rgba(0,0,0,.62);
}

.title-accent {
  display: block;
  margin-top: 10px;
  color: rgba(var(--jade-accent-rgb), .88);
  text-shadow:
    0 0 20px rgba(var(--jade-accent-rgb), .16),
    0 16px 54px rgba(0,0,0,.78);
}

.hero-sub {
  max-width: 760px;
  font-family: var(--font-serif), serif;
  font-size: clamp(1.1rem, 2.35vw, 2rem);
  color: rgba(245,255,251,.92);
  letter-spacing: 0.035em;
  margin: 0;
  text-shadow: 0 3px 26px rgba(0,0,0,.76);
}

.cta-group {
  display: flex;
  gap: 22px;
  flex-wrap: wrap;
  justify-content: center;
  margin-top: 6px;
}

.btn-primary-base {
  width: 168px;
  padding: 13px 0;
  border: none;
  border-radius: 50px;
  background: linear-gradient(135deg, var(--jade-accent) 0%, var(--jade-accent-dark) 100%);
  color: var(--jade-button-text);
  font-size: 0.92rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  cursor: pointer;
  box-shadow:
    0 12px 38px rgba(var(--jade-accent-rgb), .42),
    inset 0 1px 0 rgba(255,255,255,.22);
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.btn-primary-base:hover {
  transform: translateY(-2px);
}

.btn-secondary {
  width: 168px;
  padding: 13px 0;
  background: rgba(255,255,255,.02);
  color: rgba(235,255,249,.72);
  font-size: 0.92rem;
  font-weight: 500;
  letter-spacing: 0.08em;
  border: 1px solid rgba(217,255,241,.16);
  border-radius: 50px;
  cursor: pointer;
  backdrop-filter: blur(12px);
  transition: all 0.3s ease;
}

.btn-secondary:hover {
  color: #ffffff;
  border-color: rgba(var(--jade-accent-rgb), .36);
  background: rgba(255,255,255,.05);
}

.element-rail {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 2px;
  opacity: .55;
}

.element-rail button {
  width: 28px;
  height: 28px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: 50%;
  background: transparent;
  color: rgba(229,255,246,.46);
  font-family: var(--font-serif), serif;
  font-size: .85rem;
  cursor: pointer;
  transition: color .3s ease, border-color .3s ease, background .3s ease;
}

.element-rail button:hover,
.element-rail button.active {
  border-color: rgba(var(--jade-accent-rgb), .3);
  background: rgba(0,0,0,.18);
  color: rgba(var(--jade-accent-rgb), .9);
}

@keyframes domeSweep {
  0% {
    opacity: 0;
    transform: translate3d(0, 0, 0);
  }
  8% {
    opacity: .78;
  }
  62% {
    opacity: .9;
  }
  92% {
    opacity: 0;
    transform: translate3d(315%, 0, 0);
  }
  100% {
    opacity: 0;
    transform: translate3d(315%, 0, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .dome-core,
  .dome-glow-1,
  .dome-glow-2,
  .dome-sweep,
  .dome-sweep-2 {
    animation: none;
  }

  .dome-sweep {
    opacity: .42;
    transform: translate3d(82%, 0, 0);
  }

  .dome-sweep-2 {
    opacity: .2;
    transform: translate3d(82%, 0, 0);
  }
}

/* =============================================
   响应式
   ============================================= */
@media (max-width: 1024px) {
  .hero-dome {
    width: 138vw;
    height: 68vw;
    left: 50%;
    top: 34px;
  }
  .hero-main {
    margin-top: 0;
  }
}

@media (max-width: 640px) {
  .home-page {
    min-height: 100svh;
    padding: 58px 18px 42px;
  }
  .hero-dome {
    width: 194vw;
    height: 112vw;
    left: 50%;
    top: 22px;
    transform: translateX(-50%);
  }
  .hero-main {
    padding: 0;
    margin-top: 0;
    gap: 22px;
  }
  .title-block {
    gap: 18px;
  }
  .hero-title {
    font-size: clamp(2.55rem, 12vw, 3.55rem);
    letter-spacing: 0.1em;
    padding-left: 0.1em;
    line-height: 1.28;
  }
  .hero-sub {
    max-width: 340px;
    font-size: 1rem;
    line-height: 1.45;
  }
  .eyebrow {
    gap: 10px;
    font-size: 8px;
    letter-spacing: .18em;
  }
  .cta-group {
    flex-direction: column;
    gap: 12px;
  }
  .btn-primary-base,
  .btn-secondary {
    width: min(260px, calc(100vw - 48px));
  }
  .element-rail {
    gap: 8px;
  }
}
</style>
