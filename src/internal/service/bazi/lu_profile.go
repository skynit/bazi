package bazi

func canonicalLuProfile() ([10]string, [10]string) {
	return [10]string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"},
		[10]string{"寅", "卯", "巳", "午", "巳", "午", "申", "酉", "亥", "子"}
}

func luBranchForStem(stem string) (string, bool) {
	switch stem {
	case "甲":
		return "寅", true
	case "乙":
		return "卯", true
	case "丙", "戊":
		return "巳", true
	case "丁", "己":
		return "午", true
	case "庚":
		return "申", true
	case "辛":
		return "酉", true
	case "壬":
		return "亥", true
	case "癸":
		return "子", true
	default:
		return "", false
	}
}
