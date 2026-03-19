package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"
	"policy-issue-service/workflows"

	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"

	// apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"go.temporal.io/sdk/client"
)

// ProposalHandler handles proposal-related HTTP endpoints
type ProposalHandler struct {
	*serverHandler.Base
	proposalRepo   *repo.ProposalRepository
	productRepo    *repo.ProductRepository
	quoteRepo      *repo.QuoteRepository
	cfg            *config.Config
	temporalClient client.Client
}

// NewProposalHandler creates a new ProposalHandler instance
func NewProposalHandler(proposalRepo *repo.ProposalRepository, productRepo *repo.ProductRepository, quoteRepo *repo.QuoteRepository, cfg *config.Config, temporalClient client.Client) *ProposalHandler {
	base := serverHandler.New("Proposals").SetPrefix("/v1").AddPrefix("")
	return &ProposalHandler{Base: base, proposalRepo: proposalRepo, productRepo: productRepo, quoteRepo: quoteRepo, cfg: cfg, temporalClient: temporalClient}
}

// Routes returns the routes for the ProposalHandler
func (h *ProposalHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/proposals/indexing", h.CreateProposalIndexing).Name("Create Proposal Indexing"),
		serverRoute.GET("/proposals/:proposal_id", h.GetProposal).Name("Get Proposal"),
		serverRoute.GET("/proposals/resolve/:proposal_number", h.ResolveProposalNumber).Name("Resolve Proposal Number"),
		serverRoute.POST("/proposals/:proposal_id/first-premium", h.RecordFirstPremium).Name("Record First Premium"),
		serverRoute.PUT("/proposals/:proposal_id/sections/insured", h.UpdateInsuredDetails).Name("Update Insured Details"),
		serverRoute.PUT("/proposals/:proposal_id/sections/nominees", h.UpdateNominees).Name("Update Nominees"),
		serverRoute.PUT("/proposals/:proposal_id/sections/policy-details", h.UpdatePolicyDetails).Name("Update Policy Details"),
		serverRoute.PUT("/proposals/:proposal_id/sections/agent", h.UpdateAgentDetails).Name("Update Agent Details"),
		serverRoute.PUT("/proposals/:proposal_id/sections/medical", h.UpdateMedicalInfo).Name("Update Medical Info"),
		serverRoute.PUT("/proposals/:proposal_id/sections/declaration", h.UpdateDeclaration).Name("Update Declaration"),
		serverRoute.PUT("/proposals/:proposal_id/sections/proposer", h.UpdateProposerDetails).Name("Update Proposer Details"),
		serverRoute.PUT("/proposals/:proposal_id/submit-for-qc", h.SubmitForQC).Name("Submit for QC"),
		serverRoute.POST("/proposals/:proposal_id/start-data-entry", h.StartDataEntry).Name("Start Data Entry"),
		serverRoute.GET("/proposals/:proposal_id/summary", h.GetProposalSummary).Name("Get Proposal Summary"),
		serverRoute.GET("/proposals/queue", h.GetProposalQueue).Name("Get Proposal Queue"),
		serverRoute.GET("/proposals/sections", h.GetProposalSection).Name("Get Proposal Sections"),

		// Audit Logs
		serverRoute.GET("/proposals/:proposal_id/audit-logs", h.GetProposalAuditLogs).Name("Get Proposal Audit Logs"),
		serverRoute.GET("/proposals/:proposal_id/entities/:entity_type/:entity_id/audit-logs", h.GetEntityAuditLogs).Name("Get Entity Audit Logs"),
	}
}

// CreateProposalIndexing creates a new proposal through CPC indexing
// [POL-API-005] Create Proposal Indexing
// [FR-POL-007] New Business Indexing
func (h *ProposalHandler) CreateProposalIndexing(sctx *serverRoute.Context, req ProposalIndexingRequest) (*resp.ProposalIndexingResponse, error) {

	// Parse dates
	declarationDate, err := time.Parse("2006-01-02", req.Dates.DeclarationDate)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid declaration_date format: %v", err)
		return nil, badRequest("Invalid declaration_date format. Expected YYYY-MM-DD")
	}
	receiptDate, err := time.Parse("2006-01-02", req.Dates.ReceiptDate)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid receipt_date format: %v", err)
		return nil, badRequest("Invalid receipt_date format. Expected YYYY-MM-DD")
	}
	indexingDate, err := time.Parse("2006-01-02", req.Dates.IndexingDate)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid indexing_date format: %v", err)
		return nil, badRequest("Invalid indexing_date format. Expected YYYY-MM-DD")
	}
	proposalDate, err := time.Parse("2006-01-02", req.Dates.ProposalDate)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid proposal_date format: %v", err)
		return nil, badRequest("Invalid proposal_date format. Expected YYYY-MM-DD")
	}

	// Validate date sequence: declaration_date <= receipt_date <= indexing_date <= proposal_date
	if declarationDate.After(receiptDate) || receiptDate.After(indexingDate) || indexingDate.After(proposalDate) {
		return nil, badRequest(domain.ErrInvalidDateSequence.Error())
	}

	// [BR-POL-024] Deduplication check — prevent duplicate proposals for same customer + product
	// existingProposal, err := h.proposalRepo.CheckDuplicateProposal(sctx.Ctx, req.ProductCode)
	// if err != nil {
	// 	log.Error(sctx.Ctx, "Failed to check for duplicate proposals", err)
	// 	return nil, err
	// }
	// if existingProposal != nil {
	// 	return &resp.ProposalIndexingResponse{
	// 		StatusCodeAndMessage: port.StatusCodeAndMessage{
	// 			StatusCode: http.StatusConflict,
	// 			Message:    fmt.Sprintf("ERR-POL-055: Duplicate proposal exists — proposal %s (status: %s) for customer %d with product %s is already in progress", existingProposal.ProposalNumber, existingProposal.Status, req.ProductCode),
	// 		},
	// 	}, nil
	// }

	// Create proposal domain object with all required fields
	quoteRef := req.QuoteRefNumber
	proposal := &domain.Proposal{
		PolicyType:   domain.PolicyType(req.PolicyType),
		ProductCode:  req.ProductCode,
		InsurantName: req.InsurantName,
		// CustomerID:              req.CustomerID,
		SpouseCustomerID:        req.SpouseCustomerID,
		Channel:                 domain.Channel(req.Channel),
		EntryPath:               domain.EntryPath(req.EntryPath),
		Status:                  domain.ProposalStatusIndexed,
		QuoteRefNumber:          &quoteRef,
		SumAssured:              req.SumAssured,
		PolicyTerm:              req.PolicyTerm,
		PremiumPaymentFrequency: domain.PremiumFrequency(req.PremiumPaymentFrequency),
		BasePremium:             req.BasePremium,
		GSTAmount:               req.GSTAmount,
		TotalPremium:            req.TotalPremium,
		CreatedBy:               req.CreatedBy,
	}

	// Create indexing data with dates
	indexingData := &domain.ProposalIndexing{
		POCode:          req.POCode,
		IssueCircle:     req.IssueCircle,
		IssueHO:         req.IssueHO,
		IssuePostOffice: req.IssuePostOffice,
		DeclarationDate: declarationDate,
		ReceiptDate:     receiptDate,
		IndexingDate:    indexingDate,
		ProposalDate:    proposalDate,
	}

	// Create proposal with indexing data - inserts into proposals, proposal_indexing, and proposal_status_history
	if err := h.proposalRepo.CreateProposalWithIndexing(sctx.Ctx, proposal, indexingData); err != nil {
		log.Error(sctx.Ctx, "Error creating proposal with indexing: %v", err)
		return nil, serverError("Failed to create proposal", err)
	}

	return &resp.ProposalIndexingResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusCreated,
			Message:    "Proposal indexed successfully",
		},
		ProposalID:     proposal.ProposalID,
		ProposalNumber: proposal.ProposalNumber,
		Status:         string(proposal.Status),
	}, nil
}

