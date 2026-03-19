package repo

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	//"gitlab.cept.gov.in/it-2.0-policy/surrender-service/handler"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
)

const (
	contenttypeliteral     = "Content-Type"
	applicationjsonliteral = "application/json"
)

// SurrenderRequestRepository handles all database operations for surrender requests
// Business Rules: BR-SUR-001 to BR-SUR-018, BR-FS-001 to BR-FS-018
type SurrenderRequestRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewSurrenderRequestRepository creates a new surrender request repository
func NewSurrenderRequestRepository(db *dblib.DB, cfg *config.Config) *SurrenderRequestRepository {
	return &SurrenderRequestRepository{
		db:  db,
		cfg: cfg,
	}
}

const surrenderRequestTable = "policy_surrender_requests"

// Column list excluding search_vector (auto-managed by trigger)
const surrenderRequestColumns = `id, policy_id, request_number, request_type, previous_policy_status,
	request_date, surrender_value_calculated_date, gross_surrender_value,
	net_surrender_value, paid_up_value, bonus_amount, surrender_factor,
	unpaid_premiums_deduction, loan_deduction, other_deductions,
	disbursement_method, disbursement_amount, reason, status,
	owner, created_at, updated_at, created_by, approved_by, approved_at,
	approval_comments, deleted_at, version, metadata, search_vector`

// Create inserts a new surrender request
// Functional Requirement: FR-SUR-001, FR-FS-001
// Business Rule: BR-SUR-013 (for voluntary), BR-FS-004 (for forced)
func (r *SurrenderRequestRepository) Create(ctx context.Context, data domain.PolicySurrenderRequest) (domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(surrenderRequestTable).
		Columns(
			"policy_id", "request_number", "request_type", "previous_policy_status",
			"request_date", "surrender_value_calculated_date", "gross_surrender_value",
			"net_surrender_value", "paid_up_value", "bonus_amount", "surrender_factor",
			"unpaid_premiums_deduction", "loan_deduction", "other_deductions",
			"disbursement_method", "disbursement_amount", "reason", "status",
			"owner", "created_by", "metadata",
		).
		Values(
			data.PolicyID, data.RequestNumber, data.RequestType, data.PreviousPolicyStatus,
			data.RequestDate, data.SurrenderValueCalculatedDate, data.GrossSurrenderValue,
			data.NetSurrenderValue, data.PaidUpValue, data.BonusAmount, data.SurrenderFactor,
			data.UnpaidPremiumsDeduction, data.LoanDeduction, data.OtherDeductions,
			data.DisbursementMethod, data.DisbursementAmount, data.Reason, data.Status,
			data.Owner, data.CreatedBy, data.Metadata,
		).
		Suffix("RETURNING " + surrenderRequestColumns).
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindByID retrieves a surrender request by ID
// Functional Requirement: FR-SUR-005
func (r *SurrenderRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(surrenderRequestColumns).
		From(surrenderRequestTable).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil}).
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		if err == pgx.ErrNoRows {
			return result, err
		}
		return result, err
	}

	return result, nil
}

func (r *SurrenderRequestRepository) FindByPolicyNumber(ctx context.Context, policyno string) (*domain.PolicyDetailsOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Finding policy details for policy number: %s", policyno)

	query := dblib.Psql.Select("policy_number", "customer_id", "customer_name", "product_code", "product_name", "policy_status", "premium_frequency", "premium_amount",
		"sum_assured", "revival_count", "paid_to_date", "maturity_date", "date_of_commencement", "last_revival_date", "polissdate", "outstandingloanprinciple",
		"Outstandingloaninterest", "totalbonus", "dob").
		From("finservicemgmt.policies").
		Where(sq.Eq{"policy_number": policyno})

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.PolicyDetailsOutput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("policy number %s not found", policyno)
		}
		return nil, err
	}

	return &result, nil

}

//func (r *SurrenderRequestRepository) IndexSurrenderRequest(ctx context.Context, policyno string, channel string, ioid int, coid int, createdby int, modby int, remarks string) (string, error) {

