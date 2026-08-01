import type { FortuneDay } from '../api/fortune'

export type BlessingElement = '木' | '火' | '土' | '金' | '水'

export interface BlessingAction {
  title: string
  description: string
  image: string
}

export interface BlessingProfile {
  element: BlessingElement
  image: string
  backdrop: string
  backdropHd: string
  accent: string
  accentDark: string
  accentRgb: string
  tone: string
  summary: string
  colors: string
  objectLabel: string
  objects: string
  objectPrompt: string
  direction: string
  avoidDirection: string
  alt: string
  ritualImages: string[]
  actions: BlessingAction[]
}

const elements: BlessingElement[] = ['木', '火', '土', '金', '水']

export const blessingProfiles: Record<BlessingElement, BlessingProfile> = {
  木: {
    element: '木',
    image: '/element-assets/wood/wood-desk-plant.webp',
    backdrop: '/element-assets/wood/wood-forest-dawn.webp',
    backdropHd: '/element-assets/wood/wood-forest-dawn-hd.webp',
    accent: '#22a982',
    accentDark: '#0f7d62',
    accentRgb: '34, 169, 130',
    tone: '生发',
    summary: '给正在生长的事情留出空间，先完成一件能推动后续的小事。',
    colors: '青色、绿色',
    objectLabel: '绿植与木质文具',
    objects: '绿植、木质文具、清晨步行',
    objectPrompt: '把一件常用的木质或绿色物件放到视线可及处。',
    direction: '东方',
    avoidDirection: '西方',
    alt: '晨光中的绿植与木质文具',
    ritualImages: [
      '/element-assets/wood/wood-sprout.webp',
      '/element-assets/wood/wood-bamboo-mist.webp',
      '/element-assets/wood/wood-tea-table.webp',
    ],
    actions: [
      { title: '先理枝叶', description: '删去一项不必要的安排，让今天的主线更清楚。', image: '/element-assets/wood/wood-rain-leaf.webp' },
      { title: '推进一步', description: '选择一件可以在二十分钟内取得进展的事情。', image: '/element-assets/wood/wood-sprout.webp' },
      { title: '短暂步行', description: '离开屏幕走一小段路，再回来处理需要判断的工作。', image: '/element-assets/wood/wood-mountain.webp' },
    ],
  },
  火: {
    element: '火',
    image: '/element-assets/fire/fire-warm-desk.webp',
    backdrop: '/element-assets/fire/fire-sunrise.webp',
    backdropHd: '/element-assets/fire/fire-sunrise-hd.webp',
    accent: '#df5a68',
    accentDark: '#ad3345',
    accentRgb: '223, 90, 104',
    tone: '明朗',
    summary: '把注意力集中到需要表达和确认的事情上，清楚比热闹更重要。',
    colors: '朱红、暖紫',
    objectLabel: '暖光与红色小物',
    objects: '暖光、红色小物、清晰表达',
    objectPrompt: '调整桌面光线，用一处温暖颜色提醒自己保持专注。',
    direction: '南方',
    avoidDirection: '北方',
    alt: '暖光书桌与红色小物',
    ritualImages: [
      '/element-assets/fire/fire-candle.webp',
      '/element-assets/fire/fire-lantern.webp',
      '/element-assets/fire/fire-tea-light.webp',
    ],
    actions: [
      { title: '说清重点', description: '把最重要的一句话先写下来，再开始沟通。', image: '/element-assets/fire/fire-kiln.webp' },
      { title: '控制节奏', description: '重要回复先停一分钟，确认语气后再发送。', image: '/element-assets/fire/fire-tea-light.webp' },
      { title: '整理光线', description: '减少刺眼反光，让工作区域保持明亮但不过度刺激。', image: '/element-assets/fire/fire-warm-desk.webp' },
    ],
  },
  土: {
    element: '土',
    image: '/element-assets/earth/earth-organized-desk.webp',
    backdrop: '/element-assets/earth/earth-courtyard.webp',
    backdropHd: '/element-assets/earth/earth-courtyard-hd.webp',
    accent: '#b58a35',
    accentDark: '#8a6422',
    accentRgb: '181, 138, 53',
    tone: '承载',
    summary: '先稳定基础和秩序，今天适合把散落的事项重新放回正确位置。',
    colors: '赭黄、陶褐',
    objectLabel: '陶器与便签',
    objects: '陶器、便签、整洁桌面',
    objectPrompt: '清理手边的一小块区域，只保留今天真正会用到的东西。',
    direction: '中宫',
    avoidDirection: '东方',
    alt: '陶器、便签与整洁桌面',
    ritualImages: [
      '/element-assets/earth/earth-ceramic.webp',
      '/element-assets/earth/earth-stone-path.webp',
      '/element-assets/earth/earth-tea-stone.webp',
    ],
    actions: [
      { title: '归拢事项', description: '把零散任务收进一个清单，并只标出一个当前动作。', image: '/element-assets/earth/earth-organized-desk.webp' },
      { title: '完成收尾', description: '优先处理已经开始但尚未结束的一件小事。', image: '/element-assets/earth/earth-stone-path.webp' },
      { title: '按时进食', description: '给身体留出稳定节奏，避免用忙碌替代基本照顾。', image: '/element-assets/earth/earth-wheat-field.webp' },
    ],
  },
  金: {
    element: '金',
    image: '/element-assets/metal/metal-stationery.webp',
    backdrop: '/element-assets/metal/metal-white-mountain.webp',
    backdropHd: '/element-assets/metal/metal-white-mountain-hd.webp',
    accent: '#7b8796',
    accentDark: '#596574',
    accentRgb: '123, 135, 150',
    tone: '收敛',
    summary: '明确边界、精简步骤，把判断写成可以执行的规则。',
    colors: '银白、浅金',
    objectLabel: '金属笔与清单',
    objects: '金属笔、清单、简洁器物',
    objectPrompt: '用一张简短清单区分必须完成、可以延后和无需处理。',
    direction: '西方',
    avoidDirection: '南方',
    alt: '银灰色金属文具与清单',
    ritualImages: [
      '/element-assets/metal/metal-compass.webp',
      '/element-assets/metal/metal-white-stones.webp',
      '/element-assets/metal/metal-rain-bell.webp',
    ],
    actions: [
      { title: '划清边界', description: '为一项容易拖长的任务设定明确结束条件。', image: '/element-assets/metal/metal-compass.webp' },
      { title: '删去重复', description: '合并重复信息，减少一次不必要的往返确认。', image: '/element-assets/metal/metal-white-stones.webp' },
      { title: '记录决定', description: '把已确定的事项写下来，避免重新消耗判断力。', image: '/element-assets/metal/metal-stationery.webp' },
    ],
  },
  水: {
    element: '水',
    image: '/element-assets/water/water-cup-notebook.webp',
    backdrop: '/element-assets/water/water-river-valley.webp',
    backdropHd: '/element-assets/water/water-river-valley-hd.webp',
    accent: '#347f9f',
    accentDark: '#255f79',
    accentRgb: '52, 127, 159',
    tone: '流动',
    summary: '先观察信息如何流动，再决定从哪里进入，避免在不清楚时强推。',
    colors: '玄黑、深蓝',
    objectLabel: '水杯与记录本',
    objects: '水杯、记录本、安静复盘',
    objectPrompt: '把水和记录工具放在手边，给思考留一段不被打断的时间。',
    direction: '北方',
    avoidDirection: '中宫',
    alt: '深蓝水杯与记录本',
    ritualImages: [
      '/element-assets/water/water-ink-bowl.webp',
      '/element-assets/water/water-rain-window.webp',
      '/element-assets/water/water-tea-steam.webp',
    ],
    actions: [
      { title: '先听后答', description: '在回应之前复述一次对方的核心问题。', image: '/element-assets/water/water-ripples.webp' },
      { title: '留出缓冲', description: '两项任务之间保留短暂空档，避免注意力直接碰撞。', image: '/element-assets/water/water-rain-window.webp' },
      { title: '复盘一页', description: '写下今天已知、未知和下一步各一条。', image: '/element-assets/water/water-ink-bowl.webp' },
    ],
  },
}

export function isBlessingElement(value?: string): value is BlessingElement {
  return !!value && elements.includes(value as BlessingElement)
}

export function resolveBlessingElement(fortune?: FortuneDay | null): BlessingElement {
  if (
    fortune?.seasonal_state?.status === 'observed' &&
    isBlessingElement(fortune.seasonal_state.query_element)
  ) {
    return fortune.seasonal_state.query_element
  }

  const distribution = fortune?.today_elements ?? {}
  const dominant = Object.entries(distribution)
    .filter(([element]) => isBlessingElement(element))
    .sort((left, right) => right[1] - left[1])[0]?.[0]
  if (isBlessingElement(dominant)) return dominant

  if (
    fortune?.season_element?.status === 'observed' &&
    isBlessingElement(fortune.season_element.reference_element)
  ) {
    return fortune.season_element.reference_element
  }
  return '土'
}
