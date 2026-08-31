/**
 * 命盘页「十神结构」分区的白话解读文案。
 *
 * 与 fortuneCopy.ts 的「今日行动建议」分工不同：那边是"今天适合做什么"，
 * 这里是"这个命盘的次数统计呈现出的倾向"。关键词复用 TEN_GOD_GUIDE，
 * 保持全站十神关键词一致。
 * 口径约束：传统取象的生活化翻译，描述倾向而非定论，不作性格/吉凶断语。
 */

import { TEN_GOD_GUIDE } from './fortuneCopy'

/** 十神 → 特质长处 + 留意提醒（各一句大白话） */
export interface TenGodTraitCopy {
  /** 特质长处 */
  trait: string
  /** 留意提醒 */
  caution: string
}

export const TEN_GOD_TRAIT: Record<string, TenGodTraitCopy> = {
  比肩: {
    trait: '独立有主见，同辈朋友往往是你的助力',
    caution: '太坚持己见的时候，不妨听听别人的意见',
  },
  劫财: {
    trait: '敢闯敢拼、行动力强，对朋友讲义气',
    caution: '涉及钱和合作时多留个心眼，别冲动决策',
  },
  食神: {
    trait: '有才华、表达温和，懂得享受生活',
    caution: '别太安逸，给自己定个小目标',
  },
  伤官: {
    trait: '聪明外露、敢打破常规，常有新点子',
    caution: '说话直容易得罪人，注意分寸',
  },
  正财: {
    trait: '踏实务实、善于积累，财务观念稳',
    caution: '偶尔偏于保守，机会来了别错过',
  },
  偏财: {
    trait: '人缘广、机会多，财路灵活',
    caution: '花销大的时候记得收一收',
  },
  正官: {
    trait: '做事有章法、责任心强，容易获得上级信任',
    caution: '有时偏于拘谨，别给自己太大压力',
  },
  七杀: {
    trait: '有魄力、抗压强，能啃硬骨头',
    caution: '压力别攒太多，给自己留个出口',
  },
  正印: {
    trait: '爱学习、贵人缘好，容易得到长辈前辈帮衬',
    caution: '想得多做得少的时候，提醒自己先落地',
  },
  偏印: {
    trait: '直觉灵、点子多，适合钻研冷门或创意领域',
    caution: '容易想太多，注意劳逸结合',
  },
}

/** 十神关键词（复用今日运势的关键词表，全站一致） */
export function tenGodKeyword(name: string): string {
  return TEN_GOD_GUIDE[name]?.keyword ?? ''
}

/** 未出现十神的一句带过表述：hedged，不把"未出现"说成短板 */
export function absentTenGodSentence(names: string[]): string {
  if (!names.length) return ''
  return `${names.join('、')}在这个统计里没有出现——不代表缺乏这些特质，只是命盘里这类信号较弱。`
}
