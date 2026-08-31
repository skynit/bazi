/**
 * 命盘页术语白话映射（glossary）。
 *
 * 统一全站术语翻译口径：命理术语首次出现时给一句初中生能读懂的白话解释。
 * 与 fortuneCopy.ts / tenGodInterpret.ts 同一语气基准：像懂命理的朋友聊天，
 * 直白、具体、不绝对化；只说"传统上视为 / 可理解为"，不下吉凶断语。
 */

/** 四柱 → 传统定位（人生阶段），用于四柱卡片标题小字 */
export const PILLAR_ROLE: Record<string, string> = {
  year: '祖上与少年',
  month: '父母与青年',
  day: '自己与中年',
  hour: '子女与晚年',
}

/** 通用术语 → 一句白话解释（分区说明、括号小字共用） */
export const GLOSSARY: Record<string, string> = {
  四柱: '出生年月日时换算成的四组干支，每柱两个字',
  日主: '日柱的天干，代表你自己',
  十神: '以日主为参照，给其他天干贴的角色标签',
  透干: '藏在地支里的天干浮到天干上，叫"透出"',
  藏干: '每个地支内部"藏着"的天干',
  纳音: '传统对干支组合的另一种分类叫法，供参考',
  大运: '十年一步的人生大周期，看每个阶段的整体节奏',
  流年: '这一年的干支，用来看这一年的节奏',
  流月: '这一月的干支，用来看这一月的节奏',
  小运: '一年一换的辅助周期，比大运更细',
  月令: '出生月份的地支，传统上视为命盘的"季节背景"',
  命宫: '按出生时辰推算的"第五柱"，传统上用来辅助看整体倾向',
  调候: '传统上认为命局需要调和的五行',
  身强: '日主力量偏强（候选判断，未独立验证）',
  身弱: '日主力量偏弱（候选判断，未独立验证）',
  格局: '传统对命盘整体结构的归类叫法',
  空亡: '传统上视为"落空、打折扣"的位置标记',
}

/** 天干关系类型 → 白话解释 */
export const GAN_RELATION_PLAIN: Record<string, string> = {
  五合: '两个天干按传统配对相合，传统上视为容易合拍、互相吸引',
  相生: '一方五行生助另一方，传统上视为助力',
  相克: '一方五行克制另一方，传统上视为牵制',
  比和: '同五行的天干相遇，传统上视为气场相投',
  天干相冲: '两个天干正面对冲，传统上视为容易顶撞',
}

/** 地支关系类型 → 白话解释 */
export const ZHI_RELATION_PLAIN: Record<string, string> = {
  六冲: '两个地支正面对冲，传统上视为容易起变化、起摩擦',
  六合: '两个地支按传统配对相合，传统上视为关系融洽',
  半合: '三合局中的两支相遇，合局之力尚缺一支',
  拱合: '两支夹出中间一支，传统上视为暗中成合',
  三合局: '三支凑齐一个合局，传统上视为气力汇聚',
  六害: '两个地支相害，传统上视为暗中损耗',
  六破: '两个地支相破，传统上视为小有破损',
  相刑: '两个地支相刑，传统上视为互相折腾',
  三刑: '三支构成相刑，传统上视为压力叠加',
  半会: '三支会局中的两支相遇，会局之力尚缺一支',
  三会局: '三支会齐同一方位，传统上视为同气汇聚',
  伏吟: '相同干支重复出现，传统上视为事情容易反复',
}

export function ganRelationPlain(type?: string): string {
  return type ? GAN_RELATION_PLAIN[type] || '' : ''
}

export function zhiRelationPlain(type?: string): string {
  return type ? ZHI_RELATION_PLAIN[type] || '' : ''
}