// GetProposal retrieves proposal details
// [POL-API-006] Get Proposal
func (h *ProposalHandler) GetProposal(sctx *serverRoute.Context, req ProposalIDUri) (*resp.ProposalDetailResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching proposal: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to fetch proposal")
	}

	return &resp.ProposalDetailResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposal retrieved successfully",
		},
		ProposalID:       proposal.ProposalID,
		ProposalNumber:   proposal.ProposalNumber,
		Status:           string(proposal.Status),
		PolicyType:       string(proposal.PolicyType),
		ProductCode:      proposal.ProductCode,
		CustomerID:       proposal.CustomerID,
		SpouseCustomerID: proposal.SpouseCustomerID,
		Channel:          string(proposal.Channel),
		SumAssured:       proposal.SumAssured,
		PolicyTerm:       proposal.PolicyTerm,
	}, nil
}

// ResolveProposalNumber resolves proposal number to ID
// [POL-API-017] Resolve Proposal Number
func (h *ProposalHandler) ResolveProposalNumber(sctx *serverRoute.Context, req ProposalNumberUri) (*resp.ResolveProposalResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByNumber(sctx.Ctx, req.ProposalNumber)
	if err != nil {
		log.Error(sctx.Ctx, "Error resolving proposal number: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to resolve proposal")
	}

	return &resp.ResolveProposalResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposal resolved successfully",
		},
		ProposalID: proposal.ProposalID,
	}, nil
}

// RecordFirstPremium records first premium payment
// [POL-API-007] Record First Premium
// [BR-POL-023] First Premium Collection
func (h *ProposalHandler) RecordFirstPremium(sctx *serverRoute.Context, req FirstPremiumRequest) (*resp.FirstPremiumResponse, error) {

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid payment_date format: %v", err)
		return nil, badRequest("Invalid payment_date format. Expected YYYY-MM-DD")
	}
	// Get proposal to verify it exists and is in correct status
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching proposal for first premium: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to fetch proposal")
	}

	// Validate proposal status - must be in DATA_ENTRY or later stage
	validStatuses := []domain.ProposalStatus{
		domain.ProposalStatusIndexed,
		domain.ProposalStatusDataEntry,
		domain.ProposalStatusQCPending,
		domain.ProposalStatusQCApproved,
	}
	validStatus := false
	for _, status := range validStatuses {
		if proposal.Status == status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return &resp.FirstPremiumResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Proposal must be in DATA_ENTRY or later stage to record first premium",
			},
		}, nil
	}

	// Record first premium payment
	if err := h.proposalRepo.RecordFirstPremium(sctx.Ctx, req.ProposalID, req.Amount,
		req.PaymentMethod, req.PaymentReference, paymentDate, req.CollectedBy); err != nil {
		log.Error(sctx.Ctx, "Error recording first premium: %v", err)
		return nil, serverError("Failed to record first premium", err)
	}

	// Update section completion for first premium
	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "premium", true); err != nil {
		log.Error(sctx.Ctx, "Error updating first premium section: %v", err)
		return nil, serverError("Failed to update premium section", err)
	}

	return &resp.FirstPremiumResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "First premium recorded successfully",
		},
		Status: "RECORDED",
	}, nil
}