func (r *SurrenderRequestRepository) IndexSurrenderRequestRepo(ctx context.Context, req domain.IndexSurrenderRequestInput2) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Finding policy details for policy number inside Repo: %s", req.PolicyNumber)
	istLocation, _ := time.LoadLocation("Asia/Kolkata")
	currentTimeIST := time.Now().In(istLocation)

	serviceReqID := "SUR-" + req.PolicyNumber + "-" + currentTimeIST.Format("20060102150405")
	logger.Infof("Request: %s", serviceReqID)

	batch := &pgx.Batch{}

	queryInsert := dblib.Psql.Insert("finservicemgmt.surrender_requests").
		Columns("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "stage_name", "indexing_office_id", "cpc_office_id", "created_by", "modified_by", "remarks", "temporal_workflow_id", "pm_service_request_id", "pm_policy_db_id").
		Values(serviceReqID, "IT2", "Surrender", req.PolicyNumber, req.Stage_name, req.Indexing_office_id, req.Cpc_office_id, req.Created_by, req.Modified_by, req.Remarks, req.TemporalWorkflowID, int64(req.PMServiceRequestID),
			int64(req.PMPolicyDBID))

	dblib.QueueExecRow(batch, queryInsert)

	queryInsert1 := dblib.Psql.Insert("finservicemgmt.surrender_requests_stages").
		Columns("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "current_stage_name", "created_by", "cpc_office_id", "remarks").
		Values(serviceReqID, "IT2", "Surrender", req.PolicyNumber, req.Stage_name, req.Created_by, req.Cpc_office_id, req.Remarks)

	dblib.QueueExecRow(batch, queryInsert1)

	queryInsert2 := dblib.Psql.Insert("finservicemgmt.surrender_requests_attributes").
		Columns("surrender_request_id", "policy_number", "paidupvalue", "bonus", "grossamount", "loanprincipal", "loaninterest", "surrenderfactor", "othercharges", "surrendervalue", "bonusrate", "bonusamount", "sumassured", "paid_to_date", "polissdate", "maturitydate", "productcode", "dob", "unpaidprem", "def").
		Values(serviceReqID, req.PolicyNumber, req.Paidupvalue, req.Bonus, req.Grossamount, req.Loanprincipal, req.Loaninterest, req.Surrenderfactor, req.Othercharges, req.Surrendervalue, req.Bonusrate, req.Bonusamount, req.Sumassured, req.Paid_to_date, req.Polissdate, req.Maturitydate, req.Productcode, req.Dob, req.Unpaidprem, req.Def)

	dblib.QueueExecRow(batch, queryInsert2)

	results := r.db.SendBatch(ctx, batch).Close()
	if results != nil {
		fmt.Print("batch prepared with error")
		return serviceReqID, results
	}

	return serviceReqID, nil

}

// GetWorkflowIDBySurrenderRequestID looks up the Temporal workflow ID stored
// in surrender_requests.temporal_workflow_id so DE/QC/Approval handlers can
// signal the correct running SurrenderProcessingWorkflow instance.
func (r *SurrenderRequestRepository) GetWorkflowIDBySurrenderRequestID(ctx context.Context, srID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("temporal_workflow_id").
		From("finservicemgmt.surrender_requests").
		Where(sq.Eq{"surrender_request_id": srID}).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build query: %w", err)
	}

	var workflowID string
	err = r.db.QueryRow(ctx, sql, args...).Scan(&workflowID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("surrender request %s not found", srID)
		}
		return "", fmt.Errorf("failed to get workflow ID for surrender request %s: %w", srID, err)
	}

	return workflowID, nil
}

func (r *SurrenderRequestRepository) SRDetailsRepo(ctx context.Context, srID string) ([]domain.SRDetailsOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("ServiceReqID: %s", srID)

	query := dblib.Psql.Select("surrender_request_id", "policy_number", "paidupvalue", "bonus", "grossamount", "loanprincipal", "loaninterest", "surrenderfactor",
		"othercharges", "surrendervalue", "bonusrate", "bonusamount", "paymentmode", "bankname", "micrcode", "accounttype", "ifsccode", "accountnumber", "accountholdername",
		"branchname", "banktype", "ismicrvalidated", "policybond", "lrrb", "prb", "pdo_certificate", "application", "idproof_insurant", "addressproof_insurant", "idproof_messenger", "addressproof_messenger", "account_details_proof", "reason", "remarks",
		"sumassured", "paid_to_date", "polissdate", "maturitydate", "productcode", "dob", "unpaidprem", "def", "others").
		From("finservicemgmt.surrender_requests_attributes").
		Where(sq.Eq{"surrender_request_id": srID})

	result, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.SRDetailsOutput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("Service Request  %s not found", srID)
		}
		return nil, err
	}

	return result, nil

}

