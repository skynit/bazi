package bazi

const (
	LunarJavascriptSourceRepo   = "https://github.com/6tail/lunar-javascript"
	LunarJavascriptSourceCommit = "4c45a59f79b856125516f31aefa8295035c16afd"
	CNLunarSourceRepo           = "https://github.com/OPN48/cnlunar"
	CNLunarSourceCommit         = "1d7f868967cc533c9b577ed0c3ffb3cb67bb5352"
	Tyme4GoSourceRepo           = "https://github.com/6tail/tyme4go"
	Tyme4GoSourceCommit         = "4e841be2264123f1c8629a99e06f6c4eceed205e"
	YuanHaiZiPingPDFSHA256      = "57a130f26a4d45abd0f706405c7f9de00a8e90b6d4630676370f504ebbe2a0f5"
	SanMingTongHuiPDFSHA256     = "63eb2a85036ebbd360b815a58780edd242bda0fd1a5faaac7413e00c5f726d47"
)

func baziExternalSilverSources() []RuleSourceMeta {
	return []RuleSourceMeta{
		lunarJavascriptSilverSource(),
		{
			ID: "cnlunar", Repository: CNLunarSourceRepo, Commit: CNLunarSourceCommit,
			Files: map[string]string{
				"cnlunar/lunar.py":   "049ad17c4153208d4cbcdc358177eff30d46f6429f0caf3676ea7d4170a2cdd8",
				"cnlunar/config.py":  "aa804333238a07633057f614ae611798cc64da4621549940bb0d2025c2ab478a",
				"cnlunar/solar24.py": "a284dbe276637fb27d93f12c608c234b516cd4fbb1201fa2415df27a49886bf1",
				"cnlunar/tools.py":   "2f99d0d77a2365b23021c5adf56354ed4a233df7e1856244aff2ee592e091b6a",
			},
			License: "MIT", SourceTier: "silver_external", ValidationStatus: "cross_checked_not_gold",
		},
	}
}

func lunarJavascriptSilverSource() RuleSourceMeta {
	return RuleSourceMeta{
		ID: "lunar_javascript", Repository: LunarJavascriptSourceRepo, Commit: LunarJavascriptSourceCommit,
		Files: map[string]string{
			"lunar.js": "9750324bfe1aa63c146f8c72b1143df924466c11c8a5277d7d9225c541a18aaa",
		},
		License: "MIT", SourceTier: "silver_external", ValidationStatus: "cross_checked_not_gold",
	}
}

func shenShaClassicalSources() []RuleSourceMeta {
	return []RuleSourceMeta{
		{
			ID: "yuan_hai_zi_ping_local_pdf", Repository: "workspace://library", Commit: "not_applicable",
			Files: map[string]string{
				"library/渊海子平.pdf": YuanHaiZiPingPDFSHA256,
			},
			License: "not_recorded", SourceTier: "classical_text_local", ValidationStatus: "text_located_not_expert_gold",
		},
		{
			ID: "san_ming_tong_hui_local_pdf", Repository: "workspace://library", Commit: "not_applicable",
			Files: map[string]string{
				"library/三命通会.pdf": SanMingTongHuiPDFSHA256,
			},
			License: "not_recorded", SourceTier: "classical_text_local", ValidationStatus: "text_located_not_expert_gold",
		},
	}
}

func tyme4goTerrainSource() []RuleSourceMeta {
	return []RuleSourceMeta{
		{
			ID: "tyme4go_v1_4_2", Repository: Tyme4GoSourceRepo, Commit: Tyme4GoSourceCommit,
			Files: map[string]string{
				"tyme/HeavenStem.go": "bb8f43a430637ceb1d72e77a99ba83dc38b95de2532e81a4fc403e71fd51452f",
				"tyme/Terrain.go":    "e28efa64df57a1da9af3f9e2a3e59cf2468ca65b40913f7fbfa303247b836668",
			},
			License: "MIT", SourceTier: "silver_external", ValidationStatus: "cross_checked_not_gold",
		},
	}
}

func tyme4goDefaultChildLimitSource() []RuleSourceMeta {
	return []RuleSourceMeta{
		{
			ID: "tyme4go_v1_4_2_default_child_limit", Repository: Tyme4GoSourceRepo, Commit: Tyme4GoSourceCommit,
			Files: map[string]string{
				"tyme/DefaultChildLimitProvider.go":  "54f84ec021962e6214edc8461d0a4ae33e3c096e4ca916d681e665316338aa7e",
				"tyme/AbstractChildLimitProvider.go": "0a92bc1f552357b15abac913bd7fee4b67e53444c11bad7d6d708a1fba56373d",
				"tyme/ChildLimitInfo.go":             "09b6d0ffc74bbc09307dc9df3acf1f09382a0313c842b70c75e9584e26ddc97b",
			},
			License: "MIT", SourceTier: "silver_external", ValidationStatus: "cross_checked_not_gold",
		},
	}
}