// UpdateInsuredDetails updates insured person details
// [POL-API-008] Update Insured Details
func (h *ProposalHandler) UpdateInsuredDetails(sctx *serverRoute.Context, req InsuredDetailsRequest) (*resp.SectionUpdateResponse, error) {

	// Get proposal to verify it exists and is in DATA_ENTRY status
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.SectionUpdateResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Proposal not found",
				},
			}, nil
		}
		log.Error(sctx.Ctx, "Error fetching proposal: %v", err)
		return nil, err
	}

	// Validate status - must be in DATA_ENTRY
	if proposal.Status != domain.ProposalStatusDataEntry {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Proposal must be in DATA_ENTRY status",
			},
		}, nil
	}

	// Persist actual insured details
	insured := &domain.ProposalInsured{
		Salutation:    req.Salutation,
		FirstName:     req.FirstName,
		MiddleName:    &req.MiddleName,
		LastName:      req.LastName,
		Gender:        req.Gender,
		DateOfBirth:   req.DateOfBirth,
		MaritalStatus: &req.MaritalStatus,
		Occupation:    &req.Occupation,
		AnnualIncome:  &req.AnnualIncome,
		AddressLine1:  &req.AddressLine1,
		AddressLine2:  &req.AddressLine2,
		AddressLine3:  &req.AddressLine3,
		City:          &req.City,
		State:         &req.State,
		PinCode:       &req.PinCode,
		Mobile:        &req.Mobile,
		Email:         &req.Email,
	}

	if err := h.proposalRepo.SaveInsuredDetails(sctx.Ctx, req.ProposalID, req.CustomerID, insured, req.DataEntryBy); err != nil {
		log.Error(sctx.Ctx, "Error saving insured details: %v", err)
		return nil, handleRepoError(err, "Insured details not found", "Failed to save insured details")
	}

	// Update section completion status
	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "insured", true); err != nil {
		log.Error(sctx.Ctx, "Error updating insured section: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to update insured section")
	}

	return &resp.SectionUpdateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Insured details updated successfully",
		},
		Status:    "UPDATED",
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateNominees updates nominee details
// [POL-API-009] Update Nominees
func (h *ProposalHandler) UpdateNominees(sctx *serverRoute.Context, req NomineesRequest) (*resp.SectionUpdateResponse, error) {

	// Get proposal to verify it exists and is in DATA_ENTRY status
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching proposal: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to fetch proposal")
	}

	// Validate status - must be in DATA_ENTRY
	if proposal.Status != domain.ProposalStatusDataEntry {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Proposal must be in DATA_ENTRY status",
			},
		}, nil
	}

	// --- Cross-field nominee validation ---
	// [VAL-POL-004] Maximum 3 nominees
	const maxNominees = 3
	if len(req.Nominees) > maxNominees {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("ERR-POL-022: Maximum %d nominees allowed, got %d", maxNominees, len(req.Nominees)),
			},
		}, nil
	}

	// [VAL-POL-003] Nominee shares must total 100%
	var totalShare float64
	for _, n := range req.Nominees {
		totalShare += n.SharePercentage
	}
	// Use tolerance for floating point comparison
	if totalShare < 99.99 || totalShare > 100.01 {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("ERR-POL-023: Nominee shares must total 100%%, got %.2f%%", totalShare),
			},
		}, nil
	}

	// [VAL-POL-006] Appointee required for minor nominee
	for i, n := range req.Nominees {
		if n.IsMinor {
			if n.AppointeeName == "" || n.AppointeeRelationship == "" {
				return &resp.SectionUpdateResponse{
					StatusCodeAndMessage: port.StatusCodeAndMessage{
						StatusCode: http.StatusBadRequest,
						Message:    fmt.Sprintf("ERR-POL-024: Appointee name and relationship required for minor nominee at index %d (%s %s)", i, n.FirstName, n.LastName),
					},
				}, nil
			}
		}
	}

	// Persist nominee details using batch save to avoid N+1 query issue
	nominees := make([]*domain.ProposalNominee, len(req.Nominees))
	for i, n := range req.Nominees {
		var nomineeCustomerID *int64
		if n.NomineeCustomerID != nil {
			nomineeCustomerID = n.NomineeCustomerID
		}

		nominees[i] = &domain.ProposalNominee{
			Salutation:            n.Salutation,
			FirstName:             n.FirstName,
			MiddleName:            &n.MiddleName,
			LastName:              n.LastName,
			Gender:                n.Gender,
			DateOfBirth:           n.DateOfBirth,
			IsMinor:               n.IsMinor,
			Relationship:          n.Relationship,
			SharePercentage:       n.SharePercentage,
			AppointeeName:         &n.AppointeeName,
			AppointeeRelationship: &n.AppointeeRelationship,
			NomineeCustomerID:     nomineeCustomerID,
		}
	}

	if err := h.proposalRepo.SaveNominees(sctx.Ctx, req.ProposalID, nominees); err != nil {
		log.Error(sctx.Ctx, "Error saving nominees: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to save nominees",
			err,
		)
	}

	// Update section completion status
	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "nominees", true); err != nil {
		log.Error(sctx.Ctx, "Error updating nominee section: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to update nominee section")
	}

	return &resp.SectionUpdateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Nominee details updated successfully",
		},
		Status:    "UPDATED",
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdatePolicyDetails updates policy details
// [POL-API-010] Update Policy Details
func (h *ProposalHandler) UpdatePolicyDetails(sctx *serverRoute.Context, req PolicyDetailsRequest) (*resp.SectionUpdateResponse, error) {

	// Get proposal to verify it exists and is in DATA_ENTRY status
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.SectionUpdateResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Proposal not found",
				},
			}, nil
		}
		log.Error(sctx.Ctx, "Error fetching proposal: %v", err)
		return nil, err
	}

	// Validate status - must be in DATA_ENTRY
	if proposal.Status != domain.ProposalStatusDataEntry {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Proposal must be in DATA_ENTRY status",
			},
		}, nil
	}

	// Update proposal with new values
	proposalUpdates := map[string]interface{}{
		"sum_assured":               req.SumAssured,
		"policy_term":               req.PolicyTerm,
		"premium_ceasing_age":       req.PremiumCeasingAge,
		"premium_payment_frequency": domain.PremiumFrequency(req.PremiumFrequency),
	}
	if err := h.proposalRepo.UpdateProposalFields(sctx.Ctx, req.ProposalID, proposalUpdates); err != nil {
		log.Error(sctx.Ctx, "Error updating proposal fields: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to update proposal Fields")
	}
	// Ensure proposal_data_entry row exists
	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "policy", false); err != nil {
		log.Error(sctx.Ctx, "Error ensuring data entry row: %v", err)
		return nil, err
	}
	// Update data entry with policy details
	dataEntryUpdates := map[string]interface{}{
		"policy_taken_under":      domain.PolicyTakenUnder(req.PolicyTakenUnder),
		"age_proof_type":          domain.AgeProofType(req.AgeProofType),
		"subsequent_payment_mode": domain.SubsequentPaymentMode(req.SubsequentPaymentMode),
	}
	if err := h.proposalRepo.UpdateDataEntryFields(sctx.Ctx, req.ProposalID, dataEntryUpdates); err != nil {
		log.Error(sctx.Ctx, "Error updating data entry fields: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to update Data Entry Fields")
	}
	switch req.PolicyTakenUnder {

	case string(domain.PolicyTakenUnderHUF):

		if len(req.HUFMembers) == 0 {
			return &resp.SectionUpdateResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusBadRequest,
					Message:    "HUF members required for HUF policy",
				},
			}, nil
		}
		var members []repo.HUFMemberRepoInput

		for _, m := range req.HUFMembers {
			members = append(members, repo.HUFMemberRepoInput{
				IsFinancedHUF:                 m.IsFinancedHUF,
				KartaName:                     m.KartaName,
				HUFPan:                        m.HUFPan,
				LifeAssuredDifferentFromKarta: m.LifeAssuredDifferentFromKarta,
				KartaDifferentReason:          m.KartaDifferentReason,
				MemberName:                    m.MemberName,
				MemberRelationship:            m.MemberRelationship,
				MemberAge:                     m.MemberAge,
			})
		}
		if err := h.proposalRepo.InsertHUFMembers(sctx.Ctx, req.ProposalID, members); err != nil {
			log.Error(sctx.Ctx, "Error inserting HUF members: %v", err)
			return nil, err
		}

	case string(domain.PolicyTakenUnderMWPA):

		if req.MWPATrustee == nil {
			return &resp.SectionUpdateResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusBadRequest,
					Message:    "MWPA trustee details required",
				},
			}, nil
		}
		trustee := repo.MWPATrusteeRepoInput{
			TrustType:    req.MWPATrustee.TrustType,
			TrusteeName:  req.MWPATrustee.TrusteeName,
			TrusteeDOB:   req.MWPATrustee.TrusteeDOB,
			Relationship: req.MWPATrustee.Relationship,
			Address:      req.MWPATrustee.Address,
		}

		if err := h.proposalRepo.InsertMWPATrustee(sctx.Ctx, req.ProposalID, trustee); err != nil {
			log.Error(sctx.Ctx, "Error inserting MWPA trustee: %v", err)
			return nil, err
		}

	case string(domain.PolicyTakenUnderOther):
		// nothing required
	}

	// Update section completion status
	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "policy", true); err != nil {
		log.Error(sctx.Ctx, "Error updating policy section: %v", err)
		return nil, handleRepoError(err, "Proposal not found", "Failed to update policy section")
	}

	return &resp.SectionUpdateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Policy details updated successfully",
		},
		Status:    "UPDATED",
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateAgentDetails updates agent association
// [POL-API-011] Update Agent Details
func (h *ProposalHandler) UpdateAgentDetails(sctx *serverRoute.Context, req AgentDetailsRequest) (*resp.SectionUpdateResponse, error) {

	// Get proposal to verify it exists and is in DATA_ENTRY status
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.SectionUpdateResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Proposal not found",
				},
			}, nil
		}
		log.Error(sctx.Ctx, "Error fetching proposal: %v", err)
		return nil, err
	}

	// Validate status - must be in DATA_ENTRY
	if proposal.Status != domain.ProposalStatusDataEntry {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Proposal must be in DATA_ENTRY status",
			},
		}, nil
	}

	// Persist agent details
	// if err := h.proposalRepo.SaveAgentDetails(sctx.Ctx, req.ProposalID, req.AgentCode, req.AgentType); err != nil {
	// 	log.Error(sctx.Ctx, "Error saving agent details: %v", err)
	// 	return nil, err
	// }
	if err := h.proposalRepo.SaveAgentDetails(sctx.Ctx, req.ProposalID, req.AgentID,
		req.AgentSalutation, req.AgentName, req.AgentMobile, req.AgentEmail,
		req.AgentLandline, req.AgentStdCode, req.ReceivesCorrespondence, req.OpportunityID,
	); err != nil {

		log.Error(sctx.Ctx, "Error saving agent details: %v", err)
		return nil, err
	}

	// Update section completion status
	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "agent", true); err != nil {
		log.Error(sctx.Ctx, "Error updating agent section: %v", err)
		return nil, err
	}

	return &resp.SectionUpdateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Agent details updated successfully",
		},
		Status:    "UPDATED",
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateMedicalInfo updates medical questionnaire
// [POL-API-012] Update Medical Info
func (h *ProposalHandler) UpdateMedicalInfo(sctx *serverRoute.Context, req MedicalInfoRequest) (*resp.SectionUpdateResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.SectionUpdateResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Proposal not found",
				},
			}, nil
		}
		log.Error(sctx.Ctx, "Error fetching proposal: %v", err)
		return nil, err
	}

	if proposal.Status != domain.ProposalStatusDataEntry {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Proposal must be in DATA_ENTRY status",
			},
		}, nil
	}

	// Persist medical info
	for _, m := range req.MedicalInfo {
		var hospitalizationFrom, hospitalizationTo *time.Time
		if m.HospitalizationFrom != nil && *m.HospitalizationFrom != "" {
			parsedFrom, err := time.Parse("2006-01-02", *m.HospitalizationFrom)
			if err != nil {
				log.Error(sctx.Ctx, "Error parsing hospitalization from date: %v", err)
				return nil, err
			}
			hospitalizationFrom = &parsedFrom
		}
		if m.HospitalizationTo != nil && *m.HospitalizationTo != "" {
			parsedTo, err := time.Parse("2006-01-02", *m.HospitalizationTo)
			if err != nil {
				log.Error(sctx.Ctx, "Error parsing hospitalization to date: %v", err)
				return nil, err
			}
			hospitalizationTo = &parsedTo
		}

		medical := &domain.ProposalMedicalInfo{
			InsuredIndex:             m.InsuredIndex,
			IsSoundHealth:            m.IsSoundHealth,
			DiseaseTB:                m.DiseaseTB,
			DiseaseCancer:            m.DiseaseCancer,
			DiseaseParalysis:         m.DiseaseParalysis,
			DiseaseInsanity:          m.DiseaseInsanity,
			DiseaseHeartLungs:        m.DiseaseHeartLungs,
			DiseaseKidney:            m.DiseaseKidney,
			DiseaseBrain:             m.DiseaseBrain,
			DiseaseHIV:               m.DiseaseHIV,
			DiseaseHepatitisB:        m.DiseaseHepatitisB,
			DiseaseEpilepsy:          m.DiseaseEpilepsy,
			DiseaseNervous:           m.DiseaseNervous,
			DiseaseLiver:             m.DiseaseLiver,
			DiseaseLeprosy:           m.DiseaseLeprosy,
			DiseaseOther:             m.DiseaseOther,
			DiseasePhysicalDeformity: m.DiseasePhysicalDeformity,
			DiseaseDetails:           m.DiseaseDetails,
			FamilyHereditary:         m.FamilyHereditary,
			FamilyHereditaryDetails:  m.FamilyHereditaryDetails,
			MedicalLeave3yr:          m.MedicalLeave3yr,
			LeaveKind:                m.LeaveKind,
			LeavePeriod:              m.LeavePeriod,
			LeaveAilment:             m.LeaveAilment,
			HospitalName:             m.HospitalName,
			HospitalizationFrom:      hospitalizationFrom,
			HospitalizationTo:        hospitalizationTo,
			PhysicalDeformity:        m.PhysicalDeformity,
			DeformityType:            m.DeformityType,
			FamilyDoctorName:         m.FamilyDoctorName,
		}
		if err := h.proposalRepo.SaveMedicalInfo(sctx.Ctx, req.ProposalID, medical); err != nil {
			log.Error(sctx.Ctx, "Error saving medical info: %v", err)
			return nil, err
		}
	}

	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "medical", true); err != nil {
		log.Error(sctx.Ctx, "Error updating medical section: %v", err)
		return nil, err
	}

	return &resp.SectionUpdateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Medical information updated successfully",
		},
		Status:    "UPDATED",
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateDeclaration updates declaration status
func (h *ProposalHandler) UpdateDeclaration(sctx *serverRoute.Context, req DeclarationRequest) (*resp.SectionUpdateResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.SectionUpdateResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Proposal not found",
				},
			}, nil
		}
		log.Error(sctx.Ctx, "Error fetching proposal: %v", err)
		return nil, err
	}

	if proposal.Status != domain.ProposalStatusDataEntry {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Proposal must be in DATA_ENTRY status",
			},
		}, nil
	}

	if !req.IsAgreed {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "You must agree to the declaration",
			},
		}, nil
	}

	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "declaration", true); err != nil {
		log.Error(sctx.Ctx, "Error updating declaration section: %v", err)
		return nil, err
	}

	return &resp.SectionUpdateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Declaration updated successfully",
		},
		Status:    "UPDATED",
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateProposerDetails updates proposer details
// [POL-API-019] Update Proposer Details
func (h *ProposalHandler) UpdateProposerDetails(sctx *serverRoute.Context, req ProposerDetailsRequest) (*resp.SectionUpdateResponse, error) {

	// Get proposal details
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		return nil, handleRepoError(err, "Proposal not found", "Failed to fetch proposal")
	}

	// Validate status - must be in DATA_ENTRY
	if proposal.Status != domain.ProposalStatusDataEntry {
		return &resp.SectionUpdateResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Proposal must be in DATA_ENTRY status",
			},
		}, nil
	}
	dataEntryBy := req.DataEntryBy
	if dataEntryBy == 0 {
		dataEntryBy = 1
	}
	// Handle different proposer cases
	if req.IsSameAsInsured {
		if proposal.CustomerID == nil {
			return nil, badRequest("Insured customer ID not found")
		}

		proposer := &domain.ProposalProposer{
			ProposalID:            req.ProposalID,
			CustomerID:            *proposal.CustomerID,
			RelationshipToInsured: domain.ProposerRelationshipSelf,
		}

		err := h.proposalRepo.SaveProposerDetails(sctx.Ctx, req.ProposalID, proposer, dataEntryBy, true)
		if err != nil {
			log.Error(sctx.Ctx, "Error saving proposer: %v", err)
			return nil, serverError("Failed to save proposer details", err)
		}

	} else {

		// ================================
		// CASE 2: Proposer different
		// ================================

		if req.CustomerID == "" {
			return nil, badRequest("customer_id is required when proposer is different from insured")
		}
		customerID, err := strconv.ParseInt(req.CustomerID, 10, 64)
		if err != nil {
			return nil, badRequest("Invalid customer_id")
		}

		if req.Relationship == "" {
			return nil, badRequest("relationship is required when proposer is different from insured")
		}

		relationship := domain.ProposerRelationship(req.Relationship)

		validRelationships := []domain.ProposerRelationship{
			domain.ProposerRelationshipParent,
			domain.ProposerRelationshipSpouse,
			domain.ProposerRelationshipEmployer,
			domain.ProposerRelationshipHUFKarta,
			domain.ProposerRelationshipGuardian,
			domain.ProposerRelationshipOther,
		}

		valid := false
		for _, rel := range validRelationships {
			if relationship == rel {
				valid = true
				break
			}
		}

		if !valid {
			return nil, badRequest("Invalid relationship type")
		}

		// Validate specific cases based on relationship
		switch relationship {
		case domain.ProposerRelationshipParent:
			// Child policy validation
			// Check if product is children's policy (Bal Jeevan Bima / Gram Bal Jeevan Bima)
			// This would require product catalog lookup
			// For now, just log for future implementation
			log.Info(sctx.Ctx, "Parent relationship detected - child policy validation required")

		case domain.ProposerRelationshipSpouse:
			// Yugal Suraksha validation
			// Check if product is Yugal Suraksha (Joint Life)
			if proposal.SpouseCustomerID == nil {
				log.Warn(sctx.Ctx, "Spouse relationship but no spouse_customer_id set - may be regular policy with spouse as proposer")
			} else {
				log.Info(sctx.Ctx, "Spouse relationship with spouse_customer_id - Yugal Suraksha policy")
			}

		case domain.ProposerRelationshipEmployer:
			// Employer-sponsored policy validation
			// Check premium payer type
			if proposal.PremiumPayerType != domain.PremiumPayerEmployer {
				log.Warn(sctx.Ctx, "Employer relationship but premium payer is not EMPLOYER")
			}

		case domain.ProposerRelationshipHUFKarta:
			// HUF policy validation
			// Check if policy is taken under HUF
			// This would require checking proposal_data_entry.policy_taken_under
			log.Info(sctx.Ctx, "HUF Karta relationship - HUF policy validation required")

		case domain.ProposerRelationshipGuardian:
			// Guardian for minor validation
			// Check if insured is minor
			log.Info(sctx.Ctx, "Guardian relationship - minor insured validation required")

		case domain.ProposerRelationshipOther:
			// Other relationship - ensure relationship_details is provided
			if req.RelationshipDetails == "" {
				return nil, badRequest("Relationship details required when relationship is OTHER")
			}
		}

		// TODO: In a real implementation, we would:
		// 1. Create/update customer record for proposer in Customer Service
		// 2. Get customer_id from Customer Service
		// 3. Save proposer details with the customer_id

		// For now, create a mock proposer record
		// In production, customer_id should come from Customer Service
		// mockCustomerID := int64(999999) // This should be replaced with actual customer ID from Customer Service

		// Set relationship details if relationship is "OTHER"
		var relationshipDetails *string
		if relationship == domain.ProposerRelationshipOther {
			relationshipDetails = &req.RelationshipDetails
		}

		// if proposal.CustomerID == nil || *proposal.CustomerID == 0 {
		// 	return &resp.SectionUpdateResponse{
		// 		StatusCodeAndMessage: port.StatusCodeAndMessage{
		// 			StatusCode: http.StatusBadRequest,
		// 			Message:    "Customer ID must be set before updating proposer details",
		// 		},
		// 	}, nil
		// }
		proposer := &domain.ProposalProposer{
			ProposalID:            req.ProposalID,
			CustomerID:            customerID,
			RelationshipToInsured: relationship,
			RelationshipDetails:   relationshipDetails,
		}
		// Save proposer details
		// dataEntryBy := req.DataEntryBy
		// if dataEntryBy == 0 {
		// 	// Use a default value if not provided (in production, this should come from auth context)
		// 	dataEntryBy = 1
		// }

		if err := h.proposalRepo.SaveProposerDetails(sctx.Ctx, req.ProposalID, proposer, dataEntryBy, false); err != nil {
			log.Error(sctx.Ctx, "Error saving proposer details: %v", err)
			return nil, serverError("Failed to save proposer details", err)
		}
		// Update proposer customer_id in proposals table
		proposalUpdates := map[string]interface{}{
			"proposer_customer_id": customerID,
		}

		if err := h.proposalRepo.UpdateProposalFields(sctx.Ctx, req.ProposalID, proposalUpdates); err != nil {
			log.Error(sctx.Ctx, "Error updating proposal customer_id: %v", err)
			return nil, serverError("Failed to update proposal customer_id", err)
		}
	}
	// Update section completion status
	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "proposer", true); err != nil {
		log.Error(sctx.Ctx, "Error updating proposer section: %v", err)
		return nil, serverError("Failed to update proposer section", err)
	}

	return &resp.SectionUpdateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposer details updated successfully",
		},
		Status:    "UPDATED",
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// SubmitForQC submits proposal for quality review
// [POL-API-013] Submit for QC
func (h *ProposalHandler) SubmitForQC(sctx *serverRoute.Context, req SubmitForQCRequest) (*resp.SubmitForQCResponse, error) {

	// Get proposal details
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorNotFound,
				"Proposal not found",
				nil,
			)
		}
		log.Error(sctx.Ctx, "Failed to get proposal for QC submission", err)
		return nil, err
	}

	// Validate proposal status - must be in DATA_ENTRY to submit for QC
	if proposal.Status != domain.ProposalStatusDataEntry {
		return nil, badRequest("Proposal must be in DATA_ENTRY status to submit for QC")
	}

	// Check if all required sections are complete
	allComplete, err := h.proposalRepo.CheckAllSectionsComplete(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to check section completeness", err)
		return nil, err
	}
	if !allComplete {
		return nil, badRequest("All data entry sections must be completed before submitting for QC")
	}

	// Check if first premium is paid — premium is tracked separately in proposal_indexing
	// and must be verified before QC can proceed
	premiumPaid, err := h.proposalRepo.CheckPremiumPaid(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to check premium payment status", err)
		return nil, err
	}
	if !premiumPaid {
		return nil, badRequest("First premium must be paid before submitting for QC")
	}

	// Update proposal_data_entry status BEFORE proposal status update
	// changedBy := proposal.CreatedBy // Use the proposal creator as the change initiator
	changedBy := req.DataEntryBy
	err = h.proposalRepo.UpdateDataEntryStatus(sctx.Ctx, req.ProposalID,
		"SUBMITTED_TO_QC", changedBy,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed updating data entry status", err)
		return nil, err
	}

	// If workflow already exists → signal it
	if proposal.WorkflowID != nil {

		err = h.temporalClient.SignalWorkflow(
			sctx.Ctx,
			*proposal.WorkflowID,
			"",
			workflows.SignalSubmitForQC,
			workflows.SubmitForQCSignal{
				DataEntryID: changedBy,
			},
		)

		if err != nil {
			log.Error(sctx.Ctx, "Failed to signal workflow", err)
			return nil, err
		}

		return &resp.SubmitForQCResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusOK,
				Message:    "Proposal resubmitted for QC successfully",
			},
			Status:     string(domain.ProposalStatusQCPending),
			WorkflowID: *proposal.WorkflowID,
		}, nil
	}
	// Fetch insured details for workflow
	insured, err := h.proposalRepo.GetInsuredByProposalID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get insured details for workflow", err)
		return nil, err
	}

	// Fetch data entry details for age proof type
	dataEntry, err := h.proposalRepo.GetDataEntryByProposalID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get data entry details for workflow", err)
		return nil, err
	}

	// ageAtEntry := domain.CalculateAgeTime(insured.DateOfBirth)
	indexing, err := h.proposalRepo.GetProposalIndexingByProposalID(
		sctx.Ctx,
		req.ProposalID,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to fetch proposal indexing", err)
		return nil, err
	}

	ageAtEntry := domain.CalculateANB(
		insured.DateOfBirth,
		indexing.ProposalDate,
	)

	gender := insured.Gender
	insuredState := ""
	if insured.State != nil {
		insuredState = *insured.State
	}
	ageProofType := ""
	if dataEntry.AgeProofType != nil {
		ageProofType = string(*dataEntry.AgeProofType)
	}
	// product, err := h.productRepo.GetProductByCode(sctx.Ctx, proposal.ProductCode)
	// if err != nil {
	// 	log.Error(sctx.Ctx, "Failed to fetch product catalog", err)
	// 	return nil, err
	// }

	// if product.ProductCategory == "" {
	// 	return nil, fmt.Errorf("product category not configured")
	// }
	// changedBy := proposal.CreatedBy // Use the proposal creator as the change initiator
	if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID, domain.ProposalStatusQCPending, "Submitted for QC", changedBy); err != nil {
		log.Error(sctx.Ctx, "Failed to update proposal status to QC_PENDING", err)
		return nil, err
	}

	customerID := ""
	if proposal.CustomerID != nil {
		customerID = fmt.Sprintf("%d", *proposal.CustomerID)
	}
	workflowInput := workflows.PolicyIssuanceInput{
		ProposalID:     proposal.ProposalID,
		ProposalNumber: proposal.ProposalNumber,
		CustomerID:     customerID,
		ProductCode:    proposal.ProductCode,
		// ProductCategory:         string(product.ProductCategory),
		PolicyType:              proposal.PolicyType,
		SumAssured:              proposal.SumAssured,
		PolicyTerm:              proposal.PolicyTerm,
		PremiumPaymentFrequency: proposal.PremiumPaymentFrequency,
		AgeAtEntry:              ageAtEntry,
		Gender:                  gender,
		AgeProofType:            ageProofType,
		InsuredState:            insuredState,
		ProposalDate:            indexing.ProposalDate,
	}
	log.Info(sctx.Ctx, "Policy Issuance Workflow Input", workflowInput)
	we, err := h.temporalClient.ExecuteWorkflow(
		sctx.Ctx,
		client.StartWorkflowOptions{
			ID:        "policy-issuance-" + proposal.ProposalNumber,
			TaskQueue: "policy-issue-queue",
		},
		workflows.PolicyIssuanceWorkflow,
		workflowInput,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to start Policy Issuance Workflow", err)
		return nil, err
	}

	// Persist Workflow ID
	if err := h.proposalRepo.UpdateWorkflowID(sctx.Ctx, req.ProposalID, we.GetID()); err != nil {
		log.Error(sctx.Ctx, "Failed to update workflow ID", err)
		// We don't return error here as workflow is already started
	}

	return &resp.SubmitForQCResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposal submitted for QC successfully",
		},
		Status:     string(domain.ProposalStatusQCPending),
		WorkflowID: we.GetID(),
		RunID:      we.GetRunID(),
	}, nil
}

