-- 002_seed_auspicious.sql
-- Seed data for auspicious_rules table (黄历宜忌规则)
-- Each row defines which activities are auspicious/inauspicious for a given stem-branch day.

INSERT INTO auspicious_rules (category, day_stem, day_branch, content, created_at, updated_at) VALUES
('宜', '甲', '子', '["嫁娶","出行","祭祀","祈福","开市","交易"]', NOW(), NOW()),
('宜', '甲', '寅', '["入宅","安床","修造","上梁","立契"]', NOW(), NOW()),
('宜', '丙', '午', '["嫁娶","开光","求嗣","会友","裁衣"]', NOW(), NOW()),
('宜', '戊', '辰', '["祭祀","祈福","入学","纳采","订盟"]', NOW(), NOW()),
('宜', '庚', '申', '["出行","开市","交易","竖柱","安门"]', NOW(), NOW()),
('忌', '甲', '午', '["动土","安葬","行丧","伐木","开渠"]', NOW(), NOW()),
('忌', '乙', '酉', '["破土","啟攒","除服","成服","移柩"]', NOW(), NOW()),
('忌', '戊', '戌', '["放水","行舟","开仓","出货","置产"]', NOW(), NOW()),
('忌', '庚', '寅', '["筑堤","补垣","塞穴","入殓","开生坟"]', NOW(), NOW()),
('忌', '壬', '子', '["合寿木","谢土","苫盖","取鱼","畋猎"]', NOW(), NOW());
