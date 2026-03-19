package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"
	"policy-issue-service/workflows"

	log "gitlab.cept.gov.in/it-2.0-common/api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"go.temporal.io/sdk/client"
)

// AadhaarHandler handles Aadhaar-based proposal HTTP endpoints
type AadhaarHandler struct {
	*serverHandler.Base

	temporalClient client.Client
	aadhaarRepo    *repo.AadhaarRepository
	productRepo    *repo.ProductRepository
}

// NewAadhaarHandler creates a new AadhaarHandler instance with dependencies
func NewAadhaarHandler(temporalClient client.Client, aadhaarRepo *repo.AadhaarRepository, productRepo *repo.ProductRepository) *AadhaarHandler {
	base := serverHandler.New("Aadhaar").SetPrefix("/v1").AddPrefix("")
	return &AadhaarHandler{
		Base:           base,
		temporalClient: temporalClient,
		aadhaarRepo:    aadhaarRepo,
		productRepo:    productRepo,
	}
}

// Routes returns the routes for the AadhaarHandler
func (h *AadhaarHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/proposals/aadhaar/initiate", h.InitiateAadhaarAuth).Name("Initiate Aadhaar Auth"),
		serverRoute.POST("/proposals/aadhaar/verify-otp", h.VerifyAadhaarOTP).Name("Verify Aadhaar OTP"),
		serverRoute.POST("/proposals/aadhaar/submit", h.SubmitAadhaarProposal).Name("Submit Aadhaar Proposal"),
	}
}

// InitiateAadhaarAuth initiates Aadhaar authentication
// [POL-API-013] Initiate Aadhaar Auth
func (h *AadhaarHandler) InitiateAadhaarAuth(sctx *serverRoute.Context, req AadhaarInitiateRequest) (*resp.AadhaarInitiateResponse, error) {

	sessionID, err := generateSecureID(16)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to generate session ID", "error", err)
		return nil, fmt.Errorf("internal error")
	}

	transactionID, err := generateSecureID(12)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to generate transaction ID", "error", err)
		return nil, fmt.Errorf("internal error")
	}

	session := domain.AadhaarSession{
		SessionID:     sessionID,
		TransactionID: transactionID,
		AadhaarNumber: req.AadhaarNumber,
		OTPVerified:   false,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}

	if err := h.aadhaarRepo.StoreSession(sctx.Ctx, session); err != nil {
		log.Error(sctx.Ctx, "Failed to store Aadhaar session", "error", err)
		return nil, fmt.Errorf("failed to create Aadhaar session: %w", err)
	}

	return &resp.AadhaarInitiateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Aadhaar authentication initiated",
		},
		TransactionID: transactionID,
		SessionID:     sessionID,
	}, nil
}

// VerifyAadhaarOTP verifies Aadhaar OTP
// [POL-API-014] Verify Aadhaar OTP
func (h *AadhaarHandler) VerifyAadhaarOTP(sctx *serverRoute.Context, req AadhaarVerifyOTPRequest) (*resp.AadhaarVerifyOTPResponse, error) {

	session, err := h.aadhaarRepo.GetSessionByTransactionID(sctx.Ctx, req.TransactionID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to retrieve Aadhaar session", "transactionID", req.TransactionID, "error", err)
		return nil, fmt.Errorf("invalid transaction ID or session expired")
	}

	if session.OTPVerified {
		return &resp.AadhaarVerifyOTPResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusOK,
				Message:    "OTP already verified",
			},
			SessionID: session.SessionID,
			Status:    "already_verified",
		}, nil
	}

	// Mock response from UIDAI
	userData := map[string]interface{}{
		"name":          "JOHN DOE",
		"dob":           "1980-01-01",
		"gender":        "M",
		"address":       "MUMBAI, MAHARASHTRA",
		"photo":         "base64encodedimage",
		"mobile_number": "9876543210",
	}

	session.UserData = userData
	session.OTPVerified = true

	if err := h.aadhaarRepo.UpdateSession(sctx.Ctx, *session); err != nil {
		log.Error(sctx.Ctx, "Failed to update Aadhaar session", "sessionID", session.SessionID, "error", err)
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &resp.AadhaarVerifyOTPResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Aadhaar OTP verified successfully",
		},
		SessionID: session.SessionID,
		Status:    "success",
	}, nil
}

