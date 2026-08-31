/**
 * 今日运势概览的白话文案映射。
 *
 * 把后端返回的结构化字段（生扶/冲克条数、结构指数、十神、十二长生、月令状态）
 * 翻译成不依赖命理术语的生活化描述。术语只以括号对照形式出现。
 * 口径约束（见 API.md）：只描述状态与节奏倾向，不表述吉凶、强弱或事件概率。
 */

/** 结构关系基调分支 */
export type ToneBranch = 'support' | 'clash' | 'mixed' | 'none'

export function toneBranchOf(supporting: number, counter: number): ToneBranch {
  if (supporting > 0 && counter === 0) return 'support'
  if (counter > 0 && supporting === 0) return 'clash'
  if (supporting > 0 && counter > 0) return 'mixed'
  return 'none'
}

export interface ToneCopy {
  /** 衬线大题：一句话告诉用户今天整体感觉 */
  headline: string
  /** 一两句白话解释"为什么"，术语紧跟白话括号 */
  why: string
}

export function toneCopyOf(branch: ToneBranch, supporting: number, counter: number): ToneCopy {
  switch (branch) {
    case 'support':
      return {
        headline: '今天助力多、阻力少',
        why: `今天的日子和你的命盘之间，记录到 ${supporting} 条支持、助力型互动（命理叫「生扶」），没有冲突类信号。适合推进需要别人配合、借力完成的事。`,
      }
    case 'clash':
      return {
        headline: '今天容易有摩擦和变动',
        why: `今天的日子和你的命盘之间，记录到 ${counter} 条冲撞、消耗型互动（命理叫「冲克」），没有助力类信号。重要安排多留缓冲，遇到摩擦不必硬顶。`,
      }
    case 'mixed':
      return {
        headline: '今天机会与磕绊并存',
        why: `今天的日子和你的命盘之间，既有 ${supporting} 条支持型互动（生扶），也有 ${counter} 条冲撞型互动（冲克）。分清轻重：能借力的事往前推，容易起摩擦的事缓一缓。`,
      }
    case 'none':
      return {
        headline: '今天平平常常，按自己的节奏来',
        why: '今天的日子和你的命盘之间，没有记录到明显的助力或冲突信号。算是平常的一天，照自己的计划安排就好。',
      }
  }
}

/** 状态分分档：以中性基线为锚，只描述状态节奏，不表述吉凶 */
export interface ScoreTier {
  label: string
  caption: string
}

export function scoreTierOf(score: number, base: number): ScoreTier {
  const diff = score - base
  let label: string
  if (diff <= -5) label = '偏低'
  else if (diff < 5) label = '平常'
  else if (diff <= 20) label = '偏好'
  else label = '很好'
  return {
    label,
    caption: `${label} · 平常日约 ${base}`,
  }
}

/** 十神 → 关键词 + 行动建议（传统取象的生活化翻译，非吉凶判断） */
export interface TenGodGuide {
  keyword: string
  advice: string
}

export const TEN_GOD_GUIDE: Record<string, TenGodGuide> = {
  比肩: {
    keyword: '协作与自立',
    advice: '适合约同事、朋友一起推进事情，也适合自己独立完成任务。',
  },
  劫财: {
    keyword: '行动与竞争',
    advice: '适合主动争取、快速执行；涉及分钱、合伙、大额支出多留个心眼。',
  },
  食神: {
    keyword: '表达与创造',
    advice: '适合写作、输出想法、发挥手艺，也适合做点好吃的犒劳自己。',
  },
  伤官: {
    keyword: '才华与突破',
    advice: '适合展示作品、打破常规想新招；表达观点时留意场合与分寸。',
  },
  正财: {
    keyword: '务实与经营',
    advice: '适合处理账目、落实收入，把务实的计划往前推一步。',
  },
  偏财: {
    keyword: '机会与人脉',
    advice: '适合社交应酬、拓展资源，留意灵活的合作与进账机会。',
  },
  正官: {
    keyword: '规则与责任',
    advice: '适合处理正式事务、走流程，与上级或制度相关的事今天比较顺路。',
  },
  七杀: {
    keyword: '压力与魄力',
    advice: '适合啃硬骨头、快速决断；节奏偏紧，记得给自己留减压出口。',
  },
  正印: {
    keyword: '学习与沉淀',
    advice: '适合读书、上课、整理资料，向长辈或前辈请教容易有收获。',
  },
  偏印: {
    keyword: '钻研与灵感',
    advice: '适合学习、研究、琢磨新东西，灵光一现的想法值得记下来。',
  },
}

/** 十二长生 → 白话节奏描述 */
export const TWELVE_STAGE_PLAIN: Record<string, string> = {
  长生: '起步发力期',
  沐浴: '调整适应期',
  冠带: '稳步上升期',
  临官: '状态上行期',
  帝旺: '状态高峰期',
  衰: '缓步回落期',
  病: '状态偏弱期',
  死: '沉静收敛期',
  墓: '沉淀收纳期',
  绝: '转换酝酿期',
  胎: '萌芽孕育期',
  养: '蓄力准备期',
}

/** 月令状态（旺相休囚死）→ 白话大环境描述 */
export const SEASONAL_STATE_PLAIN: Record<string, string> = {
  旺: '顺势有力',
  相: '平稳偏顺',
  休: '平缓休整',
  囚: '施展空间偏小',
  死: '低迷，宜守不宜攻',
}
