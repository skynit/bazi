package buyi

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const Source = "倪海厦天纪六十四卦详解"

type Hexagram struct {
	Number       int
	Name         string
	HumanWay     string
	ImageReading string
}

type Reading struct {
	Hexagram Hexagram
	Score    int
	Level    string
	Summary  string
	Advice   string
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Draw() Reading {
	hexagram := Hexagrams[randomIndex(len(Hexagrams))]
	return BuildReading(hexagram)
}

func BuildReading(hexagram Hexagram) Reading {
	score := ScoreHexagram(hexagram)
	level := LevelForScore(score)
	return Reading{
		Hexagram: hexagram,
		Score:    score,
		Level:    level,
		Summary:  summaryFor(hexagram),
		Advice:   adviceFor(level),
	}
}

func ScoreHexagram(hexagram Hexagram) int {
	text := hexagram.Name + hexagram.HumanWay + hexagram.ImageReading
	score := 60

	if containsAny(text, []string{"大吉", "吉", "亨通", "泰平", "大有", "增益", "高升", "重生", "解脱", "诚信", "既济"}) {
		score += 18
	}
	if containsAny(text, []string{"守正", "谦", "渐", "节制", "稳", "顺时", "光明", "喜悦", "聚集", "升"}) {
		score += 8
	}
	if containsAny(text, []string{"争", "阻碍", "困", "险", "退避", "晦暗", "剥", "凶", "刑罚", "官司", "漂泊", "未完成"}) {
		score -= 12
	}
	if containsAny(text, []string{"危机", "小人", "穷途", "九死", "崩塌", "凶兆", "名不正", "盛极必衰"}) {
		score -= 20
	}

	if score < 35 {
		return 35
	}
	if score > 90 {
		return 90
	}
	return score
}

func LevelForScore(score int) string {
	switch {
	case score >= 82:
		return "大吉"
	case score >= 70:
		return "吉"
	case score >= 55:
		return "平"
	case score >= 40:
		return "谨慎"
	default:
		return "凶险"
	}
}

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func randomIndex(size int) int {
	if size <= 1 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(size)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

func summaryFor(hexagram Hexagram) string {
	return "今日得" + hexagram.Name + "，" + hexagram.HumanWay
}

func adviceFor(level string) string {
	switch level {
	case "大吉":
		return "今日气势较顺，可主动把握关键机会；仍以守正、谦和为底，不因顺境而躁进。"
	case "吉":
		return "今日宜顺势推进重要事项，先定规则再行动，借助贵人和团队力量会更稳。"
	case "平":
		return "今日以稳为主，适合整理计划、补足短板；不急于求成，按部就班即可。"
	case "谨慎":
		return "今日宜先守后动，减少争执和冒进；遇到阻滞时先停下来观察局势。"
	default:
		return "今日宜避险止损，少做高风险决定；把注意力放在修正错误、远离纷争。"
	}
}
