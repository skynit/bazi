ALTER TABLE birth_charts ADD COLUMN element_detail JSON AFTER day_pillar;
ALTER TABLE birth_charts ADD COLUMN body_strength JSON AFTER element_detail;
ALTER TABLE birth_charts ADD COLUMN ziwei_result JSON;
ALTER TABLE birth_charts ADD COLUMN ziwei_computed BOOLEAN DEFAULT FALSE;
