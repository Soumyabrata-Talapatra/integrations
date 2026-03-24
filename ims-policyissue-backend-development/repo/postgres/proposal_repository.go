package postgres

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"policy-issue-service/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

// ProposalRepository handles proposal-related database operations
type ProposalRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewProposalRepository creates a new ProposalRepository instance
func NewProposalRepository(db *dblib.DB, cfg *config.Config) *ProposalRepository {
	return &ProposalRepository{db: db, cfg: cfg}
}

// CreateProposalWithIndexing creates a new proposal with indexing data and status history atomically using pgx.Batch
// [FR-POL-007] New Business Indexing
// [BR-POL-015] State Machine
// CRITICAL: Uses pgx.Batch for atomic multi-table insert (single round-trip)
func (r *ProposalRepository) CreateProposalWithIndexing(ctx context.Context, proposal *domain.Proposal,
	indexing *domain.ProposalIndexing) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()

	// Step 1: Generate proposal number
	seqSQL := "SELECT nextval('policy_issue.policy_issue_seq') as seq"

	type seqResult struct {
		Seq int64 `db:"seq"`
	}

	result, err := dblib.ExecReturn(ctx, r.db, seqSQL, []any{}, pgx.RowToStructByName[seqResult])
	if err != nil {
		return fmt.Errorf("failed to generate sequence: %w", err)
	}

	proposalNumber := fmt.Sprintf("%s-%d-%08d",
		string(proposal.PolicyType),
		now.Year(),
		result.Seq,
	)

	// Step 2: First batch — insert proposal only
	batch1 := &pgx.Batch{}

	proposalSQL := `
		INSERT INTO policy_issue.proposals (
			proposal_number, insurant_name,quote_ref_number,  spouse_customer_id,
			product_code, policy_type, channel, status, entry_path, current_stage,
			sum_assured, policy_term, premium_payment_frequency,
			created_by, created_at, updated_at, version, base_premium, gst_amount,total_premium
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
			$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
		)
		RETURNING proposal_id
	`

	batch1.Queue(proposalSQL,
		proposalNumber,
		proposal.InsurantName,
		proposal.QuoteRefNumber,
		// proposal.CustomerID,
		proposal.SpouseCustomerID,
		proposal.ProductCode,
		proposal.PolicyType,
		proposal.Channel,
		proposal.Status,
		proposal.EntryPath,
		"INDEXING",
		proposal.SumAssured,
		proposal.PolicyTerm,
		proposal.PremiumPaymentFrequency,
		proposal.CreatedBy,
		now,
		now,
		1,
		proposal.BasePremium,
		proposal.GSTAmount,
		proposal.TotalPremium,
	)

	br1 := r.db.SendBatch(ctx, batch1)

	var proposalID int64
	if err := br1.QueryRow().Scan(&proposalID); err != nil {
		br1.Close()
		return fmt.Errorf("failed to insert proposal: %w", err)
	}

	if err := br1.Close(); err != nil {
		return err
	}

	// Step 3: Second batch — insert indexing + history
	batch2 := &pgx.Batch{}

	indexingSQL := `
		INSERT INTO policy_issue.proposal_indexing (
			proposal_id, po_code, issue_circle, issue_ho, issue_post_office, declaration_date, receipt_date,
			indexing_date, proposal_date, indexed_by, indexed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`

	batch2.Queue(indexingSQL,
		proposalID,
		indexing.POCode,
		indexing.IssueCircle,
		indexing.IssueHO,
		indexing.IssuePostOffice,
		indexing.DeclarationDate,
		indexing.ReceiptDate,
		indexing.IndexingDate,
		indexing.ProposalDate,
		proposal.CreatedBy,
		now,
	)

	historySQL := `
		INSERT INTO policy_issue.proposal_status_history (
			proposal_id, to_status, changed_by, changed_at, comments, version
		) VALUES ($1,$2,$3,$4,$5,$6)
	`

	batch2.Queue(historySQL,
		proposalID,
		proposal.Status,
		proposal.CreatedBy,
		now,
		"Proposal created via indexing",
		1,
	)

	br2 := r.db.SendBatch(ctx, batch2)

	// Consume indexing result
	if _, err := br2.Exec(); err != nil {
		br2.Close()
		return fmt.Errorf("failed to insert indexing: %w", err)
	}

	// Consume history result
	if _, err := br2.Exec(); err != nil {
		br2.Close()
		return fmt.Errorf("failed to insert history: %w", err)
	}

	if err := br2.Close(); err != nil {
		return err
	}

	// Hydrate struct
	proposal.ProposalID = proposalID
	proposal.ProposalNumber = proposalNumber
	proposal.CreatedAt = now
	proposal.UpdatedAt = now

	return nil
}

// func (r *ProposalRepository) CreateProposalWithIndexing(ctx context.Context, proposal *domain.Proposal, indexing *domain.ProposalIndexing) error {
// 	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
// 	defer cancel()

// 	now := time.Now()

// 	// Step 1: Generate proposal number using sequence (required for batch)
// 	seqSQL := "SELECT nextval('policy_issue_seq') as seq"
// 	type seqResult struct {
// 		Seq int64 `db:"seq"`
// 	}
// 	result, err := dblib.ExecReturn(ctx, r.db, seqSQL, []any{}, pgx.RowToStructByName[seqResult])
// 	if err != nil {
// 		return fmt.Errorf("failed to generate sequence: %w", err)
// 	}

// 	proposalNumber := fmt.Sprintf("%s-%d-%08d", string(proposal.PolicyType), now.Year(), result.Seq)

// 	// Step 2: Use pgx.Batch for atomic multi-table insert
// 	// This ensures all inserts succeed or fail together (single round-trip)
// 	batch := &pgx.Batch{}

// 	// Queue 1: Insert into proposals and get the generated ID
// 	proposalSQL := `
// 		INSERT INTO policy_issue.proposals (
// 			proposal_number, quote_ref_number, customer_id, spouse_customer_id,
// 			product_code, policy_type, channel, status, entry_path, current_stage,
// 			sum_assured, policy_term, premium_payment_frequency,
// 			created_by, created_at, updated_at, version
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
// 		RETURNING proposal_id
// 	`
// 	batch.Queue(proposalSQL,
// 		proposalNumber, proposal.QuoteRefNumber, proposal.CustomerID, proposal.SpouseCustomerID,
// 		proposal.ProductCode, proposal.PolicyType, proposal.Channel, proposal.Status,
// 		proposal.EntryPath, "INDEXING",
// 		proposal.SumAssured, proposal.PolicyTerm, proposal.PremiumPaymentFrequency,
// 		proposal.CreatedBy, now, now, 1)

// 	// Queue 2: Insert into proposal_indexing
// 	// Note: We use currval to get the proposal_id from the first insert within the same batch
// 	indexingSQL := `
// 		INSERT INTO policy_issue.proposal_indexing (
// 			proposal_id, po_code, declaration_date, receipt_date, indexing_date, proposal_date,
// 			indexed_by, indexed_at
// 		) VALUES (currval('policy_issue_seq'), $1, $2, $3, $4, $5, $6, $7)
// 	`
// 	batch.Queue(indexingSQL,
// 		indexing.POCode, indexing.DeclarationDate, indexing.ReceiptDate,
// 		indexing.IndexingDate, indexing.ProposalDate, proposal.CreatedBy, now)

// 	// Queue 3: Insert into proposal_status_history
// 	historySQL := `
// 		INSERT INTO policy_issue.proposal_status_history (
// 			proposal_id, to_status, changed_by, changed_at, comments, version
// 		) VALUES (currval('policy_issue_seq'), $1, $2, $3, $4, $5)
// 	`
// 	batch.Queue(historySQL,
// 		proposal.Status, proposal.CreatedBy, now, "Proposal created via indexing", 1)

// 	// Execute batch atomically
// 	br := r.db.SendBatch(ctx, batch)
// 	defer br.Close()

// 	// Get the proposal_id from the first query result
// 	var proposalID int64
// 	if err := br.QueryRow().Scan(&proposalID); err != nil {
// 		return fmt.Errorf("failed to insert proposal in batch: %w", err)
// 	}

// 	// Consume remaining batch results
// 	if _, err := br.Exec(); err != nil {
// 		return fmt.Errorf("failed to execute batch inserts: %w", err)
// 	}

// 	// Hydrate the proposal with generated fields
// 	proposal.ProposalID = proposalID
// 	proposal.ProposalNumber = proposalNumber
// 	proposal.CreatedAt = now
// 	proposal.UpdatedAt = now

// 	return nil
// }