// StartDataEntry starts data entry on an indexed proposal
// [POL-API-018] Start Data Entry
// State transition: INDEXED → DATA_ENTRY (BR-POL-015)
func (h *ProposalHandler) StartDataEntry(sctx *serverRoute.Context, req StartDataEntryRequest) (*resp.StartDataEntryResponse, error) {

	// Get proposal details
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.StartDataEntryResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Proposal not found",
				},
			}, nil
		}
		log.Error(sctx.Ctx, "Failed to get proposal for data entry start", err)
		return nil, err
	}

	// Validate proposal is in INDEXED status
	// if proposal.Status != domain.ProposalStatusIndexed {
	if proposal.Status != domain.ProposalStatusIndexed &&
		proposal.Status != domain.ProposalStatusQCReturned {
		return &resp.StartDataEntryResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("Proposal must be in INDEXED or QC_RETURNED status to start data entry. Current: %s", proposal.Status),
			},
		}, nil
	}

	// Update proposal status to DATA_ENTRY
	previousStatus := proposal.Status
	if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID,
		domain.ProposalStatusDataEntry, "Data entry started: "+req.Comments, req.AssignedTo); err != nil {
		log.Error(sctx.Ctx, "Failed to update proposal status to DATA_ENTRY", err)
		return &resp.StartDataEntryResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			},
		}, nil
	}
	log.Info(sctx.Ctx, "UpdateProposalStatus completed")

	log.Info(sctx.Ctx, "Calling RecordDataEntryAssignment")
	// Record assignment in proposal_data_entry table
	if err := h.proposalRepo.RecordDataEntryAssignment(sctx.Ctx, req.ProposalID, req.AssignedTo, req.Comments); err != nil {
		log.Error(sctx.Ctx, "Failed to record data entry assignment", err)
		// Don't fail the whole request if assignment recording fails
	}

	return &resp.StartDataEntryResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Data entry started successfully",
		},
		ProposalID:     proposal.ProposalID,
		ProposalNumber: proposal.ProposalNumber,
		PreviousStatus: string(previousStatus),
		NewStatus:      string(domain.ProposalStatusDataEntry),
		AssignedTo:     req.AssignedTo,
		AssignedAt:     time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// GetProposalSummary retrieves proposal summary for review