func (r *SurrenderRequestRepository) ServiceReqStagingDetailsRepo(ctx context.Context, srID string) ([]domain.SRStagingDetailsOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("ServiceReqID: %s", srID)

	query := dblib.Psql.Select("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "current_stage_name", "created_by", "created_date", "cpc_office_id", "remarks").
		From("finservicemgmt.surrender_requests_stages").
		Where(sq.Eq{"surrender_request_id": srID})

	result, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.SRStagingDetailsOutput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("Service Request  %s not found", srID)
		}
		return nil, err
	}

	return result, nil

}

func (r *SurrenderRequestRepository) DEPendingRepo(ctx context.Context, oid int) ([]domain.PendingRequestOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Ofice ID: %s", oid)

	query := dblib.Psql.Select("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "stage_name", "indexing_office_id", "cpc_office_id", "created_by", "created_date", "remarks").
		From("finservicemgmt.surrender_requests").
		Where(sq.And{
			sq.Eq{"cpc_office_id": oid},
			sq.Eq{"stage_name": "Indexed"},
		})

	result, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.PendingRequestOutput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("No details for the office   %s", oid)
		}
		return nil, err
	}

	return result, nil

}

func (r *SurrenderRequestRepository) QCPendingRepo(ctx context.Context, oid int) ([]domain.PendingRequestOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Ofice ID: %s", oid)

	query := dblib.Psql.Select("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "stage_name", "indexing_office_id", "cpc_office_id", "created_by", "created_date", "remarks").
		From("finservicemgmt.surrender_requests").
		Where(sq.And{
			sq.Eq{"cpc_office_id": oid},
			sq.Eq{"stage_name": "DataEntry"},
		})

	result, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.PendingRequestOutput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("No details for the office   %s", oid)
		}
		return nil, err
	}

	return result, nil

}

func (r *SurrenderRequestRepository) ApprovalPendingRepo(ctx context.Context, oid int) ([]domain.PendingRequestOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Ofice ID: %s", oid)

	query := dblib.Psql.Select("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "stage_name", "indexing_office_id", "cpc_office_id", "created_by", "created_date", "remarks").
		From("finservicemgmt.surrender_requests").
		Where(sq.And{
			sq.Eq{"cpc_office_id": oid},
			sq.Eq{"stage_name": "QualityCheck"},
		})

	result, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.PendingRequestOutput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("No details for the office   %s", oid)
		}
		return nil, err
	}

	return result, nil

}

func (r *SurrenderRequestRepository) AllReqPendingRepo(ctx context.Context, oid int) ([]domain.PendingRequestOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Ofice ID: %s", oid)

	query := dblib.Psql.Select("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "stage_name", "indexing_office_id", "cpc_office_id", "created_by", "created_date", "remarks").
		From("finservicemgmt.surrender_requests").
		Where(sq.And{
			sq.Eq{"cpc_office_id": oid},
			//sq.Eq{"stage_name": "QualityCheck"},
		})

	result, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.PendingRequestOutput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("No details for the office   %s", oid)
		}
		return nil, err
	}

	return result, nil

}