/** 常见神煞 → 传统寓意白话（覆盖 20+ 常见项，未收录走默认句式） */
export const SHEN_SHA_MEANING: Record<string, string> = {
  天乙贵人: '传统中的贵人运象征，遇事易得帮助',
  天德贵人: '传统中逢凶化吉的标记',
  月德贵人: '传统中逢凶化吉的标记',
  太极贵人: '传统中主聪慧好学、喜探究的标记',
  文昌贵人: '传统中与学业、文书、考试相关的标记',
  学堂: '传统中主学习、进修的标记',
  词馆: '传统中主文才、表达的标记',
  驿马: '走动、变动、出行',
  桃花: '人缘与异性缘',
  咸池: '人缘与异性缘（桃花的另一种叫法）',
  红鸾: '传统中的婚恋喜兆标记',
  天喜: '传统中的喜庆、好人缘标记',
  华盖: '传统中主文艺、玄学、喜静独处的标记',
  将星: '传统中主领导力、掌权的标记',
  金匮: '传统中与财库、积蓄相关的标记',
  金舆: '传统中与出行、车驾相关的福分标记',
  禄神: '传统中与稳定收入、俸禄相关的标记',
  羊刃: '传统中主刚强、冲劲的标记，也提示易有磕碰',
  劫煞: '传统中主波折、破耗的标记',
  灾煞: '传统中主意外、损失的标记',
  亡神: '传统中主心神不定、暗中损耗的标记',
  孤辰: '传统中主独来独往的标记',
  寡宿: '传统中主清静、少热闹的标记',
  空亡: '传统上视为"落空、打折扣"的位置标记',
  天医: '传统中与健康、医药缘分相关的标记',
  元辰: '传统中主变动、反复的标记',
  勾绞: '传统中主纠缠、是非的标记',
  丧门: '传统中主孝服、哀事的标记',
  吊客: '传统中主吊唁、哀事的标记',
  披麻: '传统中主孝服的标记',
  天罗地网: '传统中主束缚、受限的标记',
  国印贵人: '传统中主诚信、掌权印信的标记',
  德秀贵人: '传统中主品德、才华的吉庆标记',
  福星贵人: '传统中主福气、安稳的标记',
  天厨贵人: '传统中与饮食、福禄相关的标记',
  魁罡: '传统中主聪明果断、性格刚烈的标记',
  月煞: '传统中主小病小灾的标记',
  十恶大败: '传统中主破财、损耗的忌讳标记',
  雷霆煞: '传统中主惊扰、突发的标记',
  六厄: '传统中主困顿、受阻的标记',
  白虎: '传统中主刚猛、易有血光磕碰的标记',
  官符: '传统中主官非、文书的标记',
  病符: '传统中与小病小痛相关的标记',
  大耗: '传统中主大额破耗的标记',
  小耗: '传统中主小额破耗的标记',
  五鬼: '传统中主是非、暗耗的标记',
}

/** 神煞白话寓意：优先查专表，未收录时用默认句式 */
export function shenShaMeaning(name: string): string {
  return SHEN_SHA_MEANING[name] || '传统查表命中的标记，暂无固定白话寓意'
}

function joinNames(names: string[]): string {
  if (names.length <= 1) return names[0] || ''
  return `${names.slice(0, -1).join('、')}和${names[names.length - 1]}`
}

/**
 * 五行分布 → 数据驱动的大白话小结。
 * 以五行的平均分值为基准：≥1.25 倍均值算"偏多"，≤0.75 倍算"偏少"，为 0 单独说明；
 * 差异都不大时说"大体均匀"。只描述分布，不评价好坏。
 */
export function elementBalanceSentence(scores: Record<string, number>): string {
  const labels = ['木', '火', '土', '金', '水']
  const entries = labels.map((name) => ({ name, value: Number(scores[name] || 0) }))
  const total = entries.reduce((sum, item) => sum + item.value, 0)
  if (total <= 0) return ''
  const avg = total / labels.length
  const high = entries.filter((item) => item.value >= avg * 1.25).map((item) => item.name)
  const low = entries
    .filter((item) => item.value > 0 && item.value <= avg * 0.75)
    .map((item) => item.name)
  const zero = entries.filter((item) => item.value <= 0).map((item) => item.name)

  const parts: string[] = []
  if (high.length) parts.push(`${joinNames(high)}偏多`)
  if (low.length) parts.push(`${joinNames(low)}偏少`)
  if (zero.length) parts.push(`${joinNames(zero)}基本没有出现`)
  if (!parts.length) return '你的命盘里，五行分布大体均匀。'
  return `你的命盘里，${parts.join('，')}。`
}