// CreateProposalWithAadhaar creates a proposal with all Aadhaar-verified details in a single batch
func (r *ProposalRepository) CreateProposalWithAadhaar(ctx context.Context, proposal *domain.Proposal, insured *domain.ProposalInsured) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()

	// Step 1: Generate proposal number
	seqSQL := "SELECT nextval('policy_issue_seq') as seq"
	type seqResult struct {
		Seq int64 `db:"seq"`
	}
	result, err := dblib.ExecReturn(ctx, r.db, seqSQL, []any{}, pgx.RowToStructByName[seqResult])
	if err != nil {
		return fmt.Errorf("failed to generate sequence: %w", err)
	}

	proposalNumber := fmt.Sprintf("%s-%d-%08d", string(proposal.PolicyType), now.Year(), result.Seq)

	// Step 2: Atomic multi-table insert
	batch := &pgx.Batch{}

	// 1. Proposals
	proposalSQL := `
		INSERT INTO proposals (
			proposal_number, customer_id, product_code, policy_type, channel, 
			status, entry_path, current_stage, sum_assured, policy_term, 
			premium_payment_frequency, created_by, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13, 1)
		RETURNING proposal_id
	`
	batch.Queue(proposalSQL,
		proposalNumber, proposal.CustomerID, proposal.ProductCode, proposal.PolicyType, proposal.Channel,
		proposal.Status, proposal.EntryPath, mapStatusToStage(proposal.Status),
		proposal.SumAssured, proposal.PolicyTerm, proposal.PremiumPaymentFrequency,
		proposal.CreatedBy, now)

	// 2. Proposal Indexing
	indexingSQL := `
		INSERT INTO proposal_indexing (
			proposal_id, po_code, declaration_date, receipt_date, indexing_date, proposal_date,
			indexed_by, indexed_at, created_at, updated_at
		) VALUES (currval('policy_issue_seq'), 'WEB', $1, $1, $1, $1, $2, $1, $1, $1)
	`
	batch.Queue(indexingSQL, now, proposal.CreatedBy)

	// 3. Proposal Insured
	insuredSQL := `
		INSERT INTO proposal_insured (
			proposal_id, salutation, first_name, middle_name, last_name, gender, 
			date_of_birth, address_line1, city, state, pin_code, mobile, email,
			created_at, updated_at
		) VALUES (currval('policy_issue_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
	`
	batch.Queue(insuredSQL,
		insured.Salutation, insured.FirstName, insured.MiddleName, insured.LastName, insured.Gender,
		insured.DateOfBirth, insured.AddressLine1, insured.City, insured.State, insured.PinCode,
		insured.Mobile, insured.Email, now)

	// 4. Proposal Data Entry (Auto-complete some sections)
	dataEntrySQL := `
		INSERT INTO proposal_data_entry (
			proposal_id, insured_details_complete, data_entry_status, 
			data_entry_by, data_entry_started_at, created_at, updated_at
		) VALUES (currval('policy_issue_seq'), TRUE, 'IN_PROGRESS', $1, $2, $2, $2)
	`
	batch.Queue(dataEntrySQL, proposal.CreatedBy, now)

	// 5. Status History
	historySQL := `
		INSERT INTO proposal_status_history (
			proposal_id, to_status, changed_by, changed_at, comments, version
		) VALUES (currval('policy_issue_seq'), $1, $2, $3, 'Aadhaar-based proposal created', 1)
	`
	batch.Queue(historySQL, proposal.Status, proposal.CreatedBy, now)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	var proposalID int64
	if err := br.QueryRow().Scan(&proposalID); err != nil {
		return fmt.Errorf("failed to insert Aadhaar proposal: %w", err)
	}

	for i := 1; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to execute Aadhaar batch statement %d: %w", i, err)
		}
	}

	proposal.ProposalID = proposalID
	proposal.ProposalNumber = proposalNumber
	proposal.CreatedAt = now
	proposal.UpdatedAt = now

	return nil
}

// CreateProposal creates a new proposal (legacy method, use CreateProposalWithIndexing instead)
// [FR-POL-007] New Business Indexing
// [BR-POL-015] State Machine
func (r *ProposalRepository) CreateProposal(ctx context.Context, proposal *domain.Proposal) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Single round-trip using CTE: Generate proposal number and insert in one query
	sql := `
		WITH generated_number AS (
			SELECT nextval('policy_issue_seq') as seq
		),
		inserted AS (
			INSERT INTO proposals (
				proposal_number, quote_ref_number, customer_id, spouse_customer_id,
				product_code, policy_type, channel, status, created_by
			)
			SELECT 
				$1 || '-' || EXTRACT(YEAR FROM CURRENT_DATE) || '-' || LPAD(generated_number.seq::text, 8, '0'),
				$2, $3, $4, $5, $6, $7, $8, $9
			FROM generated_number
			RETURNING proposal_id, proposal_number, created_at, updated_at
		)
		SELECT * FROM inserted
	`

	results, err := dblib.ExecReturns(ctx, r.db, sql, []any{
		string(proposal.PolicyType), proposal.QuoteRefNumber, proposal.CustomerID, proposal.SpouseCustomerID,
		proposal.ProductCode, proposal.PolicyType, proposal.Channel, proposal.Status, proposal.CreatedBy,
	}, pgx.RowToStructByName[domain.Proposal])
	if err != nil {
		return fmt.Errorf("failed to create proposal: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no proposal returned after insert")
	}

	// Hydrate the proposal with generated fields
	result := results[0]
	proposal.ProposalID = result.ProposalID
	proposal.ProposalNumber = result.ProposalNumber
	proposal.CreatedAt = result.CreatedAt
	proposal.UpdatedAt = result.UpdatedAt

	return nil
}

// GetProposalByID retrieves a proposal by ID
func (r *ProposalRepository) GetProposalByID(ctx context.Context, proposalID int64) (*domain.Proposal, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("proposal_id",
		"proposal_number",
		"quote_ref_number",
		"customer_id",
		"spouse_customer_id",
		"product_code",
		"policy_type",
		"sum_assured",
		"policy_term",
		"status",
		"premium_payment_frequency",
		"workflow_id").
		From("proposals").Where(sq.Eq{"proposal_id": proposalID})

	proposal, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Proposal])
	if err != nil {
		return nil, err
	}

	return &proposal, nil
}

// GetProposalByNumber retrieves a proposal by proposal number
func (r *ProposalRepository) GetProposalByNumber(ctx context.Context, proposalNumber string) (*domain.ProposalOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(
		"proposal_id",
		"proposal_number",
		"insurant_name",
		"quote_ref_number",
		"customer_id",
		"spouse_customer_id",
		"proposer_customer_id",
		"is_proposer_same_as_insured",
		"premium_payer_type",
		"payer_customer_id",
		"product_code",
		"policy_type",
		"sum_assured",
		"policy_term",
		"premium_ceasing_age",
		"premium_payment_frequency",
		"entry_path",
		"channel",
		"status",
		"current_stage",
		"workflow_id",
		"created_by",
		"created_at",
		"updated_at",
		"deleted_at",
		"version",
		"base_premium",
		"total_premium",
		"gst_amount",
	).From("proposals").Where(sq.Eq{"proposal_number": proposalNumber})

	proposal, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.ProposalOutput])
	if err != nil {
		return nil, err
	}

	return &proposal, nil
}