func (r *SurrenderRequestRepository) SubmitDERepo(ctx context.Context, req domain.SubmitDERequest) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Req ID: %s", req.Surrender_request_id)

	batch := &pgx.Batch{}
	queryInsert1 := dblib.Psql.Insert("finservicemgmt.surrender_requests_stages").
		Columns("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "current_stage_name", "created_by", "cpc_office_id", "remarks").
		Values(req.Surrender_request_id, req.Surrender_request_channel, req.Request_name, req.PolicyNumber, req.Current_stage_name, req.Created_by, req.Cpc_office_id, req.Remarks)

	dblib.QueueExecRow(batch, queryInsert1)

	queryInsert2 := dblib.Psql.Update("finservicemgmt.surrender_requests").
		Set("stage_name", req.Current_stage_name).
		Set("modified_by", req.Modified_by).
		Set("modified_date", time.Now()).
		Set("remarks", req.Remarks).
		Where(sq.And{
			//sq.Eq{"transaction_id": receiptNumber},
			sq.Eq{"surrender_request_id": req.Surrender_request_id},
		})

	dblib.QueueExecRow(batch, queryInsert2)

	queryInsert3 := dblib.Psql.Update("finservicemgmt.surrender_requests_attributes").
		Set("paymentmode", req.Paymentmode).
		Set("bankname", req.Bankname).
		Set("micrcode", req.Micrcode).
		Set("accounttype", req.Accounttype).
		Set("ifsccode", req.Ifsccode).
		Set("accountnumber", req.Accountnumber).
		Set("accountholdername", req.Accountholdername).
		Set("branchname", req.Branchname).
		Set("banktype", req.Banktype).
		Set("ismicrvalidated", req.Ismicrvalidated).
		Set("policybond", req.Policybond).
		Set("lrrb", req.Lrrb).
		Set("prb", req.Prb).
		Set("pdo_certificate", req.Pdo_certificate).
		Set("application", req.Application).
		Set("idproof_insurant", req.Idproof_insurant).
		Set("addressproof_insurant", req.Addressproof_insurant).
		Set("idproof_messenger", req.Idproof_messenger).
		Set("addressproof_messenger", req.Addressproof_messenger).
		Set("account_details_proof", req.Account_details_proof).
		Set("others", req.Others).
		Where(sq.And{
			//sq.Eq{"transaction_id": receiptNumber},
			sq.Eq{"surrender_request_id": req.Surrender_request_id},
		})

	dblib.QueueExecRow(batch, queryInsert3)

	results := r.db.SendBatch(ctx, batch).Close()
	if results != nil {
		fmt.Print("batch prepared with error")
		return "Error in inserting", results
	}

	return "Data Entry Submitted", nil

}

func (r *SurrenderRequestRepository) SubmitQCRepo(ctx context.Context, req domain.SubmitQCRequest) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()
	logger.Infof("Inside SubmitQCRepo for Req ID:")
	logger.Infof("Req ID: %s", req.Surrender_request_id)

	batch := &pgx.Batch{}
	queryInsert1 := dblib.Psql.Insert("finservicemgmt.surrender_requests_stages").
		Columns("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "current_stage_name", "created_by", "cpc_office_id", "remarks").
		Values(req.Surrender_request_id, req.Surrender_request_channel, req.Request_name, req.PolicyNumber, req.Current_stage_name, req.Created_by, req.Cpc_office_id, req.Remarks)

	dblib.QueueExecRow(batch, queryInsert1)

	queryInsert2 := dblib.Psql.Update("finservicemgmt.surrender_requests").
		Set("stage_name", req.Current_stage_name).
		Set("modified_by", req.Modified_by).
		Set("modified_date", time.Now()).
		Set("remarks", req.Remarks).
		Where(sq.And{
			//sq.Eq{"transaction_id": receiptNumber},
			sq.Eq{"surrender_request_id": req.Surrender_request_id},
		})

	dblib.QueueExecRow(batch, queryInsert2)

	queryInsert3 := dblib.Psql.Update("finservicemgmt.surrender_requests_attributes").
		Set("paymentmode", req.Paymentmode).
		Set("bankname", req.Bankname).
		Set("micrcode", req.Micrcode).
		Set("accounttype", req.Accounttype).
		Set("ifsccode", req.Ifsccode).
		Set("accountnumber", req.Accountnumber).
		Set("accountholdername", req.Accountholdername).
		Set("branchname", req.Branchname).
		Set("banktype", req.Banktype).
		Set("ismicrvalidated", req.Ismicrvalidated).
		Set("policybond", req.Policybond).
		Set("lrrb", req.Lrrb).
		Set("prb", req.Prb).
		Set("pdo_certificate", req.Pdo_certificate).
		Set("application", req.Application).
		Set("idproof_insurant", req.Idproof_insurant).
		Set("addressproof_insurant", req.Addressproof_insurant).
		Set("idproof_messenger", req.Idproof_messenger).
		Set("addressproof_messenger", req.Addressproof_messenger).
		Set("account_details_proof", req.Account_details_proof).
		Set("others", req.Others).
		Where(sq.And{
			//sq.Eq{"transaction_id": receiptNumber},
			sq.Eq{"surrender_request_id": req.Surrender_request_id},
		})

	dblib.QueueExecRow(batch, queryInsert3)

	results := r.db.SendBatch(ctx, batch).Close()
	if results != nil {
		fmt.Print("batch prepared with error")
		return "Error in inserting", results
	}

	return "Quality Check Submitted", nil

}

