-- 20260426140000_add_multi_evidence.up.sql
ALTER TABLE attendances ADD COLUMN IF NOT EXISTS evidence_urls JSONB DEFAULT '[]'::jsonb;