// [POL-API-016] Get Proposal Summary
func (h *ProposalHandler) GetProposalSummary(sctx *serverRoute.Context, req ProposalIDUri) (*resp.ProposalSummaryResponse, error) {
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.ProposalSummaryResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Proposal not found",
				},
			}, nil
		}
		log.Error(sctx.Ctx, "Error fetching proposal summary: %v", err)
		return nil, err
	}

	return &resp.ProposalSummaryResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposal summary retrieved successfully",
		},
		ProposalID:     proposal.ProposalID,
		ProposalNumber: proposal.ProposalNumber,
		Status:         string(proposal.Status),
	}, nil
}

// GetProposalQueue retrieves proposals in queue
// [WF-POL-003] Get Proposal Queue
func (h *ProposalHandler) GetProposalQueue(sctx *serverRoute.Context, req GetProposalQueueRequest) (*resp.ProposalQueueResponse, error) {
	// Set default pagination values
	req.MetadataRequest.SetDefaults()

	// Get proposals by status
	proposals, total, err := h.proposalRepo.GetProposalsByStatus(sctx.Ctx, req.Status, int(req.Skip), int(req.Limit))
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching proposal queue: %v", err)
		return nil, err
	}

	// Map to response
	summaries := make([]resp.ProposalSummary, len(proposals))
	for i, p := range proposals {
		summaries[i] = resp.ProposalSummary{
			ProposalID:     p.ProposalID,
			ProposalNumber: p.ProposalNumber,
			Status:         string(p.Status),
			CustomerName:   "", // TODO: Fetch from customer service
			ProductCode:    p.ProductCode,
			SumAssured:     p.SumAssured,
			CreatedAt:      p.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &resp.ProposalQueueResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposal queue retrieved successfully",
		},
		MetaDataResponse: port.NewMetaDataResponse(total, int(req.Skip), int(req.Limit)),
		Proposals:        summaries,
	}, nil
}