func (r *SurrenderRequestRepository) SubmitApprovalRepo(ctx context.Context, req domain.SubmitApprovalRequest) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Req ID: %s", req.Surrender_request_id)

	batch := &pgx.Batch{}
	queryInsert1 := dblib.Psql.Insert("finservicemgmt.surrender_requests_stages").
		Columns("surrender_request_id", "surrender_request_channel", "request_name", "policy_number", "current_stage_name", "created_by", "cpc_office_id", "remarks").
		Values(req.Surrender_request_id, req.Surrender_request_channel, req.Request_name, req.PolicyNumber, req.Current_stage_name, req.Created_by, req.Cpc_office_id, req.Remarks)

	dblib.QueueExecRow(batch, queryInsert1)

	queryInsert2 := dblib.Psql.Update("finservicemgmt.surrender_requests").
		Set("stage_name", req.Current_stage_name).
		Set("modified_by", req.Modified_by).
		Set("modified_date", time.Now()).
		Set("remarks", req.Remarks).
		Where(sq.And{
			//sq.Eq{"transaction_id": receiptNumber},
			sq.Eq{"surrender_request_id": req.Surrender_request_id},
		})

	dblib.QueueExecRow(batch, queryInsert2)

	queryInsert3 := dblib.Psql.Update("finservicemgmt.surrender_requests_attributes").
		Set("paymentmode", req.Paymentmode).
		Set("bankname", req.Bankname).
		Set("micrcode", req.Micrcode).
		Set("accounttype", req.Accounttype).
		Set("ifsccode", req.Ifsccode).
		Set("accountnumber", req.Accountnumber).
		Set("accountholdername", req.Accountholdername).
		Set("branchname", req.Branchname).
		Set("banktype", req.Banktype).
		Set("ismicrvalidated", req.Ismicrvalidated).
		Set("policybond", req.Policybond).
		Set("lrrb", req.Lrrb).
		Set("prb", req.Prb).
		Set("pdo_certificate", req.Pdo_certificate).
		Set("application", req.Application).
		Set("idproof_insurant", req.Idproof_insurant).
		Set("addressproof_insurant", req.Addressproof_insurant).
		Set("idproof_messenger", req.Idproof_messenger).
		Set("addressproof_messenger", req.Addressproof_messenger).
		Set("account_details_proof", req.Account_details_proof).
		Set("others", req.Others).
		Where(sq.And{
			//sq.Eq{"transaction_id": receiptNumber},
			sq.Eq{"surrender_request_id": req.Surrender_request_id},
		})

	dblib.QueueExecRow(batch, queryInsert3)

	results := r.db.SendBatch(ctx, batch).Close()
	if results != nil {
		fmt.Print("batch prepared with error")
		return "Error in inserting", results
	}

	return "Approval Submitted", nil

}

func yearsBetween(from, to time.Time) int {
	years := to.Year() - from.Year()

	// Adjust if current date is before anniversary
	if to.Month() < from.Month() ||
		(to.Month() == from.Month() && to.Day() < from.Day()) {
		years--
	}
	return years
}

