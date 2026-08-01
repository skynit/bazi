package buyi

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const Source = "《周易》六十四卦卦序与八卦象意（现代白话整理）"

type Hexagram struct {
	Number int
	Name   string
}

type Reading struct {
	Hexagram    Hexagram
	Summary     string
	Reflection  string
	ImagePrompt string
	Advice      string
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
	return Reading{
		Hexagram:    hexagram,
		Summary:     summaryFor(hexagram),
		Reflection:  reflectionFor(hexagram),
		ImagePrompt: imagePromptFor(hexagram),
		Advice:      adviceFor(hexagram),
	}
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
	upper, lower, ok := trigramsFor(hexagram)
	if !ok {
		return fmt.Sprintf("%s的核心主题是“%s”。", hexagram.Name, themeFor(hexagram))
	}
	if upper.Name == lower.Name {
		return fmt.Sprintf(
			"上卦与下卦皆为%s（%s）。两重%s相叠，内在考量与外部处境都强调%s。此卦可用来思考“%s”。",
			upper.Name,
			upper.Symbol,
			upper.Name,
			upper.Quality,
			themeFor(hexagram),
		)
	}
	return fmt.Sprintf(
		"上卦为%s（%s），下卦为%s（%s）。下卦提示事情的起点可先看%s；上卦提示外部发展需留意%s。此卦可用来思考“%s”。",
		upper.Name,
		upper.Symbol,
		lower.Name,
		lower.Symbol,
		lower.Quality,
		upper.Quality,
		themeFor(hexagram),
	)
}

func reflectionFor(hexagram Hexagram) string {
	if hexagram.Number == 57 {
		return "当前更需要强势推进，还是通过反复沟通逐步进入？把目标、底线和可调整部分分别写清，看看是否存在“只顺应、没有确认”的情况。"
	}
	return fmt.Sprintf(
		"围绕“%s”对照当前问题：哪些事实已经清楚，哪一项关键条件仍需要先验证？",
		themeFor(hexagram),
	)
}

func imagePromptFor(hexagram Hexagram) string {
	upper, lower, ok := trigramsFor(hexagram)
	if !ok {
		return "分开记录内在判断和外部反馈，观察两者一致与冲突的地方。"
	}
	if upper.Name == lower.Name {
		return fmt.Sprintf(
			"上下两卦都强调%s。观察你的内在判断与外部做法是否一致；同时留意这种特质用得过度时，是否正在变成%s。",
			upper.Quality,
			upper.Excess,
		)
	}
	return fmt.Sprintf(
		"先分开看两层：下卦的%s是内在条件，上卦的%s是外部处境。记录两者目前一致与冲突的地方。",
		lower.Quality,
		upper.Quality,
	)
}

func adviceFor(hexagram Hexagram) string {
	if hexagram.Number == 57 {
		return "选一个阻力最小的入口，完成一次清晰沟通：说明目标、确认对方反馈，并约定下一次检查时间。不要用一味迁就代替明确结论。"
	}
	return fmt.Sprintf(
		"结合“%s”，只选择一个今天可以验证的小步骤；先写明事实与预期，再根据实际反馈调整。",
		themeFor(hexagram),
	)
}