// UpdateProposalStatus updates proposal status with transition validation and audit logging
// [BR-POL-015] State Transition Rules
func (r *ProposalRepository) UpdateProposalStatus(ctx context.Context, proposalID int64, newStatus domain.ProposalStatus, comments string, changedBy int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Get current proposal to validate transition
	proposal, err := r.GetProposalByID(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal for status update: %w", err)
	}

	// Validate state transition using domain logic
	currentStatus := proposal.Status
	if !domain.IsValidStatusTransition(currentStatus, domain.ProposalStatus(newStatus)) {
		return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
	}

	now := time.Now()

	// Map status to current_stage
	currentStage := mapStatusToStage(newStatus)
	newVersion := proposal.Version + 1

	// Update proposal status, current_stage, and version
	updateSQL := `
		UPDATE proposals 
		SET status = $1, current_stage = $2, version = $3, updated_at = $4, updated_by = $5
		WHERE proposal_id = $6
	`
	_, err = dblib.Exec(ctx, r.db, updateSQL, []any{newStatus, currentStage, newVersion, now, changedBy, proposalID})
	if err != nil {
		return fmt.Errorf("failed to update proposal status: %w", err)
	}

	// Insert status history record
	historySQL := `
		INSERT INTO proposal_status_history (
			proposal_id, from_status, to_status, changed_by, changed_at, comments, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = dblib.Exec(ctx, r.db, historySQL, []any{
		proposalID, string(currentStatus), newStatus, changedBy, now, comments, proposal.Version + 1})
	if err != nil {
		return fmt.Errorf("failed to insert status history: %w", err)
	}

	return nil
}

// GenerateProposalNumber generates a unique proposal number
// Format: {POLICY_TYPE}-{YEAR}-{SEQUENCE:08d}
func (r *ProposalRepository) GenerateProposalNumber(ctx context.Context, policyType string, stateCode string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use raw SQL for sequence generation
	sql := "SELECT nextval('policy_issue_seq') as seq"

	type seqResult struct {
		Seq int64 `db:"seq"`
	}

	result, err := dblib.ExecReturn(ctx, r.db, sql, []any{}, pgx.RowToStructByName[seqResult])
	if err != nil {
		return "", err
	}

	refNumber := fmt.Sprintf("%s-%d-%08d", policyType, time.Now().Year(), result.Seq)
	return refNumber, nil
}

// UpdateSectionComplete updates the section completion status in data entry
func (r *ProposalRepository) UpdateSectionComplete(ctx context.Context, proposalID int64, section string, complete bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Build update query based on section
	var column string
	switch section {
	case "insured":
		column = "insured_details_complete"
	case "nominees":
		column = "nominee_details_complete"
	case "policy":
		column = "policy_details_complete"
	case "agent":
		column = "agent_details_complete"
	case "medical":
		column = "medical_details_complete"
	case "documents":
		column = "documents_complete"
	case "declaration":
		column = "declaration_complete"
	case "proposer":
		column = "proposer_details_complete"
	case "premium":
		// For premium, update the proposal_indexing table instead
		return r.updatePremiumSection(ctx, proposalID, complete)
	default:
		return fmt.Errorf("unknown section: %s", section)
	}

	// Check if data entry row exists
	var exists bool
	checkQuery := dblib.Psql.Select("1").From("proposal_data_entry").Where(sq.Eq{"proposal_id": proposalID})
	_, found, err := dblib.SelectOneOK(ctx, r.db, checkQuery, pgx.RowTo[int])
	if err == nil && found {
		exists = true
	}

	if exists {
		// Update existing row
		query := dblib.Psql.Update("proposal_data_entry").
			SetMap(map[string]interface{}{
				column:       complete,
				"updated_at": time.Now(),
			}).
			Where(sq.Eq{"proposal_id": proposalID})
		_, err = dblib.Update(ctx, r.db, query)
		return err
	}

	// Insert new row with section marked complete
	columns := map[string]interface{}{
		"proposal_id":       proposalID,
		column:              complete,
		"data_entry_status": "IN_PROGRESS",
		"created_at":        time.Now(),
		"updated_at":        time.Now(),
	}
	query := dblib.Psql.Insert("proposal_data_entry").SetMap(columns)
	_, err = dblib.Insert(ctx, r.db, query)
	return err
}

// updatePremiumSection updates the first premium status in proposal_indexing
func (r *ProposalRepository) updatePremiumSection(ctx context.Context, proposalID int64, complete bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update("proposal_indexing").
		SetMap(map[string]interface{}{
			"first_premium_paid": complete,
			"updated_at":         time.Now(),
		}).
		Where(sq.Eq{"proposal_id": proposalID})

	_, err := dblib.Update(ctx, r.db, query)
	return err
}

// UpdateProposalFields updates specific proposal fields
func (r *ProposalRepository) UpdateProposalFields(ctx context.Context, proposalID int64, fields map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Add updated_at timestamp
	fields["updated_at"] = time.Now()

	query := dblib.Psql.Update("proposals").
		SetMap(fields).
		Where(sq.Eq{"proposal_id": proposalID})

	_, err := dblib.Update(ctx, r.db, query)
	return err
}

// UpdateDataEntryFields updates specific proposal_data_entry fields
func (r *ProposalRepository) UpdateDataEntryFields(ctx context.Context, proposalID int64, fields map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Add updated_at timestamp
	fields["updated_at"] = time.Now()

	query := dblib.Psql.Update("proposal_data_entry").
		SetMap(fields).
		Where(sq.Eq{"proposal_id": proposalID})

	_, err := dblib.Update(ctx, r.db, query)
	return err
}

type HUFMemberRepoInput struct {
	IsFinancedHUF                 bool
	KartaName                     *string
	HUFPan                        *string
	LifeAssuredDifferentFromKarta bool
	KartaDifferentReason          *string
	MemberName                    string
	MemberRelationship            string
	MemberAge                     int
}

func (r *ProposalRepository) InsertHUFMembers(
	ctx context.Context,
	proposalID int64,
	members []HUFMemberRepoInput,
) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	for _, m := range members {

		query := dblib.Psql.Insert("proposal_huf_member").
			SetMap(map[string]interface{}{
				"proposal_id":                       proposalID,
				"is_financed_huf":                   m.IsFinancedHUF,
				"karta_name":                        m.KartaName,
				"huf_pan":                           m.HUFPan,
				"life_assured_different_from_karta": m.LifeAssuredDifferentFromKarta,
				"karta_different_reason":            m.KartaDifferentReason,
				"member_name":                       m.MemberName,
				"member_relationship":               m.MemberRelationship,
				"member_age":                        m.MemberAge,
				"created_at":                        time.Now(),
				"updated_at":                        time.Now(),
			})

		if _, err := dblib.Insert(ctx, r.db, query); err != nil {
			return err
		}
	}

	return nil
}

type MWPATrusteeRepoInput struct {
	TrustType    string
	TrusteeName  string
	TrusteeDOB   *string
	Relationship *string
	Address      *string
}

// InsertMWPATrustee inserts a trustee record for MWPAs
func (r *ProposalRepository) InsertMWPATrustee(ctx context.Context, proposalID int64,
	trustee MWPATrusteeRepoInput) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert("proposal_mwpa_trustee").
		SetMap(map[string]interface{}{
			"proposal_id":  proposalID,
			"trust_type":   trustee.TrustType,
			"trustee_name": trustee.TrusteeName,
			"trustee_dob":  trustee.TrusteeDOB,
			"relationship": trustee.Relationship,
			"address":      trustee.Address,
			"created_at":   time.Now(),
			"updated_at":   time.Now(),
		})

	_, err := dblib.Insert(ctx, r.db, query)
	return err
}

// RecordFirstPremium records the first premium payment in proposal_indexing
func (r *ProposalRepository) RecordFirstPremium(ctx context.Context, proposalID int64, amount float64,
	paymentMode, paymentReference string, paymentDate time.Time, collectedBy int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Update proposal_indexing with first premium details
	query := dblib.Psql.Update("proposal_indexing").
		SetMap(map[string]interface{}{
			"first_premium_paid":      true,
			"first_premium_date":      paymentDate,
			"first_premium_reference": paymentReference,
			"premium_payment_method":  paymentMode,
			"initial_premium":         amount,
			"updated_at":              time.Now(),
		}).
		Where(sq.Eq{"proposal_id": proposalID})

	_, err := dblib.Update(ctx, r.db, query)
	return err
}

// GeneratePolicyNumber generates a unique policy number using the policy_number_sequence table
// Uses format_pattern from database: {prefix}-{year}-{value:06d}
func (r *ProposalRepository) GeneratePolicyNumber(ctx context.Context, policyType domain.PolicyType, seriesPrefix string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use the policy_number_sequence table to get next value with series_prefix
	// This ensures uniqueness per product_type + series_prefix combination
	updateSQL := `
		UPDATE policy_number_sequence 
		SET next_value = next_value + 1 
		WHERE product_type = $1 AND series_prefix = $2
		RETURNING series_prefix, next_value - 1 as seq, format_pattern
	`

	type seqResult struct {
		SeriesPrefix  string `db:"series_prefix"`
		Seq           int64  `db:"seq"`
		FormatPattern string `db:"format_pattern"`
	}

	// Default series prefix if not provided
	if seriesPrefix == "" {
		seriesPrefix = "STD"
	}

	result, err := dblib.ExecReturn(ctx, r.db, updateSQL, []any{policyType, seriesPrefix}, pgx.RowToStructByName[seqResult])
	if err != nil {
		return "", fmt.Errorf("failed to get policy number sequence: %w", err)
	}

	// Use format_pattern from database or default
	formatPattern := result.FormatPattern
	if formatPattern == "" {
		formatPattern = "{prefix}-{year}-{value:06d}"
	}

	// Replace placeholders in format pattern
	policyNumber := strings.ReplaceAll(formatPattern, "{prefix}", result.SeriesPrefix)
	policyNumber = strings.ReplaceAll(policyNumber, "{year}", fmt.Sprintf("%d", time.Now().Year()))
	policyNumber = strings.ReplaceAll(policyNumber, "{value:06d}", fmt.Sprintf("%06d", result.Seq))
	policyNumber = strings.ReplaceAll(policyNumber, "{value}", fmt.Sprintf("%d", result.Seq))

	return policyNumber, nil
}

// GetProposalsByStatus retrieves proposals filtered by status with pagination
func (r *ProposalRepository) GetProposalsByStatus(ctx context.Context, status string, skip, limit int) ([]domain.Proposal, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Build base query
	baseQuery := dblib.Psql.Select("proposal_id",
		"proposal_number",
		"status",
		"product_code",
		"sum_assured",
		"created_at").From("proposals")
	countQuery := dblib.Psql.Select("COUNT(*)").From("proposals")

	// Apply status filter if provided
	if status != "" {
		baseQuery = baseQuery.Where(sq.Eq{"status": status})
		countQuery = countQuery.Where(sq.Eq{"status": status})
	}

	// Get total count
	total, err := dblib.SelectOne(ctx, r.db, countQuery, pgx.RowTo[int64])
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	// offset := (page - 1) * limit
	offset := skip
	baseQuery = baseQuery.OrderBy("created_at DESC").Limit(uint64(limit)).Offset(uint64(offset))

	// Get proposals
	proposals, err := dblib.SelectRows(ctx, r.db, baseQuery, pgx.RowToStructByNameLax[domain.Proposal])
	if err != nil {
		return nil, 0, err
	}

	return proposals, total, nil
}

// CheckAllSectionsComplete checks if all required data entry sections are complete
func (r *ProposalRepository) CheckAllSectionsComplete(ctx context.Context, proposalID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := `
		SELECT 
			COALESCE(insured_details_complete, false) as insured,
			COALESCE(nominee_details_complete, false) as nominees,
			COALESCE(policy_details_complete, false) as policy,
			COALESCE(agent_details_complete, false) as agent,
			COALESCE(medical_details_complete, false) as medical,
			COALESCE(documents_complete, false) as documents,
			COALESCE(declaration_complete, false) as declaration,
			COALESCE(proposer_details_complete, false) as proposer
		FROM proposal_data_entry 
		WHERE proposal_id = $1
	`

	type sectionsResult struct {
		Insured     bool `db:"insured"`
		Nominees    bool `db:"nominees"`
		Policy      bool `db:"policy"`
		Agent       bool `db:"agent"`
		Medical     bool `db:"medical"`
		Documents   bool `db:"documents"`
		Declaration bool `db:"declaration"`
		Proposer    bool `db:"proposer"`
	}

	result, err := dblib.ExecReturn(ctx, r.db, query, []any{proposalID}, pgx.RowToStructByName[sectionsResult])
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return result.Insured && result.Nominees && result.Policy && result.Agent && result.Medical && result.Documents && result.Declaration && result.Proposer, nil
}

// UpdateWorkflowID updates the temporal workflow ID for a proposal
func (r *ProposalRepository) UpdateWorkflowID(ctx context.Context, proposalID int64, workflowID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateSQL := `
		UPDATE policy_issue.proposals 
		SET workflow_id = $1, updated_at = $2
		WHERE proposal_id = $3
	`
	_, err := dblib.Exec(ctx, r.db, updateSQL, []any{workflowID, time.Now(), proposalID})
	return err
}

// RecordQCReview records the QC review decision
func (r *ProposalRepository) RecordQCReview(
	ctx context.Context,
	proposalID int64,
	decision string,
	comments string,
	reviewerID string,
) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	reviewerIDInt, err := strconv.ParseInt(reviewerID, 10, 64)
	if err != nil {
		return err
	}

	sql := `
	INSERT INTO policy_issue.proposal_qc_review
	(proposal_id, qr_decision, qr_comments, qr_decision_by, qr_decision_at)
	VALUES ($1,$2,$3,$4,$5)
	`

	_, err = dblib.Exec(ctx, r.db, sql,
		[]any{proposalID, decision, comments, reviewerIDInt, now})

	return err
}

// func (r *ProposalRepository) RecordQCReview(ctx context.Context, proposalID int64, decision string, comments string, reviewerID string) error {
// 	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
// 	defer cancel()

// 	now := time.Now()
// 	reviewerIDInt, _ := strconv.ParseInt(reviewerID, 10, 64)

// 	upsertSQL := `
// 		INSERT INTO policy_issue.proposal_qc_review (proposal_id, qr_decision, qr_comments, qr_decision_by, qr_decision_at)
// 		VALUES ($1, $2, $3, $4, $5)
// 		ON CONFLICT (proposal_id) DO UPDATE SET
// 			qr_decision = EXCLUDED.qr_decision,
// 			qr_comments = EXCLUDED.qr_comments,
// 			qr_decision_by = EXCLUDED.qr_decision_by,
// 			qr_decision_at = EXCLUDED.qr_decision_at
// 	`
// 	_, err := dblib.Exec(ctx, r.db, upsertSQL, []any{proposalID, decision, comments, reviewerIDInt, now})
// 	return err
// }

// RecordApproval records the approver decision
func (r *ProposalRepository) RecordApproval(ctx context.Context, proposalID int64, decision string, comments string, approverID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	approverIDInt, _ := strconv.ParseInt(approverID, 10, 64)

	upsertSQL := `
		INSERT INTO proposal_approval (proposal_id, approver_decision, approver_comments, approver_decision_by, approver_decision_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (proposal_id) DO UPDATE SET
			approver_decision = EXCLUDED.approver_decision,
			approver_comments = EXCLUDED.approver_comments,
			approver_decision_by = EXCLUDED.approver_decision_by,
			approver_decision_at = EXCLUDED.approver_decision_at
			
	`
	_, err := dblib.Exec(ctx, r.db, upsertSQL, []any{proposalID, decision, comments, approverIDInt, now})
	return err
}

// ApprovalRoutingConfig represents a row from the approval_routing_config table
// [BR-POL-016] Approval routing by SA bands
type ApprovalRoutingConfig struct {
	ConfigID      int64   `db:"config_id"`
	SAMin         float64 `db:"sa_min"`
	SAMax         float64 `db:"sa_max"`
	ApproverLevel int     `db:"approver_level"`
	ApproverRole  string  `db:"approver_role"`
}

// GetApprovalRoutingConfig returns the approval routing config for a given sum assured
// [BR-POL-016] Approval Routing by Sum Assured
func (r *ProposalRepository) GetApprovalRoutingConfig(ctx context.Context, sumAssured float64) (*ApprovalRoutingConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("config_id", "sa_min", "sa_max", "approver_level", "approver_role").
		From("approval_routing_config").
		Where(sq.LtOrEq{"sa_min": sumAssured}).
		Where(sq.GtOrEq{"sa_max": sumAssured}).
		Where(sq.Eq{"is_active": true}).
		Limit(1)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[ApprovalRoutingConfig])
	if err != nil {
		return nil, fmt.Errorf("failed to get approval routing config for SA %.2f: %w", sumAssured, err)
	}

	return &result, nil
}