func (r *SurrenderRequestRepository) CalcSurrenderValuerepo(ctx context.Context, policyno string, prodcode string, polissuedate time.Time, matdate time.Time, dob time.Time) ([]domain.SurrenderFactorOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	logger.Infof("Req ID: %s", policyno)
	logger.Info("Step 1 - Before reading config")

	sfurl := strings.TrimSpace(r.cfg.GetString("url.SFUrl"))
	logger.Infof("Step 2 - Raw URL from config: %s", sfurl)

	// sfurl = strings.Replace(sfurl, "#1", strconv.FormatInt(5001, 10), 1)
	// sfurl = strings.Replace(sfurl, "#2", strconv.FormatInt(20, 10), 1)
	// sfurl = strings.Replace(sfurl, "#3", strconv.FormatInt(35, 10), 1)

	sfurl = strings.Replace(sfurl, "#1", prodcode, 1)
	currentDate := time.Now()
	policyYears := yearsBetween(dob, currentDate)
	if policyYears < 0 {
		return nil, fmt.Errorf("invalid policy years calculated: %d", policyYears)
	}

	sfurl = strings.Replace(sfurl, "#2", strconv.Itoa(policyYears), 1)

	maturityAge := yearsBetween(dob, matdate)

	if maturityAge < 0 {
		return nil, fmt.Errorf("invalid maturity age calculated: %d", maturityAge)
	}
	sfurl = strings.Replace(sfurl, "#3", strconv.Itoa(maturityAge), 1)

	// logger.Info("Req ID: %s", prodcode)
	// logger.Info("Req ID: %d", policyYears)
	// logger.Info("Req ID: %d", maturityAge)

	logger.Infof("Policy Issue Date: %v", polissuedate)
	logger.Infof("Maturity Date: %v", matdate)
	logger.Infof("DOB: %v", dob)

	finalURL := sfurl
	fmt.Println(finalURL, "sfurl")

	logger.Infof("Step 3 - Final URL: %s", sfurl)

	if sfurl == "" {
		return nil, fmt.Errorf("SFUrl is empty in config")
	}

	surrenderFactor, err := fetchSurrenderFactor(finalURL)
	if err != nil {
		return nil, fmt.Errorf("error:%s while calling pis API: %s",
			err.Error(), sfurl)
	}
	if surrenderFactor == nil {
		return nil, fmt.Errorf("Employee data not found in pis")
	}

	//sfFulldata, ok := surrenderFactor["data"].(map[string]interface{})
	sfFulldata, ok := surrenderFactor["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format in response")
	}
	var sfFullData []domain.SurrenderFactorOutput
	srDataJson, err := json.Marshal(sfFulldata)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(srDataJson, &sfFullData); err != nil {
		return nil, err

	}

	return sfFullData, nil

}

func (r *SurrenderRequestRepository) CalcBonusValuerepo(ctx context.Context, prodcode string, polissuedate time.Time) ([]domain.BonusOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()
	logger.Info("Step 1 - Before reading config")

	bonusurl := strings.TrimSpace(r.cfg.GetString("url.Bonus"))
	logger.Infof("Step 2 - Raw URL from config: %s", bonusurl)

	bonusurl = strings.Replace(bonusurl, "#1", prodcode, 1)
	currentDate := time.Now()
	currentYears := currentDate.Year()

	bonusurl = strings.Replace(bonusurl, "#2", strconv.Itoa(polissuedate.Year()), 1)

	bonusurl = strings.Replace(bonusurl, "#3", strconv.Itoa(currentYears), 1)

	logger.Infof("Policy Issue Date: %v", polissuedate)

	finalURL := bonusurl
	fmt.Println(finalURL, "sfurl")

	logger.Infof("Step 3 - Final URL: %s", bonusurl)

	if bonusurl == "" {
		return nil, fmt.Errorf("SFUrl is empty in config")
	}

	bonusDetails, err := fetchBonusValues(finalURL)
	if err != nil {
		return nil, fmt.Errorf("error:%s while calling pis API: %s",
			err.Error(), bonusurl)
	}
	if bonusDetails == nil {
		return nil, fmt.Errorf("Employee data not found in pis")
	}
	bonusFulldata, ok := bonusDetails["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format in response")
	}
	var bonusFullData []domain.BonusOutput
	bonusDataJson, err := json.Marshal(bonusFulldata)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bonusDataJson, &bonusFullData); err != nil {
		return nil, err

	}

	return bonusFullData, nil

}

