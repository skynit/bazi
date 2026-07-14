import type { FortuneGuide } from '../api/fortune'
import earthBackdrop from '../assets/background/土.png'
import fireBackdrop from '../assets/background/火.png'
import metalBackdrop from '../assets/background/金.png'
import waterBackdrop from '../assets/background/水.png'
import woodBackdrop from '../assets/background/木.png'
import earthImage from '../assets/blessing/earth.png'
import fireImage from '../assets/blessing/fire.png'
import metalImage from '../assets/blessing/metal.png'
import waterImage from '../assets/blessing/water.png'
import woodImage from '../assets/blessing/wood.png'

export type BlessingElement = '木' | '火' | '土' | '金' | '水'

export interface BlessingProfile {
  element: BlessingElement
  image: string
  backdrop: string
  accent: string
  accentDark: string
  accentRgb: string
  colors: string
  objectLabel: string
  objects: string
  direction: string
  avoidDirection: string
  alt: string
}

const elementKeys: BlessingElement[] = ['木', '火', '土', '金', '水']

export const blessingProfiles: Record<BlessingElement, BlessingProfile> = {
  木: {
    element: '木',
    image: woodImage,
    backdrop: woodBackdrop,
    accent: '#22c59e',
    accentDark: '#0f8f6e',
    accentRgb: '34, 197, 158',
    colors: '绿色系、青色系',
    objectLabel: '绿植',
    objects: '绿植、木质文具、清晨步行',
    direction: '东方',
    avoidDirection: '西方',
    alt: '国风二次元风格的青绿色绿植与木质文具插画',
  },
  火: {
    element: '火',
    image: fireImage,
    backdrop: fireBackdrop,
    accent: '#fb7185',
    accentDark: '#be123c',
    accentRgb: '251, 113, 133',
    colors: '红色系、紫色系',
    objectLabel: '暖光',
    objects: '暖光、红紫配饰、公开表达',
    direction: '南方',
    avoidDirection: '北方',
    alt: '国风二次元风格的红紫暖光台灯与配饰插画',
  },
  土: {
    element: '土',
    image: earthImage,
    backdrop: earthBackdrop,
    accent: '#f2bd4d',
    accentDark: '#b7791f',
    accentRgb: '242, 189, 77',
    colors: '黄色系、棕色系',
    objectLabel: '陶瓷',
    objects: '陶瓷、黄色便签、整理桌面',
    direction: '中宫',
    avoidDirection: '东方',
    alt: '国风二次元风格的陶瓷杯与黄色便签插画',
  },
  金: {
    element: '金',
    image: metalImage,
    backdrop: metalBackdrop,
    accent: '#cbd5e1',
    accentDark: '#64748b',
    accentRgb: '203, 213, 225',
    colors: '白色系、金色系',
    objectLabel: '金属笔',
    objects: '金属笔、白金色配饰、清单工具',
    direction: '西方',
    avoidDirection: '南方',
    alt: '国风二次元风格的金属笔与清单工具插画',
  },
  水: {
    element: '水',
    image: waterImage,
    backdrop: waterBackdrop,
    accent: '#22d3ee',
    accentDark: '#2563eb',
    accentRgb: '34, 211, 238',
    colors: '黑色系、蓝色系',
    objectLabel: '水杯',
    objects: '水杯、蓝黑色配饰、复盘记录',
    direction: '北方',
    avoidDirection: '中宫',
    alt: '国风二次元风格的蓝黑色水杯与记录本插画',
  },
}

export function isBlessingElement(value?: string): value is BlessingElement {
  return !!value && elementKeys.includes(value as BlessingElement)
}

export function elementFromColorText(text?: string): BlessingElement | undefined {
  if (!text) return undefined
  if (/青|绿|翠|木/.test(text)) return '木'
  if (/红|紫|朱|火/.test(text)) return '火'
  if (/黄|棕|咖|褐|土/.test(text)) return '土'
  if (/白|金|银|灰/.test(text)) return '金'
  if (/黑|蓝|水/.test(text)) return '水'
  return undefined
}

export function resolveBlessingElement(guide?: FortuneGuide, luckyColor?: string): BlessingElement {
  const guideColor = guide?.lucky_colors?.find((item) => isBlessingElement(item.element))
  if (isBlessingElement(guideColor?.element)) return guideColor.element
  if (isBlessingElement(guide?.primary_element)) return guide.primary_element
  if (isBlessingElement(guide?.secondary_element)) return guide.secondary_element

  const colorText = guide?.lucky_colors?.[0]?.value || luckyColor
  return elementFromColorText(colorText) ?? '土'
}

export function splitGuideValues(value?: string) {
  return (value ?? '')
    .split(/[、,，+＋]/)
    .map((item) => item.trim())
    .filter(Boolean)
}