// IssuanceData represents key fields from proposal_issuance for FLC validation
type IssuanceData struct {
	IssuanceID         int64      `db:"issuance_id"`
	ProposalID         int64      `db:"proposal_id"`
	PolicyNumber       *string    `db:"policy_number"`
	PolicyIssueDate    *time.Time `db:"policy_issue_date"`
	DispatchDate       *time.Time `db:"dispatch_date"`
	DeliveryDate       *time.Time `db:"delivery_date"`
	FLCStartDate       *time.Time `db:"flc_start_date"`
	FLCEndDate         *time.Time `db:"flc_end_date"`
	FLCStatus          *string    `db:"flc_status"`
	FLCConfigID        *int64     `db:"flc_config_id"`
	FLCCancelRefundAmt *float64   `db:"flc_cancel_refund_amount"`
}

// GetIssuanceByProposalID retrieves issuance data for a proposal
func (r *ProposalRepository) GetIssuanceByProposalID(ctx context.Context, proposalID int64) (*IssuanceData, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(
		"issuance_id", "proposal_id", "policy_number",
		"policy_issue_date", "dispatch_date", "delivery_date",
		"flc_start_date", "flc_end_date", "flc_status", "flc_config_id",
		"flc_cancel_refund_amount",
	).From("proposal_issuance").Where(sq.Eq{"proposal_id": proposalID})

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[IssuanceData])
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// FLCConfig represents a row from the free_look_config table
type FLCConfig struct {
	ConfigID      int64  `db:"config_id"`
	Channel       string `db:"channel"`
	PeriodDays    int    `db:"period_days"`
	StartDateRule string `db:"start_date_rule"`
}

// GetFLCConfig retrieves the free look config for a given channel and optional product type
// [BR-POL-021] Free Look Period Duration
// [BR-POL-028] FLC Start Date Determination
func (r *ProposalRepository) GetFLCConfig(ctx context.Context, channel string, productType *string) (*FLCConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Try product-specific config first, fall back to channel-only
	query := dblib.Psql.Select("config_id", "channel", "period_days", "start_date_rule").
		From("free_look_config").
		Where(sq.Eq{"channel": channel, "is_active": true}).
		OrderBy("product_type NULLS LAST"). // product-specific first
		Limit(1)

	if productType != nil {
		query = dblib.Psql.Select("config_id", "channel", "period_days", "start_date_rule").
			From("free_look_config").
			Where(sq.Eq{"channel": channel, "is_active": true}).
			Where(sq.Or{sq.Eq{"product_type": *productType}, sq.Eq{"product_type": nil}}).
			OrderBy("product_type NULLS LAST").
			Limit(1)
	}

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[FLCConfig])
	if err != nil {
		return nil, fmt.Errorf("failed to get FLC config for channel %s: %w", channel, err)
	}
	return &result, nil
}

// UpdateFLCCancellation updates the proposal_issuance with FLC refund details
func (r *ProposalRepository) UpdateFLCCancellation(ctx context.Context, proposalID int64, reason string, refundAmount float64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	updateSQL := `
		UPDATE proposal_issuance
		SET flc_status = 'CANCELLED',
			flc_cancel_requested_at = $2,
			flc_cancel_reason = $3,
			flc_cancel_refund_amount = $4,
			flc_cancel_processed_at = $5,
			updated_at = $5
		WHERE proposal_id = $1
	`
	_, err := dblib.Exec(ctx, r.db, updateSQL, []any{proposalID, now, reason, refundAmount, now})
	return err
}

// GetFirstPremiumAmount retrieves the initial premium amount for a proposal from proposal_indexing
func (r *ProposalRepository) GetFirstPremiumAmount(ctx context.Context, proposalID int64) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COALESCE(initial_premium, 0)").
		From("proposal_indexing").
		Where(sq.Eq{"proposal_id": proposalID})

	amount, err := dblib.SelectOne(ctx, r.db, query, pgx.RowTo[float64])
	if err != nil {
		return 0, err
	}
	return amount, nil
}

// CheckPremiumPaid checks if first premium has been paid for a proposal
func (r *ProposalRepository) CheckPremiumPaid(ctx context.Context, proposalID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COALESCE(first_premium_paid, false)").
		From("proposal_indexing").
		Where(sq.Eq{"proposal_id": proposalID})

	paid, err := dblib.SelectOne(ctx, r.db, query, pgx.RowTo[bool])
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return paid, nil
}

// RecordFLCRequest records a Free Look Cancellation request
func (r *ProposalRepository) RecordFLCRequest(ctx context.Context, proposalID int64, reason string, comments string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	updateSQL := `
		UPDATE proposal_issuance 
		SET flc_status = 'CANCELLED', flc_cancel_requested_at = $2, flc_cancel_reason = $3, updated_at = $4
		WHERE proposal_id = $1
	`
	_, err := dblib.Exec(ctx, r.db, updateSQL, []any{proposalID, now, reason, now})
	return err
}

// mapStatusToStage maps proposal status to current_stage
func mapStatusToStage(status domain.ProposalStatus) string {
	switch status {
	case domain.ProposalStatusDraft, domain.ProposalStatusIndexed:
		return "INDEXING"
	case domain.ProposalStatusDataEntry:
		return "DATA_ENTRY"
	case domain.ProposalStatusQCPending, domain.ProposalStatusQCApproved, domain.ProposalStatusQCRejected, domain.ProposalStatusQCReturned:
		return "QC_REVIEW"
	case domain.ProposalStatusPendingMedical, domain.ProposalStatusMedicalApproved, domain.ProposalStatusMedicalRejected:
		return "MEDICAL"
	case domain.ProposalStatusApprovalPending, domain.ProposalStatusApproved, domain.ProposalStatusRejected:
		return "APPROVAL"
	case domain.ProposalStatusIssued, domain.ProposalStatusDispatched, domain.ProposalStatusFreeLookActive, domain.ProposalStatusActive:
		return "ISSUANCE"
	default:
		return "INDEXING"
	}
}

