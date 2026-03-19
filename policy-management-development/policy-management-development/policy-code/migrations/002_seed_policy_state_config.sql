 

SET search_path TO policy_mgmt;

-- Convert config_value to TEXT safely
ALTER TABLE policy_state_config
ALTER COLUMN config_value TYPE TEXT
USING config_value::TEXT;

-- =====================================================
-- SECTION 1: Remission Slab Configuration
-- =====================================================

INSERT INTO policy_state_config (config_key, config_value, description, data_type) VALUES
('remission_min_policy_life_months','6','Policy must be ≥6 months old to qualify for remission','INTEGER'),
('lapse_age_threshold_months','36','Policy age threshold separating VOID_LAPSE and INACTIVE_LAPSE','INTEGER'),
('remission_slab_6mo_days','30','Remission duration for policy life 6–12 months','INTEGER'),
('remission_slab_12mo_days','60','Remission duration for policy life 12–24 months','INTEGER'),
('remission_slab_24mo_days','90','Remission duration for policy life 24–36 months','INTEGER'),
('remission_slab_36mo_months','12','Remission duration for policy life ≥36 months','INTEGER')

ON CONFLICT (config_key) DO UPDATE
SET config_value = EXCLUDED.config_value,
description = EXCLUDED.description,
updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SECTION 2: Pay Recovery Protection
-- =====================================================

INSERT INTO policy_state_config (config_key, config_value, description, data_type) VALUES
('pay_recovery_protection_months','12','Pay recovery policies remain ACTIVE for 12 months','INTEGER')

ON CONFLICT (config_key) DO UPDATE
SET config_value = EXCLUDED.config_value,
description = EXCLUDED.description,
updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SECTION 3: Forced Surrender
-- =====================================================

INSERT INTO policy_state_config (config_key, config_value, description, data_type) VALUES
('forced_surrender_reminder_count','3','Number of reminders before forced surrender','INTEGER'),
('forced_surrender_loan_ratio_threshold_pct','100','Loan outstanding % triggering forced surrender','INTEGER')

ON CONFLICT (config_key) DO UPDATE
SET config_value = EXCLUDED.config_value,
description = EXCLUDED.description,
updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SECTION 4: WLA Maturity Age
-- =====================================================

INSERT INTO policy_state_config (config_key, config_value, description, data_type) VALUES
('wla_maturity_age_years','80','Whole Life Assurance maturity age','INTEGER')

ON CONFLICT (config_key) DO UPDATE
SET config_value = EXCLUDED.config_value,
description = EXCLUDED.description,
updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SECTION 5: Continue-As-New Threshold
-- =====================================================

INSERT INTO policy_state_config (config_key, config_value, description, data_type) VALUES
('continue_as_new_max_history_mb','50','Temporal workflow history size limit','INTEGER')

ON CONFLICT (config_key) DO UPDATE
SET config_value = EXCLUDED.config_value,
description = EXCLUDED.description,
updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SECTION 6: Product Catalog
-- =====================================================

INSERT INTO policy_state_config (config_key, config_value, description, data_type) VALUES
(
'product_catalog.PLI',
'[{"code":"WLA","name":"Suraksha"},{"code":"EA","name":"Santosh"},{"code":"CWLA","name":"Suvidha"},{"code":"AEA","name":"Sumangal"},{"code":"CP","name":"Children Policy"},{"code":"JEA","name":"Yugal Suraksha"}]',
'PLI product catalog',
'STRING'
),
(
'product_catalog.RPLI',
'[{"code":"RWLA","name":"Gram Suraksha"},{"code":"REA","name":"Gram Santosh"},{"code":"RCWLA","name":"Gram Suvidha"},{"code":"RAEA","name":"Gram Sumangal"},{"code":"RCP","name":"Rural Children Policy"},{"code":"GP","name":"Gram Priya"}]',
'RPLI product catalog',
'STRING'
)

ON CONFLICT (config_key) DO UPDATE
SET config_value = EXCLUDED.config_value,
description = EXCLUDED.description,
updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SECTION 7: Dashboard Refresh Interval
-- =====================================================

INSERT INTO policy_state_config (config_key, config_value, description, data_type) VALUES
('dashboard_mv_refresh_interval_minutes','15','Materialized view refresh interval','INTEGER')

ON CONFLICT (config_key) DO UPDATE
SET config_value = EXCLUDED.config_value,
description = EXCLUDED.description,
updated_at = CURRENT_TIMESTAMP;
 