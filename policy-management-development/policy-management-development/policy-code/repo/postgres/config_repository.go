package repo

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"policy-management/core/domain"
)

// ConfigRepository handles reads from the policy_state_config table.
// Config rows are seeded by migrations 001 + 002 and rarely change at runtime.
// All values are strings; use domain.PolicyStateConfig.AsInt() / AsFloat() for typed access.
//
// Constraint C7: All SQL uses policy_mgmt. schema prefix.
// [§8.7, FR-PM-001 — config drives FLC period, routing timeouts, cooling durations, etc.]
type ConfigRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewConfigRepository constructs a ConfigRepository.
func NewConfigRepository(db *dblib.DB, cfg *config.Config) *ConfigRepository {
	return &ConfigRepository{db: db, cfg: cfg}
}

const configTable = "policy_mgmt.policy_state_config"

// ─────────────────────────────────────────────────────────────────────────────
// GetConfig — single key lookup
// ─────────────────────────────────────────────────────────────────────────────

// GetConfig retrieves the string value for a config key.
// Returns pgx.ErrNoRows if the key does not exist (treat as fatal misconfiguration).
// Key constants are defined in domain.ConfigKey* constants.
// [§8.7]
func (r *ConfigRepository) GetConfig(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("config_key", "config_value", "data_type", "description", "updated_at", "updated_by").
		From(configTable).
		Where(sq.Eq{"config_key": key})

	row, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.PolicyStateConfig])
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("config key %q not found — check migration 001/002: %w", key, err)
		}
		return "", fmt.Errorf("GetConfig key=%s: %w", key, err)
	}
	return row.ConfigValue, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetAllConfigs — full config map (used to warm workflow state cache)
// ─────────────────────────────────────────────────────────────────────────────

// GetAllConfigs retrieves all policy_state_config rows and returns them as a map.
// Used by workflow activities that need multiple config values in one DB round-trip.
// The map key is config_key; the map value is config_value (always a string).
// The activity caches the map in workflow state to avoid repeated DB calls. [§18]
// [§8.7]
func (r *ConfigRepository) GetAllConfigs(ctx context.Context) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("config_key", "config_value", "data_type", "description", "updated_at", "updated_by").
		From(configTable).
		OrderBy("config_key ASC")

	rows, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[domain.PolicyStateConfig])
	if err != nil {
		return nil, fmt.Errorf("GetAllConfigs: %w", err)
	}

	result := make(map[string]string, len(rows))
	for _, row := range rows {
		result[row.ConfigKey] = row.ConfigValue
	}
	return result, nil
}