// SaveInsuredDetails persists insured details to proposal_insured and proposal_data_entry tables
func (r *ProposalRepository) SaveInsuredDetails(ctx context.Context, proposalID int64, customerID int64, insured *domain.ProposalInsured, dataEntryBy int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()

	// Use batch to update both tables
	batch := &pgx.Batch{}
	// 1. Update customer_id in proposals table
	updateProposalSQL := `
		UPDATE proposals
		SET customer_id = $1,
			updated_at = $2
		WHERE proposal_id = $3
		  AND deleted_at IS NULL
	`
	batch.Queue(updateProposalSQL, customerID, now, proposalID)
	// 2. Upsert into proposal_insured
	insuredSQL := `
		INSERT INTO proposal_insured (
			proposal_id, salutation, first_name, middle_name, last_name, gender, 
			date_of_birth, marital_status, occupation, annual_income,
			address_line1, address_line2, address_line3, city, state, pin_code,
			mobile, email, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (proposal_id) DO UPDATE SET
			salutation = EXCLUDED.salutation,
			first_name = EXCLUDED.first_name,
			middle_name = EXCLUDED.middle_name,
			last_name = EXCLUDED.last_name,
			gender = EXCLUDED.gender,
			date_of_birth = EXCLUDED.date_of_birth,
			marital_status = EXCLUDED.marital_status,
			occupation = EXCLUDED.occupation,
			annual_income = EXCLUDED.annual_income,
			address_line1 = EXCLUDED.address_line1,
			address_line2 = EXCLUDED.address_line2,
			address_line3 = EXCLUDED.address_line3,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			pin_code = EXCLUDED.pin_code,
			mobile = EXCLUDED.mobile,
			email = EXCLUDED.email,
			updated_at = EXCLUDED.updated_at
	`
	batch.Queue(insuredSQL,
		proposalID, insured.Salutation, insured.FirstName, insured.MiddleName, insured.LastName, insured.Gender,
		insured.DateOfBirth, insured.MaritalStatus, insured.Occupation, insured.AnnualIncome,
		insured.AddressLine1, insured.AddressLine2, insured.AddressLine3, insured.City, insured.State, insured.PinCode,
		insured.Mobile, insured.Email, now, now,
	)

	// 3. Update proposal_data_entry metadata
	dataEntrySQL := `
		INSERT INTO proposal_data_entry (
			proposal_id, data_entry_by, data_entry_started_at, 
			data_entry_status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (proposal_id) DO UPDATE SET
			data_entry_started_at = EXCLUDED.data_entry_started_at,
			data_entry_by = EXCLUDED.data_entry_by,
			updated_at = EXCLUDED.updated_at
	`
	batch.Queue(dataEntrySQL, proposalID, dataEntryBy, now, "IN_PROGRESS", now, now)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to execute batch statement %d: %w", i, err)
		}
	}

	return nil
}

// SaveProposerDetails persists proposer details to proposal_proposer table and updates proposals table
func (r *ProposalRepository) SaveProposerDetails(ctx context.Context, proposalID int64,
	proposer *domain.ProposalProposer, dataEntryBy int64, isSameAsInsured bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()
	batch := &pgx.Batch{}

	// 1. Upsert into proposal_proposer
	proposerSQL := `
		INSERT INTO proposal_proposer (
			proposal_id, customer_id, relationship_to_insured, relationship_details,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (proposal_id) DO UPDATE SET
			customer_id = EXCLUDED.customer_id,
			relationship_to_insured = EXCLUDED.relationship_to_insured,
			relationship_details = EXCLUDED.relationship_details,
			updated_at = EXCLUDED.updated_at
	`
	batch.Queue(proposerSQL,
		proposalID, proposer.CustomerID, proposer.RelationshipToInsured, proposer.RelationshipDetails,
		now, now)

	// 2. Update proposals table to set is_proposer_same_as_insured = FALSE
	var proposerCustomerID *int64

	if !isSameAsInsured {
		proposerCustomerID = &proposer.CustomerID
	}
	updateProposalSQL := `
		UPDATE proposals 
		SET is_proposer_same_as_insured = $1,
			proposer_customer_id = $2,
			updated_at = $3
		WHERE proposal_id = $4
	`
	batch.Queue(updateProposalSQL, isSameAsInsured, proposerCustomerID, now, proposalID)

	// 3. Update data_entry_by in proposal_data_entry if not already set
	updateDataEntrySQL := `
		UPDATE proposal_data_entry 
		SET data_entry_by = $1,
			updated_at = $2
		WHERE proposal_id = $3 AND data_entry_by IS NULL
	`
	batch.Queue(updateDataEntrySQL, dataEntryBy, now, proposalID)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to execute batch statement %d: %w", i, err)
		}
	}

	return nil
}

func (r *ProposalRepository) SetProposerSameAsInsured(
	ctx context.Context,
	proposalID int64,
) error {

	query := `
	UPDATE policy_issue.proposals
	SET is_proposer_same_as_insured = TRUE,
	    proposer_customer_id = NULL,
	    updated_at = NOW()
	WHERE proposal_id = $1
	`

	_, err := r.db.Exec(ctx, query, proposalID)
	return err
}

// GetProposerByProposalID retrieves proposer details by proposal ID
func (r *ProposalRepository) GetProposerByProposalID(ctx context.Context, proposalID int64) (*domain.ProposalProposer, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(
		"proposer_id", "proposal_id", "customer_id",
		"relationship_to_insured",
		"relationship_details",
	).
		From("proposal_proposer").Where(sq.Eq{"proposal_id": proposalID})

	proposer, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.ProposalProposer])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No proposer record means proposer is same as insured
		}
		return nil, err
	}

	return &proposer, nil
}

// GetInsuredByProposalID retrieves insured details by proposal ID
func (r *ProposalRepository) GetInsuredByProposalID(ctx context.Context, proposalID int64) (*domain.ProposalInsuredOuptput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").From("proposal_insured").Where(sq.Eq{"proposal_id": proposalID})

	insured, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.ProposalInsuredOuptput])
	if err != nil {
		return nil, err
	}

	return &insured, nil
}

// GetDataEntryByProposalID retrieves data entry details by proposal ID
func (r *ProposalRepository) GetDataEntryByProposalID(ctx context.Context, proposalID int64) (*domain.ProposalDataEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Select(
			"data_entry_id",
			"proposal_id",
			"policy_taken_under",
			"aadhaar_photo_document_id",
			"age_proof_type",
			"subsequent_payment_mode",
			"data_entry_status",
			"insured_details_complete",
			"nominee_details_complete",
			"policy_details_complete",
			"agent_details_complete",
			"medical_details_complete",
			"declaration_complete",
			"proposer_details_complete",
			"documents_complete",
			"data_entry_by",
			"created_at",
			"updated_at",
		).From("proposal_data_entry").Where(sq.Eq{"proposal_id": proposalID})

	dataEntry, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.ProposalDataEntry])
	if err != nil {
		return nil, err
	}

	return &dataEntry, nil
}

// SaveNominees persists multiple nominee details to proposal_nominee table atomically using pgx.Batch

func (r *ProposalRepository) SaveNominees(ctx context.Context, proposalID int64,
	nominees []*domain.ProposalNominee) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()
	batch := &pgx.Batch{}

	for _, nominee := range nominees {

		query := `
		INSERT INTO policy_issue.proposal_nominee (
			proposal_id,
			salutation,
			first_name,
			middle_name,
			last_name,
			gender,
			date_of_birth,
			is_minor,
			relationship,
			share_percentage,
			appointee_name,
			appointee_relationship,
			created_at,
			updated_at,
			nominee_customer_id
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
		)
		ON CONFLICT (proposal_id, nominee_customer_id)
		DO UPDATE SET
			salutation = EXCLUDED.salutation,
			first_name = EXCLUDED.first_name,
			middle_name = EXCLUDED.middle_name,
			last_name = EXCLUDED.last_name,
			gender = EXCLUDED.gender,
			date_of_birth = EXCLUDED.date_of_birth,
			is_minor = EXCLUDED.is_minor,
			relationship = EXCLUDED.relationship,
			share_percentage = EXCLUDED.share_percentage,
			appointee_name = EXCLUDED.appointee_name,
			appointee_relationship = EXCLUDED.appointee_relationship,
			updated_at = EXCLUDED.updated_at,
			nominee_customer_id = EXCLUDED.nominee_customer_id
		`

		batch.Queue(
			query,
			proposalID,
			nominee.Salutation,
			nominee.FirstName,
			nominee.MiddleName,
			nominee.LastName,
			nominee.Gender,
			nominee.DateOfBirth,
			nominee.IsMinor,
			nominee.Relationship,
			nominee.SharePercentage,
			nominee.AppointeeName,
			nominee.AppointeeRelationship,
			now,
			now,
			nominee.NomineeCustomerID,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed nominee batch statement %d: %w", i, err)
		}
	}

	return nil
}

// func (r *ProposalRepository) SaveNominees(ctx context.Context, proposalID int64, nominees []*domain.ProposalNominee) error {
// 	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
// 	defer cancel()

// 	now := time.Now()
// 	batch := &pgx.Batch{}

// 	// First, clear existing nominees for this proposal to handle updates correctly
// 	batch.Queue("DELETE FROM proposal_nominee WHERE proposal_id = $1", proposalID)

// 	// Queue insertions for each nominee
// 	for _, nominee := range nominees {
// 		query := `
// 			INSERT INTO proposal_nominee (
// 				proposal_id, salutation, first_name, middle_name, last_name,
// 				gender, date_of_birth, is_minor, relationship, share_percentage,
// 				appointee_name, appointee_relationship, created_at, updated_at
// 			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
// 		`
// 		batch.Queue(query,
// 			proposalID, nominee.Salutation, nominee.FirstName, nominee.MiddleName, nominee.LastName,
// 			nominee.Gender, nominee.DateOfBirth, nominee.IsMinor, nominee.Relationship, nominee.SharePercentage,
// 			nominee.AppointeeName, nominee.AppointeeRelationship, now, now)
// 	}

// 	br := r.db.SendBatch(ctx, batch)
// 	defer br.Close()