func fetchBonusValues(url string) (map[string]interface{}, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
			Renegotiation:      tls.RenegotiateOnceAsClient,
		},
		DisableKeepAlives: true,
	}
	header := map[string]string{
		"User-Agent":   "MyAPI/1.0",
		"Content-Type": applicationjsonliteral,
	}
	client := resty.New().
		SetTimeout(30 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client.SetTransport(tr)
	response, err := client.R().
		SetHeaders(header).
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, err
	}
	if response.IsSuccess() {
		body := response.Body()
		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			return nil, fmt.Errorf("error unmarshaling response body: %w", err)
		}
		return responseData, nil
	} else {
		if response.StatusCode() != 404 {
			return nil, fmt.Errorf("failed request, status code: %d, message: %s",
				response.StatusCode(), response.String())
		}
		return nil, nil
	}
}

func fetchSurrenderFactor(url string) (map[string]interface{}, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
			Renegotiation:      tls.RenegotiateOnceAsClient,
		},
		DisableKeepAlives: true,
	}
	header := map[string]string{
		"User-Agent":   "MyAPI/1.0",
		"Content-Type": applicationjsonliteral,
	}
	client := resty.New().
		SetTimeout(30 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client.SetTransport(tr)
	response, err := client.R().
		SetHeaders(header).
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, err
	}
	if response.IsSuccess() {
		body := response.Body()
		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			return nil, fmt.Errorf("error unmarshaling response body: %w", err)
		}
		return responseData, nil
	} else {
		if response.StatusCode() != 404 {
			return nil, fmt.Errorf("failed request, status code: %d, message: %s",
				response.StatusCode(), response.String())
		}
		return nil, nil
	}
}

// FindByPolicyID retrieves surrender requests by policy ID
// Business Rule: BR-SUR-014 (only one active surrender per policy)
func (r *SurrenderRequestRepository) FindByPolicyID(ctx context.Context, policyID string) ([]domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select(surrenderRequestColumns).
		From(surrenderRequestTable).
		Where(sq.Eq{"policy_id": policyID}).
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar)

	results, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		return results, err
	}

	return results, nil
}

// FindActiveByPolicyID retrieves active surrender request by policy ID
// Business Rule: BR-SUR-014 (check for existing active surrender)
func (r *SurrenderRequestRepository) FindActiveByPolicyID(ctx context.Context, policyID string) (domain.PolicySurrenderRequest, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	activeStatuses := []string{
		string(domain.SurrenderStatusPendingDocumentUpload),
		string(domain.SurrenderStatusPendingVerification),
		string(domain.SurrenderStatusPendingApproval),
		string(domain.SurrenderStatusPendingAutoCompletion),
	}

	query := dblib.Psql.Select(surrenderRequestColumns).
		From(surrenderRequestTable).
		Where(sq.Eq{"policy_id": policyID}).
		Where(sq.Eq{"status": activeStatuses}).
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("created_at DESC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	result, found, err := dblib.SelectOneOK(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	return result, found, err
}

// FindByRequestNumber retrieves a surrender request by request number
func (r *SurrenderRequestRepository) FindByRequestNumber(ctx context.Context, requestNumber string) (domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(surrenderRequestColumns).
		From(surrenderRequestTable).
		Where(sq.Eq{"request_number": requestNumber}).
		Where(sq.Eq{"deleted_at": nil}).
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		return result, err
	}

	return result, nil
}

// UpdateStatus updates the status of a surrender request
// Functional Requirement: FR-SUR-006, FR-FS-005
func (r *SurrenderRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SurrenderStatus, updatedBy uuid.UUID, comments *string) (domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(surrenderRequestTable).
		Set("status", status).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + surrenderRequestColumns).
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		return result, err
	}

	return result, nil
}

// UpdateApprovalInfo updates approval information
// Functional Requirement: FR-FS-007
func (r *SurrenderRequestRepository) UpdateApprovalInfo(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID, comments string) (domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	query := dblib.Psql.Update(surrenderRequestTable).
		Set("approved_by", approvedBy).
		Set("approved_at", now).
		Set("approval_comments", comments).
		Set("status", domain.SurrenderStatusApproved).
		Set("updated_at", now).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		return result, err
	}

	return result, nil
}

// UpdatePreviousPolicyStatus updates the previous policy status field
// Business Rule: BR-FS-018 (store status for reversion on rejection)
func (r *SurrenderRequestRepository) UpdatePreviousPolicyStatus(ctx context.Context, id uuid.UUID, previousStatus domain.PreviousPolicyStatus) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(surrenderRequestTable).
		Set("previous_policy_status", previousStatus).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	_, err := dblib.Update(ctx, r.db, query)
	return err
}

