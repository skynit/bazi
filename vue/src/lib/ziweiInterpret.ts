export interface ZiWeiEvidenceLike {
  type: string
  label: string
  value: string
  basis?: string
}

export interface ZiWeiPlainReadingInput {
  palaceName: string
  palaceFocus?: string
  evidence?: ZiWeiEvidenceLike[]
  sanfangContext?: {
    opposite?: string
    trine1?: string
    trine2?: string
  } | null
}

export interface ZiWeiPlainPoint {
  key: string
  label: string
  text: string
  evidence: string
}

const PALACE_TOPICS: Record<string, string> = {
  命宫: '自我定位与做事方式',
  命宮: '自我定位与做事方式',
  兄弟宫: '同辈、协作与资源分配',
  兄弟宮: '同辈、协作与资源分配',
  夫妻宫: '亲密关系中的互动与协商',
  夫妻宮: '亲密关系中的互动与协商',
  子女宫: '创造、表达与带领后辈',
  子女宮: '创造、表达与带领后辈',
  财帛宫: '资源取得、使用与管理',
  財帛宮: '资源取得、使用与管理',
  疾厄宫: '生活节奏与身心照顾主题',
  疾厄宮: '生活节奏与身心照顾主题',
  迁移宫: '外部环境、出行与社会互动',
  遷移宮: '外部环境、出行与社会互动',
  仆役宫: '朋友、团队与合作关系',
  僕役宮: '朋友、团队与合作关系',
  交友宫: '朋友、团队与合作关系',
  交友宮: '朋友、团队与合作关系',
  官禄宫: '工作责任、组织角色与发展方向',
  官祿宮: '工作责任、组织角色与发展方向',
  事业宫: '工作责任、组织角色与发展方向',
  事業宮: '工作责任、组织角色与发展方向',
  田宅宫: '家庭、居住与长期安定感',
  田宅宮: '家庭、居住与长期安定感',
  福德宫: '兴趣、精神生活与内在节奏',
  福德宮: '兴趣、精神生活与内在节奏',
  父母宫: '长辈、制度与支持来源',
  父母宮: '长辈、制度与支持来源',
}

const STAR_THEMES: Record<string, string> = {
  紫微: '统筹、组织与主导',
  天府: '稳定、承接与储备',
  天机: '思考、规划与应变',
  太阳: '表达、担当与公开性',
  武曲: '执行、效率与资源管理',
  天同: '调和、体验与包容',
  廉贞: '原则、边界与选择',
  贪狼: '兴趣、社交与多样尝试',
  巨门: '沟通、辨析与质疑',
  天相: '协调、规则与支持',
  天梁: '保护、判断与责任',
  七杀: '决断、开创与压力应对',
  破军: '调整、突破与重建',
  太阴: '观察、细节与内在感受',
}

const STAR_ALIASES: Record<string, string> = {
  天機: '天机',
  太陽: '太阳',
  廉貞: '廉贞',
  貪狼: '贪狼',
  巨門: '巨门',
  七殺: '七杀',
  破軍: '破军',
  太陰: '太阴',
}

const AUX_THEMES: Record<string, string> = {
  左辅: '协作与支持',
  右弼: '协作与支持',
  文昌: '学习、文字与表达',
  文曲: '学习、审美与表达',
  天魁: '外部资源与协助',
  天钺: '外部资源与协助',
  禄存: '资源积累与守成',
  天马: '移动、变化与执行',
  擎羊: '阻力、竞争与推进方式',
  陀罗: '延迟、反复与耐力',
  火星: '突发节奏与快速反应',
  铃星: '压力信号与即时反应',
  地空: '落差、取舍与重新评估',
  地劫: '落差、取舍与资源边界',
}

const AUX_ALIASES: Record<string, string> = {
  左輔: '左辅',
  右弼: '右弼',
  天鉞: '天钺',
  祿存: '禄存',
  擎羊: '擎羊',
  陀羅: '陀罗',
  鈴星: '铃星',
  地劫: '地劫',
}

const HUA_THEMES: Record<string, string> = {
  化禄: '资源与承接',
  化祿: '资源与承接',
  化权: '责任、主导与执行',
  化權: '责任、主导与执行',
  化科: '学习、表达与认可',
  化忌: '阻力、执着与反复',
}

function cleanTopic(value?: string): string {
  return String(value || '')
    .replace(/[。；;]+$/g, '')
    .replace(/主题$/g, '')
    .trim()
}

function canonicalName(value: string, aliases: Record<string, string>): string {
  return aliases[value] || value
}

function matchedNames(
  values: string[],
  themes: Record<string, string>,
  aliases: Record<string, string>,
): string[] {
  const names = new Set<string>()
  const candidates = [...Object.keys(themes), ...Object.keys(aliases)]
  for (const value of values) {
    for (const candidate of candidates) {
      if (value.includes(candidate)) names.add(canonicalName(candidate, aliases))
    }
  }
  return [...names]
}

function unique(values: string[]): string[] {
  return [...new Set(values.filter(Boolean))]
}