func (h *ProposalHandler) GetProposalSection(sctx *serverRoute.Context, req GetProposalSectionRequest,
) (*resp.ProposalSectionResponse, error) {

	if req.Section == "" {
		return nil, badRequest("section is required")
	}

	if req.Section == "quote" {

		if req.QuoteRefNumber == "" {
			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorBadRequest,
				"quote_ref_number is required",
				nil,
			)
		}

		quote, err := h.quoteRepo.GetQuoteByRefNumber(sctx.Ctx, req.QuoteRefNumber)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching quote: %v", err)
			return nil, serverError("Failed to fetch quote", err)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Quote fetched successfully",
			},
			Section: req.Section,
			Data:    quote,
		}, nil
	}

	if req.ProposalNumber == "" {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"proposal_number is required",
			nil,
		)
	}

	proposal, err := h.proposalRepo.GetProposalByNumber(sctx.Ctx, req.ProposalNumber)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching proposal: %v", err)

		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to fetch proposal",
			err,
		)
	}

	if proposal == nil {
		return nil, notFound("Proposal not found")
	}

	proposalID := proposal.ProposalID
	switch req.Section {

	case "proposal":

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Proposal fetched successfully",
			},
			Section: req.Section,
			Data:    resp.FetchProposalResponse(proposal),
		}, nil

	case "indexing":

		data, err := h.proposalRepo.GetIndexingSection(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching indexing: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch indexing",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Indexing fetched successfully",
			},
			Section: req.Section,
			Data:    data,
		}, nil

	case "first_premium":

		data, err := h.proposalRepo.GetFirstPremiumSection(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching first premium: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch first premium",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "First premium fetched successfully",
			},
			Section: req.Section,
			Data:    data,
		}, nil

	case "insured":

		data, err := h.proposalRepo.GetInsuredByProposalID(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching insured: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch insured",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Insured fetched successfully",
			},
			Section: req.Section,
			Data:    data,
		}, nil

	case "nominee":

		data, err := h.proposalRepo.GetNomineesByProposalID(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching nominees: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch nominees",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Nominees fetched successfully",
			},
			Section: req.Section,
			Data:    resp.FetchNomineeResponse(data),
		}, nil

	case "agent":

		data, err := h.proposalRepo.GetAgentByProposalID(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching agent: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch agent",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Agent fetched successfully",
			},
			Section: req.Section,
			Data:    data,
		}, nil

	case "data_entry":

		data, err := h.proposalRepo.GetDataEntryByProposalID(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching data entry: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch data entry",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Data entry fetched successfully",
			},
			Section: req.Section,
			Data:    resp.FetchDataEntryResponse(data),
		}, nil

	case "medical_info":

		data, err := h.proposalRepo.GetMedicalInfoByProposalID(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching medical info: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch medical info",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Medical info fetched successfully",
			},
			Section: req.Section,
			Data:    data,
		}, nil

	case "proposer":

		data, err := h.proposalRepo.GetProposerByProposalID(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching proposer: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch proposer",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Proposer fetched successfully",
			},
			Section: req.Section,
			Data:    resp.FetchProposerResponse(data),
		}, nil

	case "qc_review":

		data, err := h.proposalRepo.GetQCReviewByProposalID(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching qc review: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch qc review",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "QC review fetched successfully",
			},
			Section: req.Section,
			Data:    resp.FetchQCReviewResponse(data),
		}, nil
	case "issuance":

		data, err := h.proposalRepo.GetIssuanceByProposalID(sctx.Ctx, proposalID)
		if err != nil {
			log.Error(sctx.Ctx, "Error fetching issuance: %v", err)

			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorServerError,
				"Failed to fetch issuance",
				err,
			)
		}

		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "Issuance fetched successfully",
			},
			Section: req.Section,
			Data:    data,
		}, nil

	case "all":

		indexing, _ := h.proposalRepo.GetIndexingSection(sctx.Ctx, proposalID)
		firstPremium, _ := h.proposalRepo.GetFirstPremiumSection(sctx.Ctx, proposalID)
		insured, _ := h.proposalRepo.GetInsuredByProposalID(sctx.Ctx, proposalID)
		nominees, _ := h.proposalRepo.GetNomineesByProposalID(sctx.Ctx, proposalID)
		agent, _ := h.proposalRepo.GetAgentByProposalID(sctx.Ctx, proposalID)
		dataEntry, _ := h.proposalRepo.GetDataEntryByProposalID(sctx.Ctx, proposalID)
		medical, _ := h.proposalRepo.GetMedicalInfoByProposalID(sctx.Ctx, proposalID)
		proposer, _ := h.proposalRepo.GetProposerByProposalID(sctx.Ctx, proposalID)
		qc, _ := h.proposalRepo.GetQCReviewByProposalID(sctx.Ctx, proposalID)
		issuance, _ := h.proposalRepo.GetIssuanceByProposalID(sctx.Ctx, proposalID)

		data := map[string]interface{}{
			"proposal":      resp.FetchProposalResponse(proposal),
			"indexing":      indexing,
			"first_premium": firstPremium,
			"insured":       resp.FetchInsuredResponse(insured),
			"nominee":       resp.FetchNomineeResponse(nominees),
			"agent":         agent,
			"data_entry":    resp.FetchDataEntryResponse(dataEntry),
			"medical_info":  resp.FetchMedicalInfoResponse(medical),
			"proposer":      resp.FetchProposerResponse(proposer),
			"qc_review":     resp.FetchQCReviewResponse(qc),
			"issuance":      issuance,
		}
		return &resp.ProposalSectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Message:    "All sections fetched successfully",
			},
			Section: req.Section,
			Data:    data,
		}, nil

	default:

		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"Invalid section",
			nil,
		)
	}
}