// 	// Execute all statements in batch
// 	for i := 0; i < batch.Len(); i++ {
// 		if _, err := br.Exec(); err != nil {
// 			return fmt.Errorf("failed to execute batch nominee statement %d: %w", i, err)
// 		}
// 	}

// 	return nil
// }

// SaveAgentDetails persists agent details to proposal_agent table
// func (r *ProposalRepository) SaveAgentDetails(ctx context.Context, proposalID int64, agentCode string, agentType string) error {
// 	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
// 	defer cancel()

// 	now := time.Now()

//		upsertSQL := `
//			INSERT INTO proposal_agent (proposal_id, agent_code, agent_type, created_at, updated_at)
//			VALUES ($1, $2, $3, $4, $4)
//			ON CONFLICT (proposal_id) DO UPDATE SET
//				agent_code = EXCLUDED.agent_code,
//				agent_type = EXCLUDED.agent_type,
//				updated_at = EXCLUDED.updated_at
//		`
//		_, err := dblib.Exec(ctx, r.db, upsertSQL, []any{proposalID, agentCode, agentType, now})
//		return err
//	}
func (r *ProposalRepository) SaveAgentDetails(ctx context.Context, proposalID int64,
	agentID string, agentSalutation string, agentName string, agentMobile string,
	agentEmail string, agentLandline string, agentStdCode string, receivesCorrespondence bool,
	opportunityID string) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	upsertSQL := `
		INSERT INTO policy_issue.proposal_agent (proposal_id, agent_id,agent_salutation,
			agent_name,agent_mobile,agent_email,agent_landline,agent_std_code,
			receives_correspondence,opportunity_id,created_at,updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
		)
		ON CONFLICT (proposal_id) DO UPDATE SET
			agent_id = EXCLUDED.agent_id,
			agent_salutation = EXCLUDED.agent_salutation,
			agent_name = EXCLUDED.agent_name,
			agent_mobile = EXCLUDED.agent_mobile,
			agent_email = EXCLUDED.agent_email,
			agent_landline = EXCLUDED.agent_landline,
			agent_std_code = EXCLUDED.agent_std_code,
			receives_correspondence = EXCLUDED.receives_correspondence,
			opportunity_id = EXCLUDED.opportunity_id,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dblib.Exec(ctx, r.db, upsertSQL, []any{proposalID, agentID, agentSalutation,
		agentName, agentMobile, agentEmail, agentLandline, agentStdCode,
		receivesCorrespondence, opportunityID, now, now,
	})

	return err
}

// SaveMedicalInfo persists medical questionnaire to proposal_medical_info table
func (r *ProposalRepository) SaveMedicalInfo(ctx context.Context, proposalID int64, m *domain.ProposalMedicalInfo) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	upsertSQL := `
	INSERT INTO policy_issue.proposal_medical_info (
		proposal_id, insured_index, is_sound_health,
		disease_tb, disease_cancer, disease_paralysis, disease_insanity,
		disease_heart_lungs, disease_kidney, disease_brain, disease_hiv,
		disease_hepatitis_b, disease_epilepsy, disease_nervous,
		disease_liver, disease_leprosy, disease_physical_deformity,
		disease_other, disease_details,
		family_hereditary, family_hereditary_details,
		medical_leave_3yr, leave_kind, leave_period, leave_ailment,
		hospital_name, hospitalization_from, hospitalization_to,
		physical_deformity, deformity_type, family_doctor_name,
		created_at, updated_at
	) VALUES (
	 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,
		$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33
	)
	ON CONFLICT (proposal_id, insured_index)
	DO UPDATE SET
		is_sound_health = EXCLUDED.is_sound_health,
		disease_tb = EXCLUDED.disease_tb,
		disease_cancer = EXCLUDED.disease_cancer,
		disease_paralysis = EXCLUDED.disease_paralysis,
		disease_insanity = EXCLUDED.disease_insanity,
		disease_heart_lungs = EXCLUDED.disease_heart_lungs,
		disease_kidney = EXCLUDED.disease_kidney,
		disease_brain = EXCLUDED.disease_brain,
		disease_hiv = EXCLUDED.disease_hiv,
		disease_hepatitis_b = EXCLUDED.disease_hepatitis_b,
		disease_epilepsy = EXCLUDED.disease_epilepsy,
		disease_nervous = EXCLUDED.disease_nervous,
		disease_liver = EXCLUDED.disease_liver,
		disease_leprosy = EXCLUDED.disease_leprosy,
		disease_physical_deformity = EXCLUDED.disease_physical_deformity,
		disease_other = EXCLUDED.disease_other,
		disease_details = EXCLUDED.disease_details,
		family_hereditary = EXCLUDED.family_hereditary,
		family_hereditary_details = EXCLUDED.family_hereditary_details,
		medical_leave_3yr = EXCLUDED.medical_leave_3yr,
		leave_kind = EXCLUDED.leave_kind,
		leave_period = EXCLUDED.leave_period,
		leave_ailment = EXCLUDED.leave_ailment,
		hospital_name = EXCLUDED.hospital_name,
		hospitalization_from = EXCLUDED.hospitalization_from,
		hospitalization_to = EXCLUDED.hospitalization_to,
		physical_deformity = EXCLUDED.physical_deformity,
		deformity_type = EXCLUDED.deformity_type,
		family_doctor_name = EXCLUDED.family_doctor_name,
		updated_at = EXCLUDED.updated_at
	`

	_, err := dblib.Exec(ctx, r.db, upsertSQL, []any{proposalID, m.InsuredIndex,
		m.IsSoundHealth, m.DiseaseTB, m.DiseaseCancer, m.DiseaseParalysis,
		m.DiseaseInsanity, m.DiseaseHeartLungs, m.DiseaseKidney, m.DiseaseBrain,
		m.DiseaseHIV, m.DiseaseHepatitisB, m.DiseaseEpilepsy, m.DiseaseNervous,
		m.DiseaseLiver, m.DiseaseLeprosy, m.DiseasePhysicalDeformity, m.DiseaseOther,
		m.DiseaseDetails, m.FamilyHereditary, m.FamilyHereditaryDetails, m.MedicalLeave3yr,
		m.LeaveKind, m.LeavePeriod, m.LeaveAilment, m.HospitalName, m.HospitalizationFrom, m.HospitalizationTo,
		m.PhysicalDeformity, m.DeformityType, m.FamilyDoctorName, now, now,
	})

	return err
}

// func (r *ProposalRepository) SaveMedicalInfo(ctx context.Context, proposalID int64, medicalInfo *domain.ProposalMedicalInfo) error {
// 	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
// 	defer cancel()

// 	now := time.Now()

// 	upsertSQL := `
// 		INSERT INTO proposal_medical_info (
// 			proposal_id, insured_index, is_sound_health, disease_tb, disease_cancer,
// 			disease_paralysis, disease_insanity, disease_heart_lungs, disease_kidney,
// 			disease_brain, disease_hiv, disease_hepatitis_b, disease_epilepsy,
// 			disease_nervous, disease_liver, disease_leprosy, other_diseases,
// 			disease_details, created_at, updated_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $19)
// 		ON CONFLICT (proposal_id, insured_index) DO UPDATE SET
// 			is_sound_health = EXCLUDED.is_sound_health,
// 			disease_tb = EXCLUDED.disease_tb,
// 			disease_cancer = EXCLUDED.disease_cancer,
// 			disease_paralysis = EXCLUDED.disease_paralysis,
// 			disease_insanity = EXCLUDED.disease_insanity,
// 			disease_heart_lungs = EXCLUDED.disease_heart_lungs,
// 			disease_kidney = EXCLUDED.disease_kidney,
// 			disease_brain = EXCLUDED.disease_brain,
// 			disease_hiv = EXCLUDED.disease_hiv,
// 			disease_hepatitis_b = EXCLUDED.disease_hepatitis_b,
// 			disease_epilepsy = EXCLUDED.disease_epilepsy,
// 			disease_nervous = EXCLUDED.disease_nervous,
// 			disease_liver = EXCLUDED.disease_liver,
// 			disease_leprosy = EXCLUDED.disease_leprosy,
// 			other_diseases = EXCLUDED.other_diseases,
// 			disease_details = EXCLUDED.disease_details,
// 			updated_at = EXCLUDED.updated_at
// 	`
// 	_, err := dblib.Exec(ctx, r.db, upsertSQL, []any{
// 		proposalID, medicalInfo.InsuredIndex, medicalInfo.IsSoundHealth,
// 		medicalInfo.DiseaseTB, medicalInfo.DiseaseCancer, medicalInfo.DiseaseParalysis,
// 		medicalInfo.DiseaseInsanity, medicalInfo.DiseaseHeartLungs, medicalInfo.DiseaseKidney,
// 		medicalInfo.DiseaseBrain, medicalInfo.DiseaseHIV, medicalInfo.DiseaseHepatitisB,
// 		medicalInfo.DiseaseEpilepsy, medicalInfo.DiseaseNervous, medicalInfo.DiseaseLiver,
// 		medicalInfo.DiseaseLeprosy, medicalInfo.OtherDiseases, medicalInfo.DiseaseDetails,
// 		now})
// 	return err
// }

// GetCustomerAggregateSA retrieves the total sum assured for a customer across active proposals/policies
// [VAL-POL-008] Aggregate SA check
// [INT-POL-002] Customer Service integration (future: delegate to Customer Service)
func (r *ProposalRepository) GetCustomerAggregateSA(ctx context.Context, customerID int64, policyType string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Select("COALESCE(SUM(sum_assured), 0)").
		From("proposals").
		Where(sq.Eq{"customer_id": customerID}).
		Where(sq.NotEq{"status": []string{"REJECTED", "CANCELLED_DEATH", "FLC_CANCELLED"}}).
		Where(sq.Eq{"deleted_at": nil})

	if policyType != "" {
		query = query.Where(sq.Eq{"policy_type": policyType})
	}

	type aggregateResult struct {
		Total float64 `db:"coalesce"`
	}

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[aggregateResult])
	if err != nil {
		return 0, err
	}

	return result.Total, nil
}