function evidenceByType(evidence: ZiWeiEvidenceLike[], types: string[]): ZiWeiEvidenceLike[] {
  return evidence.filter((item) => types.includes(item.type))
}

function evidenceValues(items: ZiWeiEvidenceLike[]): string[] {
  return unique(items.map((item) => item.value).filter(Boolean))
}

function joinedEvidence(items: ZiWeiEvidenceLike[]): string {
  return evidenceValues(items).join('、')
}

export function ziweiPalaceTopic(palaceName: string, palaceFocus?: string): string {
  return cleanTopic(palaceFocus) || PALACE_TOPICS[palaceName] || '这个宫位所代表的生活主题'
}

export function buildZiweiPlainOverview(reading: ZiWeiPlainReadingInput): string {
  const evidence = reading.evidence || []
  const topic = ziweiPalaceTopic(reading.palaceName, reading.palaceFocus)
  const mainItems = evidenceByType(evidence, ['main_star'])
  const mainNames = matchedNames(evidenceValues(mainItems), STAR_THEMES, STAR_ALIASES)
  const borrowedItems = evidenceByType(evidence, ['borrowed_star'])

  if (mainNames.length) {
    const themes = unique(mainNames.map((name) => STAR_THEMES[name]))
    return `${reading.palaceName}主要看${topic}。当前以${mainNames.join('、')}为核心，传统上重点观察${themes.join('、')}这些方向。`
  }

  if (borrowedItems.length || reading.sanfangContext) {
    const context = reading.sanfangContext
    const related = unique([context?.opposite || '', context?.trine1 || '', context?.trine2 || ''])
    return `${reading.palaceName}主要看${topic}。本宫没有主星，需要连同${related.join('、') || '对宫与三合宫'}一起看；空宫不等于没有内容。`
  }

  return `${reading.palaceName}主要看${topic}。当前先按本宫已排出的星曜和四化理解，不延伸为具体事件。`
}

export function buildZiweiPlainPoints(evidence: ZiWeiEvidenceLike[] = []): ZiWeiPlainPoint[] {
  const points: ZiWeiPlainPoint[] = []
  const mainItems = evidenceByType(evidence, ['main_star'])
  const mainNames = matchedNames(evidenceValues(mainItems), STAR_THEMES, STAR_ALIASES)
  if (mainNames.length) {
    const descriptions = mainNames.map((name) => `${name}偏向${STAR_THEMES[name]}`)
    const hasBrightness = evidenceValues(mainItems).some((value) => /[庙廟旺得利平陷]/.test(value))
    points.push({
      key: 'main-star',
      label: '主星怎么读',
      text: `${descriptions.join('；')}。${hasBrightness ? '星曜后的“庙、旺、平、陷”等字是传统亮度等级，只说明表达是否更直接，不是吉凶分数。' : '这些词描述传统关注方向，不等于固定性格。'}`,
      evidence: joinedEvidence(mainItems),
    })
  }

  const borrowedItems = evidenceByType(evidence, ['borrowed_star'])
  if (borrowedItems.length) {
    points.push({
      key: 'borrowed-star',
      label: '空宫怎么读',
      text: '本宫没有主星时，会参考对宫与三合宫的主星。这里的“借星”是补充阅读上下文，不是把别宫结论直接搬过来。',
      evidence: joinedEvidence(borrowedItems),
    })
  }

  const huaItems = evidenceByType(evidence, ['four_hua'])
  const huaNames = unique(
    evidenceValues(huaItems).flatMap((value) =>
      Object.keys(HUA_THEMES).filter((name) => value.includes(name)),
    ),
  )
  if (huaNames.length) {
    const themes = unique(huaNames.map((name) => HUA_THEMES[name]))
    points.push({
      key: 'four-hua',
      label: '四化在说什么',
      text: `本宫四化把注意力引向${themes.join('、')}。它描述这个宫位较容易被反复关注的主题，不代表现实结果必然发生。`,
      evidence: joinedEvidence(huaItems),
    })
  }

  const supportItems = evidenceByType(evidence, ['soft_star', 'tough_star', 'aux_star'])
  const supportNames = matchedNames(evidenceValues(supportItems), AUX_THEMES, AUX_ALIASES)
  if (supportNames.length) {
    const themes = unique(supportNames.map((name) => AUX_THEMES[name]))
    points.push({
      key: 'support-stars',
      label: '辅助星曜怎么读',
      text: `这些星曜补充了${themes.join('、')}等观察角度。辅曜和煞曜都是传统分类标签，需要与主星、四化和三方四正合看。`,
      evidence: joinedEvidence(supportItems),
    })
  }

  const bodyItems = evidenceByType(evidence, ['body_palace'])
  if (bodyItems.length) {
    points.push({
      key: 'body-palace',
      label: '身宫提示',
      text: '身宫落在这里，传统上会提高这个宫位在实际行动层面的阅读权重；它不是职业、关系或人生结果的直接结论。',
      evidence: joinedEvidence(bodyItems),
    })
  }

  return points
}