// GetProposalAuditLogs retrieves audit logs for a specific proposal
// [FR-POL-033] Proposal Summary View (audit trail component)
func (h *ProposalHandler) GetProposalAuditLogs(sctx *serverRoute.Context, req struct {
	ProposalID int64 `param:"proposal_id"`
}) (*resp.AuditLogsResponse, error) {

	domainLogs, err := h.proposalRepo.GetAuditLogsByProposal(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get proposal audit logs: %v", err)
		return nil, err
	}

	// Convert domain logs to response logs
	auditLogs := make([]resp.AuditLog, 0, len(domainLogs))
	for _, domainLog := range domainLogs {
		auditLogs = append(auditLogs, resp.AuditLog{
			AuditID:      domainLog.AuditID,
			ProposalID:   domainLog.ProposalID,
			EntityType:   domainLog.EntityType,
			EntityID:     domainLog.EntityID,
			FieldName:    domainLog.FieldName,
			OldValue:     domainLog.OldValue,
			NewValue:     domainLog.NewValue,
			ChangeType:   string(domainLog.ChangeType),
			ChangedBy:    domainLog.ChangedBy,
			ChangedAt:    domainLog.ChangedAt.Format(time.RFC3339),
			ChangeReason: domainLog.ChangeReason,
			Metadata:     domainLog.Metadata,
		})
	}

	return &resp.AuditLogsResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposal audit logs retrieved successfully",
		},
		AuditLogs: auditLogs,
	}, nil
}

