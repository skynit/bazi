import {
  getShaderColorFromString,
  getShaderNoiseTexture,
  grainGradientFragmentShader,
  ditheringFragmentShader,
  dotGridFragmentShader,
  halftoneCmykFragmentShader,
  spiralFragmentShader,
  GrainGradientShapes,
  DitheringShapes,
  DitheringTypes,
  DotGridShapes,
  HalftoneCmykTypes,
  ShaderFitOptions,
  type ShaderMountUniforms
} from '@paper-design/shaders'

import muBg from '../assets/background/木.png'
import huoBg from '../assets/background/火.png'
import tuBg from '../assets/background/土.png'
import jinBg from '../assets/background/金.png'
import shuiBg from '../assets/background/水.png'

export type WuxingKey = 'mu' | 'huo' | 'tu' | 'jin' | 'shui'
export type ShaderType = 'grainGradient' | 'dithering' | 'dotGrid' | 'halftoneCmyk' | 'spiral'
export type ShaderThemeMode = 'light' | 'dark'

export type WuxingTheme = {
  accentRgb: string
  accentHex: string
  accentDark: string
  buttonText: string
  shaderBase: string
  shaderGlow: string
  shaderBaseLight: string
  shaderGlowLight: string
  /** HalftoneCMYK paper background */
  hcColorBack: string
  /** HalftoneCMYK cyan ink */
  hcColorC: string
  /** HalftoneCMYK magenta ink — the dominant element color */
  hcColorM: string
  /** HalftoneCMYK yellow ink */
  hcColorY: string
  /** HalftoneCMYK black ink */
  hcColorK: string
  /** shadcn --primary in light mode (oklch) */
  primaryOklchLight: string
  /** shadcn --primary in dark mode (oklch) */
  primaryOklchDark: string
}

export const bgImages: Record<WuxingKey, string> = {
  mu: muBg,
  huo: huoBg,
  tu: tuBg,
  jin: jinBg,
  shui: shuiBg,
}

export const wuxingThemes: Record<WuxingKey, WuxingTheme> = {
  mu: {
    accentRgb: '34, 211, 153',
    accentHex: '#34d399',
    accentDark: '#059669',
    buttonText: '#00140e',
    shaderBase: '#147b33',
    shaderGlow: '#43dfabcf',
    shaderBaseLight: '#bdf7d5',
    shaderGlowLight: '#71eec0cf',
    hcColorBack: '#fbfaf4',
    hcColorC: '#00b3ff',
    hcColorM: '#34d399',
    hcColorY: '#ffd900',
    hcColorK: '#231f20',
    primaryOklchLight: 'oklch(0.58 0.20 160)',
    primaryOklchDark: 'oklch(0.62 0.17 160)',
  },
  huo: {
    accentRgb: '251, 113, 133',
    accentHex: '#fb7185',
    accentDark: '#e11d48',
    buttonText: '#190005',
    shaderBase: '#7f1d1d',
    shaderGlow: '#fb7185cf',
    shaderBaseLight: '#ffe1de',
    shaderGlowLight: '#ff8f9fcf',
    hcColorBack: '#fbfaf4',
    hcColorC: '#00b3ff',
    hcColorM: '#fc4f9d',
    hcColorY: '#ffd900',
    hcColorK: '#231f20',
    primaryOklchLight: 'oklch(0.62 0.22 12)',
    primaryOklchDark: 'oklch(0.66 0.20 12)',
  },
  tu: {
    accentRgb: '252, 211, 77',
    accentHex: '#fcd34d',
    accentDark: '#d97706',
    buttonText: '#1a1000',
    shaderBase: '#7c4f08',
    shaderGlow: '#fcd34dcf',
    shaderBaseLight: '#fff1b8',
    shaderGlowLight: '#f6c85acf',
    hcColorBack: '#fbfaf4',
    hcColorC: '#00b3ff',
    hcColorM: '#d97706',
    hcColorY: '#ffd900',
    hcColorK: '#231f20',
    primaryOklchLight: 'oklch(0.74 0.16 78)',
    primaryOklchDark: 'oklch(0.78 0.14 78)',
  },
  jin: {
    accentRgb: '226, 232, 240',
    accentHex: '#e2e8f0',
    accentDark: '#94a3b8',
    buttonText: '#030404',
    shaderBase: '#334155',
    shaderGlow: '#e2e8f0cf',
    shaderBaseLight: '#e8eef5',
    shaderGlowLight: '#aebfd3cf',
    hcColorBack: '#f4f0f8',
    hcColorC: '#7b9ecf',
    hcColorM: '#94a3b8',
    hcColorY: '#e8e0d0',
    hcColorK: '#1a1a2e',
    primaryOklchLight: 'oklch(0.62 0.04 240)',
    primaryOklchDark: 'oklch(0.78 0.03 240)',
  },
  shui: {
    accentRgb: '34, 211, 238',
    accentHex: '#22d3ee',
    accentDark: '#2563eb',
    buttonText: '#001116',
    shaderBase: '#075985',
    shaderGlow: '#22d3eecf',
    shaderBaseLight: '#c6f4fe',
    shaderGlowLight: '#7fe2f4cf',
    hcColorBack: '#f0f8ff',
    hcColorC: '#2563eb',
    hcColorM: '#22d3ee',
    hcColorY: '#a0c4ff',
    hcColorK: '#0d1b2a',
    primaryOklchLight: 'oklch(0.66 0.14 220)',
    primaryOklchDark: 'oklch(0.74 0.14 220)',
  },
}

const shaderNoiseTexture = getShaderNoiseTexture()