// DuplicateProposalInfo contains info about a potential duplicate proposal
type DuplicateProposalInfo struct {
	ProposalID     int64  `db:"proposal_id"`
	ProposalNumber string `db:"proposal_number"`
	Status         string `db:"status"`
	ProductCode    string `db:"product_code"`
}

// CheckDuplicateProposal checks for active proposals with the same customer + product combination
// [BR-POL-024] Deduplication check before proposal creation
// Returns nil if no duplicate found, otherwise returns info about the existing proposal
func (r *ProposalRepository) CheckDuplicateProposal(ctx context.Context, productCode string) (*DuplicateProposalInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Check for active proposals (not rejected, not cancelled, not issued)
	// Same customer + same product + still in progress = duplicate
	activeStatuses := []string{
		"DRAFT", "INDEXED", "DATA_ENTRY", "QC_PENDING", "QC_APPROVED", "QC_RETURNED",
		"PENDING_MEDICAL", "MEDICAL_APPROVED", "APPROVAL_PENDING", "APPROVED",
	}

	query := dblib.Psql.Select("proposal_id", "proposal_number", "status", "product_code").
		From("proposals").
		// Where(sq.Eq{"customer_id": customerID}).
		Where(sq.Eq{"product_code": productCode}).
		Where(sq.Eq{"status": activeStatuses}).
		Where(sq.Eq{"deleted_at": nil}).
		Limit(1)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[DuplicateProposalInfo])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No duplicate found
		}
		return nil, err
	}
	return &result, nil
}

// RecordDataEntryAssignment records CPC user assignment when starting data entry
func (r *ProposalRepository) RecordDataEntryAssignment(ctx context.Context, proposalID int64, assignedTo int64, comments string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()

	// Check if data entry record already exists
	checkQuery := dblib.Psql.Select("COUNT(*)").
		From("proposal_data_entry").
		Where(sq.Eq{"proposal_id": proposalID})

	type countResult struct {
		Count int `db:"count"`
	}

	var result countResult
	result, err := dblib.SelectOne(ctx, r.db, checkQuery, pgx.RowToStructByName[countResult])
	if err != nil {
		return err
	}

	if result.Count == 0 {
		// Insert new data entry record
		insertSQL, insertArgs, insertErr := dblib.Psql.Insert("proposal_data_entry").
			Columns("proposal_id", "assigned_to", "assignment_comments", "created_at", "updated_at").
			Values(proposalID, assignedTo, comments, now, now).
			ToSql()
		if insertErr != nil {
			return insertErr
		}

		_, execErr := dblib.Exec(ctx, r.db, insertSQL, insertArgs)
		return execErr
	}

	// Update existing data entry record
	updateSQL, updateArgs, updateErr := dblib.Psql.Update("proposal_data_entry").
		Set("assigned_to", assignedTo).
		Set("assignment_comments", comments).
		Set("updated_at", now).
		Where(sq.Eq{"proposal_id": proposalID}).
		ToSql()
	if updateErr != nil {
		return updateErr
	}

	_, updateExecErr := dblib.Exec(ctx, r.db, updateSQL, updateArgs)
	return updateExecErr
}

// CheckDuplicateQuoteConversion checks if a quote has already been converted to a proposal
func (r *ProposalRepository) CheckDuplicateQuoteConversion(ctx context.Context, quoteRefNumber string) (*DuplicateProposalInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("proposal_id", "proposal_number", "status", "product_code").
		From("proposals").
		Where(sq.Eq{"quote_ref_number": quoteRefNumber}).
		Where(sq.Eq{"deleted_at": nil}).
		Limit(1)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[DuplicateProposalInfo])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No existing conversion
		}
		return nil, err
	}
	return &result, nil
}
func (r *ProposalRepository) UpdateDataEntryStatus(ctx context.Context, proposalID int64, status string,
	changedBy int64) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateSQL := `
		UPDATE proposal_data_entry
		SET data_entry_status = $1,
		    data_entry_completed_at = $2,
		    data_entry_by = $3
		WHERE proposal_id = $4
	`

	_, err := dblib.Exec(ctx, r.db, updateSQL, []any{status, time.Now(), changedBy, proposalID})
	if err != nil {
		return fmt.Errorf("failed to update proposal_data_entry status: %w", err)
	}

	return nil
}

func (r *ProposalRepository) GetProposalIndexingByProposalID(ctx context.Context, proposalID int64) (*domain.ProposalIndexing, error) {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Select("proposal_id", "proposal_date").
		From("proposal_indexing").
		Where(sq.Eq{"proposal_id": proposalID})

	indexing, err := dblib.SelectOne(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.ProposalIndexing],
	)
	if err != nil {
		return nil, err
	}

	return &indexing, nil
}

func (r *ProposalRepository) InsertProposalIssuance(ctx context.Context, proposalID int64, policyNumber string,
	issueDate time.Time, commencementDate time.Time, maturityDate time.Time) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Insert("policy_issue.proposal_issuance").
		Columns(
			"proposal_id",
			"policy_number",
			"policy_issue_date",
			"acceptance_date",
			"policy_commencement_date",
			"maturity_date",
		).
		Values(
			proposalID,
			policyNumber,
			issueDate,
			issueDate, // acceptance_date same as issueDate
			commencementDate,
			maturityDate,
		).
		Suffix("ON CONFLICT (proposal_id) DO NOTHING")

	_, err := dblib.Insert(ctx, r.db, query)
	if err != nil {
		return fmt.Errorf("failed to insert proposal issuance: %w", err)
	}

	return nil
}

func (r *ProposalRepository) UpdateBondDetails(ctx context.Context, proposalID int64,
	bondDocumentID string, bondGeneratedBy int64,
) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Update("policy_issue.proposal_issuance").
		Set("bond_generated", true).
		Set("bond_document_id", bondDocumentID).
		Set("bond_generated_at", time.Now().UTC()).
		Set("bond_generated_by", bondGeneratedBy).
		Where(sq.Eq{"proposal_id": proposalID})

	cmdTag, err := dblib.Update(ctx, r.db, query)
	if err != nil {
		return fmt.Errorf("failed to update bond details: %w", err)
	}

	// Optional safety check
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *ProposalRepository) GetIndexingSection(ctx context.Context, proposalID int64,
) (*domain.ProposalIndexingSection, error) {

	query := dblib.Psql.
		Select(
			"declaration_date",
			"receipt_date",
			"indexing_date",
			"proposal_date",
			"po_code",
			"issue_circle",
			"issue_ho",
			"issue_post_office",
		).
		From("policy_issue.proposal_indexing").
		Where(sq.Eq{
			"proposal_id": proposalID,
		})

	data, err := dblib.SelectOne(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.ProposalIndexingSection],
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *ProposalRepository) GetFirstPremiumSection(ctx context.Context, proposalID int64,
) (*domain.ProposalFirstPremium, error) {

	query := dblib.Psql.
		Select(
			"first_premium_paid",
			"first_premium_date",
			"first_premium_reference",
			"first_premium_receipt_number",
			"premium_payment_method",
			"initial_premium",
			"short_excess_premium",
		).
		From("policy_issue.proposal_indexing").
		Where(sq.Eq{
			"proposal_id": proposalID,
		})

	data, err := dblib.SelectOne(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.ProposalFirstPremium],
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *ProposalRepository) GetAgentByProposalID(
	ctx context.Context,
	proposalID int64,
) (*domain.ProposalAgentOutput, error) {

	query := dblib.Psql.
		Select(
			"agent_id",
			"agent_salutation",
			"agent_name",
			"agent_mobile",
			"agent_email",
			"agent_landline",
			"agent_std_code",
			"receives_correspondence",
			"opportunity_id",
		).
		From("policy_issue.proposal_agent").
		Where(sq.Eq{
			"proposal_id": proposalID,
		})

	data, err := dblib.SelectOne(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.ProposalAgentOutput],
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *ProposalRepository) GetMedicalInfoByProposalID(
	ctx context.Context,
	proposalID int64,
) ([]domain.ProposalMedicalInfo, error) {

	query := dblib.Psql.
		Select(
			"insured_index",
			"is_sound_health",
			"disease_tb",
			"disease_cancer",
			"disease_paralysis",
			"disease_insanity",
			"disease_heart_lungs",
			"disease_kidney",
			"disease_brain",
			"disease_hiv",
			"disease_hepatitis_b",
			"disease_epilepsy",
			"disease_nervous",
			"disease_liver",
			"disease_leprosy",
			"disease_physical_deformity",
			"disease_other",
			"disease_details",
			"family_hereditary",
			"family_hereditary_details",
		).
		From("policy_issue.proposal_medical_info").
		Where(sq.Eq{
			"proposal_id": proposalID,
		})

	data, err := dblib.SelectRows(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.ProposalMedicalInfo],
	)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *ProposalRepository) GetQCReviewByProposalID(
	ctx context.Context,
	proposalID int64,
) (*domain.ProposalQCReview, error) {

	query := dblib.Psql.
		Select(
			"qc_review_id",
			"proposal_id",
			"qr_decision",
			"qr_comments",
			"return_count",
			"last_return_reason",
		).
		From("policy_issue.proposal_qc_review").
		Where(sq.Eq{
			"proposal_id": proposalID,
		})

	data, err := dblib.SelectOne(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.ProposalQCReview],
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *ProposalRepository) GetApprovalByProposalID(
	ctx context.Context,
	proposalID int64,
) (*domain.ProposalApproval, error) {

	query := dblib.Psql.
		Select(
			"approval_id",
			"proposal_id",
			"approval_level",
			"approver_role",
			"assigned_approver_id",
			"approver_decision",
			"approver_comments",
			"approver_rejection_reason",
			"approver_decision_by",
			"approver_decision_at",
			"approval_due_date",
			"approval_reminder_sent",
		).
		From("policy_issue.proposal_approval").
		Where(sq.Eq{"proposal_id": proposalID})

	// 2. Debug: If you still see zeroes, try RowToStructByPositional
	// to see if the 'Lax' name-based mapping is the culprit.
	data, err := dblib.SelectOne(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.ProposalApproval],
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// This is what returns the "zeroed" object you are seeing.
			// If you get here, the DB literally has no row for that ID.
			return &domain.ProposalApproval{ProposalID: proposalID}, nil
		}
		return nil, err
	}

	return &data, nil
}