// SubmitAadhaarProposal submits Aadhaar-based proposal
// [POL-API-015] Submit Aadhaar Proposal
func (h *AadhaarHandler) SubmitAadhaarProposal(sctx *serverRoute.Context, req AadhaarSubmitRequest) (*resp.AadhaarSubmitResponse, error) {

	session, err := h.aadhaarRepo.GetSessionByID(sctx.Ctx, req.SessionID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to retrieve Aadhaar session", "sessionID", req.SessionID, "error", err)
		return nil, fmt.Errorf("invalid session ID or session expired")
	}

	if !session.OTPVerified {
		return nil, fmt.Errorf("OTP not verified for session")
	}

	name, _ := session.UserData["name"].(string)
	dob, _ := session.UserData["dob"].(string)
	gender, _ := session.UserData["gender"].(string)
	address, _ := session.UserData["address"].(string)

	// Basic state extraction for mock
	stateCode := "DL"
	if strings.Contains(address, "MAHARASHTRA") {
		stateCode = "MH"
	}

	// [WF-PI-002] Instant Issuance Eligibility Check
	// Criteria: Aadhaar verified + Age ≤ 50 + SA < 20L + Non-medical + Payment completed
	const (
		maxInstantAge = 50
		maxInstantSA  = 2000000.0 // ₹20,00,000
	)

	// Compute age from DOB using domain helper
	eligibleForInstant := true
	var ineligibilityReasons []string

	age := domain.CalculateAge(dob)
	if age > maxInstantAge {
		eligibleForInstant = false
		ineligibilityReasons = append(ineligibilityReasons,
			fmt.Sprintf("Age %d exceeds maximum %d for instant issuance", age, maxInstantAge))
	}

	// Check SA limit
	if req.SumAssured >= maxInstantSA {
		eligibleForInstant = false
		ineligibilityReasons = append(ineligibilityReasons,
			fmt.Sprintf("Sum assured ₹%.2f exceeds ₹%.2f limit for instant issuance", req.SumAssured, maxInstantSA))
	}

	// Check medical requirement from product catalog
	product, productErr := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
	if productErr == nil && product != nil {
		if product.IsMedicalRequired(req.SumAssured) {
			eligibleForInstant = false
			ineligibilityReasons = append(ineligibilityReasons,
				"Medical examination required for this SA — instant issuance not applicable")
		}
	}

	// Check payment
	if !req.PremiumPaid {
		eligibleForInstant = false
		ineligibilityReasons = append(ineligibilityReasons,
			"Premium payment not completed")
	}

	workflowInput := workflows.InstantIssuanceInput{
		ProposalNumber:      fmt.Sprintf("aadhaar_%s", req.SessionID),
		CustomerID:      "1001", // Mock customer ID
		ProductCode:     req.ProductCode,
		PolicyType:      req.PolicyType,
		SumAssured:      req.SumAssured,
		PolicyTerm:      req.PolicyTerm,
		Frequency:       req.Frequency,
		Channel:         req.Channel,
		AadhaarVerified: true,
		PremiumPaid:     req.PremiumPaid,
		PaymentRef:      req.PaymentRef,
		CustomerName:    name,
		CustomerDOB:     dob,
		Gender:          gender,
		Address:         address,
		MobileNumber:    req.MobileNumber,
		Email:           req.Email,
		StateCode:       stateCode,
	}

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "policy-issue-tq",
	}

	if eligibleForInstant {
		// Start WF-PI-002 (Instant Issuance)
		workflowOptions.ID = fmt.Sprintf("ii-%s", req.SessionID)
		we, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, workflows.InstantIssuanceWorkflow, workflowInput)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to start instant issuance workflow", "sessionID", req.SessionID, "error", err)
			return nil, fmt.Errorf("failed to start workflow: %w", err)
		}

		return &resp.AadhaarSubmitResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusAccepted,
				Message:    "Proposal accepted for instant issuance",
			},
			ProposalID:   0,
			WorkflowID:   we.GetID(),
			RunID:        we.GetRunID(),
			IssuanceType: "INSTANT",
		}, nil
	}

	// Not eligible for instant issuance — fall back to WF-PI-001 (Standard)
	// [ERR-POL-048] Not eligible for instant issuance
	log.Info(sctx.Ctx, "Proposal not eligible for instant issuance, falling back to standard workflow",
		"sessionID", req.SessionID, "reasons", ineligibilityReasons)

	standardInput := workflows.PolicyIssuanceInput{
		ProposalNumber:          workflowInput.ProposalNumber,
		CustomerID:              workflowInput.CustomerID,
		ProductCode:             workflowInput.ProductCode,
		PolicyType:              workflowInput.PolicyType,
		SumAssured:              workflowInput.SumAssured,
		PolicyTerm:              workflowInput.PolicyTerm,
		PremiumPaymentFrequency: workflowInput.Frequency,
	}

	workflowOptions.ID = fmt.Sprintf("pi-%s", req.SessionID)
	we, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, workflows.PolicyIssuanceWorkflow, standardInput)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to start standard issuance workflow", "sessionID", req.SessionID, "error", err)
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	return &resp.AadhaarSubmitResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusAccepted,
			Message:    "ERR-POL-048: Not eligible for instant issuance. Proposal accepted for standard processing.",
		},
		ProposalID:   0,
		WorkflowID:   we.GetID(),
		RunID:        we.GetRunID(),
		IssuanceType: "STANDARD",
	}, nil
}

func generateSecureID(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