// List retrieves all surrender requests with pagination
// Functional Requirement: FR-SUR-005
func (r *SurrenderRequestRepository) List(ctx context.Context, skip, limit uint64, orderBy, sortType string, filters map[string]interface{}) ([]domain.PolicySurrenderRequest, uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Build base query for count
	countQuery := dblib.Psql.Select("COUNT(*)").
		From(surrenderRequestTable).
		Where(sq.Eq{"deleted_at": nil})

	// Build select query
	selectQuery := dblib.Psql.Select(surrenderRequestColumns).
		From(surrenderRequestTable).
		Where(sq.Eq{"deleted_at": nil})

	// Apply filters
	for key, value := range filters {
		countQuery = countQuery.Where(sq.Eq{key: value})
		selectQuery = selectQuery.Where(sq.Eq{key: value})
	}

	// Use batch for parallel execution
	batch := &pgx.Batch{}
	var totalCount uint64
	var results []domain.PolicySurrenderRequest

	// Queue count query
	if err := dblib.QueueReturnRow(batch,
		countQuery.PlaceholderFormat(sq.Dollar),
		func(row pgx.CollectableRow) (uint64, error) {
			var count uint64
			err := row.Scan(&count)
			return count, err
		},
		&totalCount); err != nil {
		return nil, 0, err
	}

	// Queue select query with pagination and sorting
	if orderBy == "" {
		orderBy = "created_at"
	}
	if sortType == "" {
		sortType = "DESC"
	}

	selectQuery = selectQuery.
		OrderBy(orderBy + " " + sortType).
		Limit(limit).
		Offset(skip).
		PlaceholderFormat(sq.Dollar)

	if err := dblib.QueueReturn(batch,
		selectQuery,
		pgx.RowToStructByName[domain.PolicySurrenderRequest],
		&results); err != nil {
		return nil, 0, err
	}

	// Execute batch
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	if err := br.Close(); err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

// ListPendingAutoCompletion retrieves forced surrender requests pending auto-completion
// Business Rule: BR-FS-007 (30-day payment window expiry)
// Temporal Workflow: TEMP-002
func (r *SurrenderRequestRepository) ListPendingAutoCompletion(ctx context.Context) ([]domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select(surrenderRequestColumns).
		From(surrenderRequestTable).
		Where(sq.Eq{
			"request_type": domain.SurrenderRequestTypeForced,
			"status":       domain.SurrenderStatusPendingAutoCompletion,
			"deleted_at":   nil,
		}).
		OrderBy("request_date ASC").
		PlaceholderFormat(sq.Dollar)

	results, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		return results, err
	}

	return results, nil
}

// RecalculateSurrenderValue updates the surrender value calculations
// Business Rule: BR-FS-010 (recalculation in approval queue)
func (r *SurrenderRequestRepository) RecalculateSurrenderValue(ctx context.Context, id uuid.UUID, grossSV, netSV, paidUpValue float64, bonusAmount *float64, loanDeduction, unpaidPremiums float64) (domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(surrenderRequestTable).
		Set("gross_surrender_value", grossSV).
		Set("net_surrender_value", netSV).
		Set("paid_up_value", paidUpValue).
		Set("bonus_amount", bonusAmount).
		Set("loan_deduction", loanDeduction).
		Set("unpaid_premiums_deduction", unpaidPremiums).
		Set("disbursement_amount", netSV).
		Set("surrender_value_calculated_date", time.Now()).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		return result, err
	}

	return result, nil
}

// SoftDelete soft deletes a surrender request
func (r *SurrenderRequestRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(surrenderRequestTable).
		Set("deleted_at", time.Now()).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	_, err := dblib.Update(ctx, r.db, query)
	return err
}

func (r *SurrenderRequestRepository) GetPolicyDetails(ctx context.Context, id uuid.UUID) (domain.PolicySurrenderRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(surrenderRequestTable).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil}).
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.PolicySurrenderRequest])
	if err != nil {
		if err == pgx.ErrNoRows {
			return result, err
		}
		return result, err
	}

	return result, nil
}
