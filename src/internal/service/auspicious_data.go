package service

// AuspiciousData provides hardcoded fortune lookup rules for
// lucky colors, numbers, wealth directions, clash zodiacs,
// auspicious hours, and activity catalogs.
//
// All rules are derived from traditional Chinese HuangLi (黄历) principles.
// This is a write-only data store; no database connection is required.
type AuspiciousData struct {
	LuckColorRules  map[string]string
	LuckNumberRules map[string][]int
	WealthDirection map[string]string
	ClashZodiac     map[string]string
	AuspiciousHours map[string][]string
	ActivityCatalog map[string][]string // "宜" or "忌" → list of activity names
}

// NewAuspiciousData creates a fully-populated AuspiciousData instance
// with hardcoded rules covering all 10 heavenly stems and 12 earthly branches.
func NewAuspiciousData() *AuspiciousData {
	return &AuspiciousData{
		LuckColorRules:  buildLuckColorRules(),
		LuckNumberRules: buildLuckNumberRules(),
		WealthDirection: buildWealthDirection(),
		ClashZodiac:     buildClashZodiac(),
		AuspiciousHours: buildAuspiciousHours(),
		ActivityCatalog: buildActivityCatalog(),
	}
}

// GetLuckyColor returns the lucky color description for a given heavenly stem.
// Returns empty string if the stem is not recognized.
func (d *AuspiciousData) GetLuckyColor(stem string) string {
	return d.LuckColorRules[stem]
}

// GetLuckyNumbers returns the lucky numbers for a given heavenly stem.
// Returns nil if the stem is not recognized.
func (d *AuspiciousData) GetLuckyNumbers(stem string) []int {
	return d.LuckNumberRules[stem]
}

// GetWealthDirection returns the wealth direction for a given heavenly stem.
// Returns empty string if the stem is not recognized.
func (d *AuspiciousData) GetWealthDirection(stem string) string {
	return d.WealthDirection[stem]
}

// GetClashZodiac returns the clash zodiac animal for a given earthly branch.
// Returns empty string if the branch is not recognized.
func (d *AuspiciousData) GetClashZodiac(branch string) string {
	return d.ClashZodiac[branch]
}

// GetAuspiciousHours returns the auspicious time ranges for a given earthly branch.
// Returns nil if the branch is not recognized.
func (d *AuspiciousData) GetAuspiciousHours(branch string) []string {
	return d.AuspiciousHours[branch]
}

// GetActivities returns the list of activities for a given category.
// Category must be "宜" (auspicious) or "忌" (inauspicious).
// Returns nil for unrecognized categories.
func (d *AuspiciousData) GetActivities(category string) []string {
	return d.ActivityCatalog[category]
}

// --- Hardcoded data builders ---

// buildLuckColorRules returns lucky colors by heavenly stem.
// 甲乙→绿/青(木), 丙丁→红/紫(火), 戊己→黄/棕(土),
// 庚辛→白/金(金), 壬癸→黑/蓝(水)
func buildLuckColorRules() map[string]string {
	return map[string]string{
		"甲": "绿色系",
		"乙": "青色系",
		"丙": "红色系",
		"丁": "紫色系",
		"戊": "黄色系",
		"己": "棕色系",
		"庚": "白色系",
		"辛": "金色系",
		"壬": "黑色系",
		"癸": "蓝色系",
	}
}

// buildLuckNumberRules returns lucky numbers by heavenly stem.
// 木→3,8; 火→2,7; 土→5,0; 金→4,9; 水→1,6
func buildLuckNumberRules() map[string][]int {
	return map[string][]int{
		"甲": {3, 8},
		"乙": {3, 8},
		"丙": {2, 7},
		"丁": {2, 7},
		"戊": {5, 0},
		"己": {5, 0},
		"庚": {4, 9},
		"辛": {4, 9},
		"壬": {1, 6},
		"癸": {1, 6},
	}
}

// buildWealthDirection returns wealth direction by heavenly stem.
// 甲乙→东北, 丙丁→正西, 戊己→正北, 庚辛→正东, 壬癸→正南
func buildWealthDirection() map[string]string {
	return map[string]string{
		"甲": "东北",
		"乙": "东北",
		"丙": "正西",
		"丁": "正西",
		"戊": "正北",
		"己": "正北",
		"庚": "正东",
		"辛": "正东",
		"壬": "正南",
		"癸": "正南",
	}
}

// buildClashZodiac returns the clash (六冲) zodiac for each earthly branch.
// 子→午马, 丑→未羊, 寅→申猴, 卯→酉鸡, 辰→戌狗, 巳→亥猪,
// 午→子鼠, 未→丑牛, 申→寅虎, 酉→卯兔, 戌→辰龙, 亥→巳蛇
func buildClashZodiac() map[string]string {
	return map[string]string{
		"子": "马",
		"丑": "羊",
		"寅": "猴",
		"卯": "鸡",
		"辰": "狗",
		"巳": "猪",
		"午": "鼠",
		"未": "牛",
		"申": "虎",
		"酉": "兔",
		"戌": "龙",
		"亥": "蛇",
	}
}

// buildAuspiciousHours returns 2-3 auspicious two-hour periods for each earthly branch.
// Each entry describes a traditional Chinese double-hour (时辰).
func buildAuspiciousHours() map[string][]string {
	return map[string][]string{
		"子": {"子时(23:00-01:00)", "丑时(01:00-03:00)"},
		"丑": {"寅时(03:00-05:00)", "卯时(05:00-07:00)", "巳时(09:00-11:00)"},
		"寅": {"子时(23:00-01:00)", "丑时(01:00-03:00)", "辰时(07:00-09:00)"},
		"卯": {"寅时(03:00-05:00)", "午时(11:00-13:00)", "未时(13:00-15:00)"},
		"辰": {"子时(23:00-01:00)", "辰时(07:00-09:00)", "申时(15:00-17:00)"},
		"巳": {"丑时(01:00-03:00)", "午时(11:00-13:00)", "辰时(07:00-09:00)"},
		"午": {"寅时(03:00-05:00)", "未时(13:00-15:00)", "酉时(17:00-19:00)"},
		"未": {"子时(23:00-01:00)", "卯时(05:00-07:00)", "申时(15:00-17:00)"},
		"申": {"辰时(07:00-09:00)", "未时(13:00-15:00)", "戌时(19:00-21:00)"},
		"酉": {"寅时(03:00-05:00)", "午时(11:00-13:00)", "酉时(17:00-19:00)"},
		"戌": {"卯时(05:00-07:00)", "巳时(09:00-11:00)", "申时(15:00-17:00)"},
		"亥": {"子时(23:00-01:00)", "辰时(07:00-09:00)", "午时(11:00-13:00)"},
	}
}

// buildActivityCatalog returns 20+ 宜 activities and 20+ 忌 activities.
func buildActivityCatalog() map[string][]string {
	return map[string][]string{
		"宜": {
			"嫁娶", "出行", "开市", "交易", "立契",
			"入宅", "安床", "修造", "上梁", "竖柱",
			"安门", "作灶", "祭祀", "祈福", "求嗣",
			"开光", "解除", "纳采", "订盟", "移徙",
			"会友", "入学", "裁衣", "纳畜", "栽种",
		},
		"忌": {
			"动土", "安葬", "行丧", "伐木", "开渠",
			"放水", "行舟", "开仓", "出货", "置产",
			"筑堤", "补垣", "塞穴", "破土", "啟攒",
			"除服", "成服", "移柩", "入殓", "开生坟",
			"合寿木", "谢土", "苫盖", "取鱼", "畋猎",
		},
	}
}