function grainGradientUniforms(theme: WuxingTheme, mode: ShaderThemeMode = 'dark'): ShaderMountUniforms {
  const isDark = mode === 'dark'
  return {
    u_colorBack: getShaderColorFromString(isDark ? '#000000' : '#e5f2ed'),
    u_colors: [
      getShaderColorFromString(theme.shaderBase),
      getShaderColorFromString(theme.shaderGlow),
      getShaderColorFromString(isDark ? '#000000' : '#F5FBF6'),
      getShaderColorFromString(isDark ? '#000000' : '#edf7ef')
    ],
    u_colorsCount: 4,
    u_softness: 1,
    u_intensity: 1,
    u_noise: 0,
    u_shape: GrainGradientShapes.sphere,
    u_noiseTexture: shaderNoiseTexture,
    u_fit: ShaderFitOptions.cover,
    u_scale: 0.8,
    u_rotation: 360,
    u_originX: 0.5,
    u_originY:  0.5,
    u_offsetX: 0,
    u_offsetY: 0.3,
    u_worldWidth: 0,
    u_worldHeight: 0
  }
}

function ditheringUniforms(theme: WuxingTheme): ShaderMountUniforms {
  return {
    u_colorBack: getShaderColorFromString('#030404'),
    u_colorFront: getShaderColorFromString(theme.accentHex),
    u_shape: DitheringShapes.simplex,
    u_type: DitheringTypes['4x4'],
    u_pxSize: 2.5,
    u_fit: ShaderFitOptions.cover,
    u_scale: 1,
    u_rotation: 0,
    u_originX: 0.5,
    u_originY: 0.5,
    u_offsetX: 0,
    u_offsetY: 0,
    u_worldWidth: 0,
    u_worldHeight: 0
  }
}

function dotGridUniforms(theme: WuxingTheme): ShaderMountUniforms {
  return {
    u_colorBack: getShaderColorFromString('#030404'),
    u_colorFill: getShaderColorFromString(theme.accentHex),
    u_colorStroke: getShaderColorFromString(theme.accentDark),
    u_dotSize: 3,
    u_gapX: 24,
    u_gapY: 24,
    u_strokeWidth: 0,
    u_sizeRange: 0.3,
    u_opacityRange: 0.5,
    u_shape: DotGridShapes.circle,
    u_fit: ShaderFitOptions.cover,
    u_scale: 1,
    u_rotation: 0,
    u_originX: 0.5,
    u_originY: 0.5,
    u_offsetX: 0,
    u_offsetY: 0,
    u_worldWidth: 0,
    u_worldHeight: 0
  }
}

function halftoneCmykUniforms(theme: WuxingTheme, image: HTMLImageElement): ShaderMountUniforms {
  return {
    u_image: image,
    u_noiseTexture: shaderNoiseTexture,
    u_colorBack: getShaderColorFromString(theme.hcColorBack),
    u_colorC: getShaderColorFromString(theme.hcColorC),
    u_colorM: getShaderColorFromString(theme.hcColorM),
    u_colorY: getShaderColorFromString(theme.hcColorY),
    u_colorK: getShaderColorFromString(theme.hcColorK),
    u_size: 0.2,
    u_contrast: 1,
    u_softness: 1,
    u_grainSize: 0.5,
    u_grainMixer: 0,
    u_grainOverlay: 0,
    u_gridNoise: 0.2,
    u_floodC: 0.15,
    u_floodM: 0,
    u_floodY: 0,
    u_floodK: 0,
    u_gainC: 0.3,
    u_gainM: 0,
    u_gainY: 0.2,
    u_gainK: 0,
    u_type: HalftoneCmykTypes.ink,
    u_fit: ShaderFitOptions.cover,
    u_scale: 1,
    u_rotation: 0,
    u_originX: 0.5,
    u_originY: 0.5,
    u_offsetX: 0,
    u_offsetY: 0,
    u_worldWidth: 0,
    u_worldHeight: 0
  }
}

function spiralUniforms(_theme: WuxingTheme): ShaderMountUniforms {
  return {
    u_colorBack: getShaderColorFromString('#ffffff'),
    u_colorFront: getShaderColorFromString('#000000'),
    u_density: 0.16,
    u_distortion: 0,
    u_strokeWidth: 0.53,
    u_strokeTaper: 0,
    u_strokeCap: 0,
    u_noise: 0,
    u_noiseFrequency: 0,
    u_softness: 0.33,
    u_fit: ShaderFitOptions.cover,
    u_scale: 4,
    u_rotation: 0,
    u_originX: 0.5,
    u_originY: 0.5,
    u_offsetX: 0,
    u_offsetY: 0,
    u_worldWidth: 0,
    u_worldHeight: 0
  }
}

export function getShaderFragment(type: ShaderType): string {
  switch (type) {
    case 'dithering': return ditheringFragmentShader
    case 'dotGrid': return dotGridFragmentShader
    case 'halftoneCmyk': return halftoneCmykFragmentShader
    case 'spiral': return spiralFragmentShader
    default: return grainGradientFragmentShader
  }
}

export function createShaderUniforms(
  type: ShaderType,
  yongshen: WuxingKey,
  image?: HTMLImageElement | null,
  mode: ShaderThemeMode = 'dark'
): ShaderMountUniforms {
  const theme = wuxingThemes[yongshen]
  switch (type) {
    case 'dithering': return ditheringUniforms(theme)
    case 'dotGrid': return dotGridUniforms(theme)
    case 'halftoneCmyk': return halftoneCmykUniforms(theme, image!)
    case 'spiral': return spiralUniforms(theme)
    default: return grainGradientUniforms(theme, mode)
  }
}

export function getShaderSpeed(type: ShaderType): number {
  switch (type) {
    case 'spiral': return 0.4
    case 'grainGradient': return 1.5
    default: return 0
  }
}