// GetEntityAuditLogs retrieves audit logs for a specific entity within a proposal
func (h *ProposalHandler) GetEntityAuditLogs(sctx *serverRoute.Context, req struct {
	ProposalID int64  `param:"proposal_id"`
	EntityType string `param:"entity_type"`
	EntityID   int64  `param:"entity_id"`
}) (*resp.AuditLogsResponse, error) {

	domainLogs, err := h.proposalRepo.GetAuditLogsByEntity(sctx.Ctx, req.EntityType, req.EntityID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get entity audit logs: %v", err)
		return nil, err
	}

	// Convert domain logs to response logs
	auditLogs := make([]resp.AuditLog, 0, len(domainLogs))
	for _, domainLog := range domainLogs {
		auditLogs = append(auditLogs, resp.AuditLog{
			AuditID:      domainLog.AuditID,
			ProposalID:   domainLog.ProposalID,
			EntityType:   domainLog.EntityType,
			EntityID:     domainLog.EntityID,
			FieldName:    domainLog.FieldName,
			OldValue:     domainLog.OldValue,
			NewValue:     domainLog.NewValue,
			ChangeType:   string(domainLog.ChangeType),
			ChangedBy:    domainLog.ChangedBy,
			ChangedAt:    domainLog.ChangedAt.Format(time.RFC3339),
			ChangeReason: domainLog.ChangeReason,
			Metadata:     domainLog.Metadata,
		})
	}

	return &resp.AuditLogsResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Entity audit logs retrieved successfully",
		},
		AuditLogs: auditLogs,
	}, nil
}