func (r *ProposalRepository) GetNomineesByProposalID(ctx context.Context, proposalID int64,
) ([]domain.ProposalNominee, error) {

	query := dblib.Psql.
		Select(
			"nominee_id",
			"salutation",
			"first_name",
			"middle_name",
			"last_name",
			"gender",
			"date_of_birth::text AS date_of_birth",
			"is_minor",
			"relationship",
			"share_percentage",
			"appointee_name",
			"appointee_relationship",
		).
		From("policy_issue.proposal_nominee").
		Where(sq.Eq{
			"proposal_id": proposalID,
		})

	data, err := dblib.SelectRows(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.ProposalNominee],
	)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreateAuditLog creates a single audit log entry
func (r *ProposalRepository) CreateAuditLog(ctx context.Context, auditLog *domain.ProposalAuditLog) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	sql := `
		INSERT INTO policy_issue.proposal_audit_log (
			proposal_id, entity_type, entity_id, field_name, old_value, new_value,
			change_type, changed_by, changed_at, change_reason, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING audit_id
	`

	results, err := dblib.ExecReturns(ctx, r.db, sql, []any{
		auditLog.ProposalID,
		auditLog.EntityType,
		auditLog.EntityID,
		auditLog.FieldName,
		auditLog.OldValue,
		auditLog.NewValue,
		auditLog.ChangeType,
		auditLog.ChangedBy,
		auditLog.ChangedAt,
		auditLog.ChangeReason,
		auditLog.Metadata,
	}, pgx.RowToStructByName[domain.ProposalAuditLog])

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no audit log returned after insert")
	}

	auditLog.AuditID = results[0].AuditID
	return nil
}

// CreateAuditLogs creates multiple audit log entries in a batch
func (r *ProposalRepository) CreateAuditLogs(ctx context.Context, auditLogs []*domain.ProposalAuditLog) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	if len(auditLogs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, auditLog := range auditLogs {
		batch.Queue(`
			INSERT INTO policy_issue.proposal_audit_log (
				proposal_id, entity_type, entity_id, field_name, old_value, new_value,
				change_type, changed_by, changed_at, change_reason, metadata
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			auditLog.ProposalID,
			auditLog.EntityType,
			auditLog.EntityID,
			auditLog.FieldName,
			auditLog.OldValue,
			auditLog.NewValue,
			auditLog.ChangeType,
			auditLog.ChangedBy,
			auditLog.ChangedAt,
			auditLog.ChangeReason,
			auditLog.Metadata,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to execute audit log batch statement %d: %w", i, err)
		}
	}

	return nil
}

// GetAuditLogsByProposal retrieves audit logs for a specific proposal
func (r *ProposalRepository) GetAuditLogsByProposal(ctx context.Context, proposalID int64) ([]domain.ProposalAuditLog, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	sql := `
		SELECT audit_id, proposal_id, entity_type, entity_id, field_name, old_value, new_value,
		       change_type, changed_by, changed_at, change_reason, metadata
		FROM policy_issue.proposal_audit_log
		WHERE proposal_id = $1
		ORDER BY changed_at DESC, audit_id DESC
	`

	results, err := dblib.ExecReturns(ctx, r.db, sql, []any{proposalID}, pgx.RowToStructByName[domain.ProposalAuditLog])
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return results, nil
}

// GetAuditLogsByEntity retrieves audit logs for a specific entity
func (r *ProposalRepository) GetAuditLogsByEntity(ctx context.Context, entityType string, entityID int64) ([]domain.ProposalAuditLog, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	sql := `
		SELECT audit_id, proposal_id, entity_type, entity_id, field_name, old_value, new_value,
		       change_type, changed_by, changed_at, change_reason, metadata
		FROM policy_issue.proposal_audit_log
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY changed_at DESC, audit_id DESC
	`

	results, err := dblib.ExecReturns(ctx, r.db, sql, []any{entityType, entityID}, pgx.RowToStructByName[domain.ProposalAuditLog])
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by entity: %w", err)
	}

	return results, nil
}

// valueToString converts a value to string pointer for audit logging
func valueToString(value interface{}) *string {
	if value == nil {
		return nil
	}

	// Handle pointers
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
		value = val.Interface()
	}

	// Convert to string
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		str = fmt.Sprintf("%v", v)
	case float32, float64:
		str = fmt.Sprintf("%v", v)
	case bool:
		str = fmt.Sprintf("%v", v)
	case time.Time:
		str = v.Format(time.RFC3339)
	default:
		// For complex types, use JSON or simple string representation
		str = fmt.Sprintf("%v", v)
	}

	return &str
}

func (r *ProposalRepository) UpdatePolicyNumber(ctx context.Context, proposalID int64, policyNumber string) error {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	batch := &pgx.Batch{}

	// Update proposals table
	batch.Queue(`
		UPDATE policy_issue.proposals
		SET policy_number = $1,
		    updated_at = now()
		WHERE proposal_id = $2
	`, policyNumber, proposalID)

	// Update proposal_nominee table
	batch.Queue(`
		UPDATE policy_issue.proposal_nominee
		SET policy_number = $1
		WHERE proposal_id = $2
	`, policyNumber, proposalID)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	// Execute both statements
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed policy number update batch statement %d: %w", i, err)
		}
	}

	return nil
}

// MarkPMSignalSent records a successful PM SignalWithStart for the given policy.
func (r *ProposalRepository) MarkPMSignalSent(ctx context.Context, policyNumber, plwWorkflowID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Update("policy_issue.proposal_issuance").
		Set("pm_signal_status", "SENT").
		Set("pm_signal_sent_at", time.Now().UTC()).
		Set("pm_plw_workflow_id", plwWorkflowID).
		Set("pm_signal_last_error", nil).
		Where(sq.Eq{"policy_number": policyNumber})

	cmdTag, err := dblib.Update(ctx, r.db, query)
	if err != nil {
		return fmt.Errorf("MarkPMSignalSent: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("MarkPMSignalSent: no issuance row found for policy %s", policyNumber)
	}
	return nil
}

// IncrementPMSignalAttempts bumps the pm_signal_attempts counter without
// changing pm_signal_status. Called before a SignalWithStart attempt so the
// count is accurate regardless of whether the call succeeds or fails.
func (r *ProposalRepository) IncrementPMSignalAttempts(ctx context.Context, policyNumber string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Update("policy_issue.proposal_issuance").
		Set("pm_signal_attempts", sq.Expr("pm_signal_attempts + 1")).
		Where(sq.Eq{"policy_number": policyNumber})

	_, err := dblib.Update(ctx, r.db, query)
	if err != nil {
		return fmt.Errorf("IncrementPMSignalAttempts: %w", err)
	}
	return nil
}

// MarkPMSignalFailed records a failed PM SignalWithStart attempt.
func (r *ProposalRepository) MarkPMSignalFailed(ctx context.Context, policyNumber, errMsg string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Update("policy_issue.proposal_issuance").
		Set("pm_signal_status", "FAILED").
		Set("pm_signal_last_error", errMsg).
		Where(sq.Eq{"policy_number": policyNumber})

	_, err := dblib.Update(ctx, r.db, query)
	if err != nil {
		return fmt.Errorf("MarkPMSignalFailed: %w", err)
	}
	return nil
}

// PMSignalTarget holds the data needed to replay a PM SignalWithStart.
type PMSignalTarget struct {
	PolicyNumber    string    `db:"policy_number"`
	ProposalID      int64     `db:"proposal_id"`
	PolicyType      string    `db:"policy_type"`
	PolicyIssueDate time.Time `db:"policy_issue_date"`
	Attempts        int       `db:"pm_signal_attempts"`
}

// FindUnsignalledPolicies returns policies whose PM signal is PENDING (older than
// gracePeriod) or FAILED, and whose attempt count is below maxAttempts.
func (r *ProposalRepository) FindUnsignalledPolicies(ctx context.Context, gracePeriod time.Duration, maxAttempts int) ([]PMSignalTarget, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	cutoff := time.Now().UTC().Add(-gracePeriod)

	rows, err := r.db.Query(ctx, `
		SELECT
			pi.policy_number,
			pi.proposal_id,
			p.policy_type,
			pi.policy_issue_date,
			pi.pm_signal_attempts
		FROM policy_issue.proposal_issuance pi
		JOIN policy_issue.proposals p USING (proposal_id)
		WHERE (
			(pi.pm_signal_status = 'PENDING'  AND pi.policy_issue_date < $1)
			OR pi.pm_signal_status = 'FAILED'
		)
		AND pi.pm_signal_attempts < $2
		ORDER BY pi.policy_issue_date
		LIMIT 100
	`, cutoff, maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("FindUnsignalledPolicies query: %w", err)
	}
	defer rows.Close()

	var targets []PMSignalTarget
	for rows.Next() {
		var t PMSignalTarget
		if err := rows.Scan(
			&t.PolicyNumber,
			&t.ProposalID,
			&t.PolicyType,
			&t.PolicyIssueDate,
			&t.Attempts,
		); err != nil {
			return nil, fmt.Errorf("FindUnsignalledPolicies scan: %w", err)
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}
