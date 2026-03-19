package handler

import (
	"fmt"
	"net/http"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// DocumentHandler handles document management HTTP endpoints
// Phase 8: [DOC-POL-001] to [DOC-POL-007]
type DocumentHandler struct {
	*serverHandler.Base
	documentRepo *repo.DocumentRepository
	proposalRepo *repo.ProposalRepository
	productRepo  *repo.ProductRepository
}

// NewDocumentHandler creates a new DocumentHandler instance
func NewDocumentHandler(documentRepo *repo.DocumentRepository, proposalRepo *repo.ProposalRepository, productRepo *repo.ProductRepository) *DocumentHandler {
	base := serverHandler.New("Document").SetPrefix("/v1").AddPrefix("")
	return &DocumentHandler{
		Base:         base,
		documentRepo: documentRepo,
		proposalRepo: proposalRepo,
		productRepo:  productRepo,
	}
}

// Routes returns the routes for the DocumentHandler
func (h *DocumentHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		// Document Checklist APIs
		serverRoute.GET("/proposals/:proposal_id/required-documents", h.GetRequiredDocuments).Name("Get Required Documents"),
		serverRoute.POST("/proposals/:proposal_id/documents", h.UploadDocument).Name("Upload Document"),
		serverRoute.GET("/proposals/:proposal_id/documents/:document_id", h.DownloadDocument).Name("Download Document"),
		serverRoute.DELETE("/proposals/:proposal_id/documents/:document_id", h.RemoveDocument).Name("Remove Document"),
		// Missing Documents APIs
		serverRoute.GET("/proposals/:proposal_id/missing-documents", h.GetMissingDocuments).Name("Get Missing Documents"),
		serverRoute.POST("/proposals/:proposal_id/missing-documents", h.AddMissingDocument).Name("Add Missing Document"),
		serverRoute.PUT("/proposals/:proposal_id/missing-documents/:missing_doc_id/resolve", h.ResolveMissingDocument).Name("Resolve Missing Document"),
	}
}

// GetRequiredDocuments returns the dynamic document checklist for a proposal
// [DOC-POL-001] Dynamic document checklist
// [FR-POL-029] Document Management
// Logic: Required documents depend on product type, SA, age, and medical requirement
func (h *DocumentHandler) GetRequiredDocuments(sctx *serverRoute.Context, req ProposalIDUri) (*resp.DocumentChecklistResponse, error) {
	// Step 1: Get the proposal to determine requirements
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-001] Error fetching proposal %d: %v", req.ProposalID, err)
		return nil, err
	}

	// Step 2: Determine medical requirement from product catalog
	// [BR-POL-013] Medical Requirement Determination
	// NOTE: proposals table does NOT have is_medical_required column.
	// It lives in proposal_medical (E-007D). We compute it at runtime
	// from product.MedicalSAThreshold and proposal.SumAssured.
	isMedicalRequired := false
	var policyType string
	product, err := h.productRepo.GetProductByCode(sctx.Ctx, proposal.ProductCode)
	if err != nil {
		log.Warn(sctx.Ctx, "[DOC-POL-001] Could not fetch product %s for medical check: %v", proposal.ProductCode, err)
		// Non-fatal: default to false (no medical docs required)
	} else {
		isMedicalRequired = product.IsMedicalRequired(proposal.SumAssured)
		policyType = string(product.ProductType)
	}

	// Step 3: Get existing uploaded documents for this proposal
	existingDocs, err := h.documentRepo.GetDocumentsByProposalID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-001] Error fetching existing documents for proposal %d: %v", req.ProposalID, err)
		return nil, err
	}

	// Build a map of uploaded document types for quick lookup
	uploadedMap := make(map[domain.DocumentType]*domain.ProposalDocumentRef)
	for i := range existingDocs {
		uploadedMap[existingDocs[i].DocumentType] = &existingDocs[i]
	}

	// Step 4: Define required documents based on proposal characteristics
	params := checklistParams{
		SumAssured:              proposal.SumAssured,
		IsMedicalRequired:       isMedicalRequired,
		IsProposerSameAsInsured: proposal.IsProposerSameAsInsured,
		PolicyType:              policyType,
	}
	requiredDocs := buildRequiredDocuments(params, uploadedMap)

	// Step 5: Calculate completion percentage
	uploaded := 0
	mandatoryCount := 0
	mandatoryUploaded := 0
	for _, doc := range requiredDocs {
		if doc.IsUploaded {
			uploaded++
		}
		if doc.IsMandatory {
			mandatoryCount++
			if doc.IsUploaded {
				mandatoryUploaded++
			}
		}
	}

	completionPct := 0
	if mandatoryCount > 0 {
		completionPct = (mandatoryUploaded * 100) / mandatoryCount
	}

	return &resp.DocumentChecklistResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Document checklist retrieved successfully",
		},
		ProposalID:           req.ProposalID,
		RequiredDocuments:    requiredDocs,
		CompletionPercentage: completionPct,
	}, nil
}

// checklistParams holds all the context needed to determine required documents.
// Extracted from proposal and product data; avoids passing the full structs into
// the pure-logic buildRequiredDocuments function.
type checklistParams struct {
	SumAssured              float64
	IsMedicalRequired       bool
	IsProposerSameAsInsured bool
	PolicyType              string // PLI or RPLI
}

// buildRequiredDocuments determines required documents based on proposal characteristics
// [FR-POL-029] Document Management
// Conditional rules:
//   - PAN: mandatory when SA ≥ 50,000 (Indian insurance KYC/tax rules)
//   - Medical report: computed from product.MedicalSAThreshold
//   - Income proof: SA > 5,00,000
//   - Employment proof: SA > 10,00,000
//   - Proposer docs: when proposer ≠ insured, proposer identity/address proof required
func buildRequiredDocuments(params checklistParams, uploadedMap map[domain.DocumentType]*domain.ProposalDocumentRef) []resp.DocumentChecklistItem {
	var checklist []resp.DocumentChecklistItem

	// Helper to add a checklist item with upload-status lookup.
	// subType is an optional qualifier for distinguishing multiple entries
	// of the same document_type_enum value (e.g., "PROPOSER_PHOTO_ID").
	addItem := func(docType domain.DocumentType, subType string, isMandatory bool, reason string) {
		item := resp.DocumentChecklistItem{
			DocumentType:   string(docType),
			SubType:        subType,
			IsMandatory:    isMandatory,
			ReasonRequired: reason,
		}
		// For items with a SubType, match upload by checking comments for sub_type metadata.
		// For standard items, match by document type only.
		if subType != "" {
			// Look for an uploaded document whose comments contain the sub_type identifier
			for _, doc := range uploadedMap {
				if doc.DocumentType == docType && doc.Comments != nil && *doc.Comments == "sub_type:"+subType {
					item.IsUploaded = true
					docID := doc.DocumentID
					item.DocumentID = &docID
					uploadedAt := doc.UploadedAt
					item.UploadDate = &uploadedAt
					break
				}
			}
		} else if doc, exists := uploadedMap[docType]; exists {
			item.IsUploaded = true
			docID := doc.DocumentID
			item.DocumentID = &docID
			uploadedAt := doc.UploadedAt
			item.UploadDate = &uploadedAt
		}
		checklist = append(checklist, item)
	}

	// ── Always-required documents ──
	addItem(domain.DocumentTypeProposalForm, "", true, "Signed proposal form is mandatory for all new business")
	addItem(domain.DocumentTypeDOBProof, "", true, "Age proof is mandatory for life insurance issuance")
	addItem(domain.DocumentTypePhoto, "", true, "Passport-size photo is mandatory for policy bond")
	addItem(domain.DocumentTypeAddressProof, "", true, "Address proof is mandatory for policy issuance")
	addItem(domain.DocumentTypePaymentCopy, "", true, "First premium payment receipt/copy is mandatory")
	addItem(domain.DocumentTypeHealthDeclaration, "", true, "Health declaration is mandatory for underwriting")

	// ── Photo ID / PAN requirement ──
	// PAN is mandatory for SA ≥ 50,000 per IRDAI KYC norms and Income Tax Act Sec 114B.
	// Since the DDL enum uses PHOTO_ID (which covers PAN card), we differentiate
	// the reason text based on the PAN requirement threshold.
	panRequired := params.SumAssured >= 50000
	if panRequired {
		addItem(domain.DocumentTypePhotoID, "", true, "PAN card / Photo ID mandatory for SA ≥ ₹50,000 (IRDAI KYC / IT Act Sec 114B)")
	} else {
		addItem(domain.DocumentTypePhotoID, "", true, "Photo ID is mandatory for KYC compliance")
	}

	// ── Medical report: conditional on product SA threshold ──
	// [BR-POL-013] Medical Requirement Determination
	if params.IsMedicalRequired {
		addItem(domain.DocumentTypeMedicalReport, "", true, "Medical report required based on sum assured / age threshold")
	} else {
		addItem(domain.DocumentTypeMedicalReport, "", false, "Medical report not required for this proposal")
	}

	// ── Income proof: conditional on high SA (SA > 5,00,000) ──
	incomeRequired := params.SumAssured > 500000
	if incomeRequired {
		addItem(domain.DocumentTypeIncomeProof, "", true, "Income proof required for sum assured > ₹5,00,000")
	} else {
		addItem(domain.DocumentTypeIncomeProof, "", false, "Income proof not required for this sum assured")
	}

	// ── Employment proof: conditional on very high SA (SA > 10,00,000) ──
	// For high-value policies, employment proof is needed for underwriting
	employmentRequired := params.SumAssured > 1000000
	if employmentRequired {
		addItem(domain.DocumentTypeEmploymentProof, "", true, "Employment proof required for sum assured > ₹10,00,000")
	} else {
		addItem(domain.DocumentTypeEmploymentProof, "", false, "Employment proof not required for this sum assured")
	}

	// ── Proposer identity documents: when proposer ≠ insured ──
	// When the proposer is a different person, separate identity/address docs
	// are required for the proposer (e.g., parent proposing for minor child).
	// We use the standard DDL enum types (PHOTO_ID, ADDRESS_PROOF) with a SubType
	// qualifier to distinguish proposer documents from insured documents.
	// When uploading, set comments = "sub_type:PROPOSER_PHOTO_ID" to match.
	if !params.IsProposerSameAsInsured {
		addItem(domain.DocumentTypePhotoID, "PROPOSER_PHOTO_ID", true, "Proposer Photo ID required (proposer is different from insured)")
		addItem(domain.DocumentTypeAddressProof, "PROPOSER_ADDRESS_PROOF", true, "Proposer Address Proof required (proposer is different from insured)")
	}

	return checklist
}

// UploadDocument handles document upload for a proposal
// [DOC-POL-002] Upload document
// [INT-POL-008] DMS Integration
// [VR-PI-021] File type: PDF, JPG, PNG, JPEG only
// [VR-PI-022] File size: max 5 MB
// [VR-PI-023] Document date not in future
func (h *DocumentHandler) UploadDocument(sctx *serverRoute.Context, req DocumentUploadRequest) (*resp.DocumentUploadResponse, error) {
	// Step 1: Validate proposal exists
	_, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-002] Proposal %d not found: %v", req.ProposalID, err)
		return nil, err
	}

	// Step 2: Validate file type [VR-PI-021]
	validMimeTypes := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/png":       true,
		"image/jpg":       true,
	}
	if !validMimeTypes[req.MimeType] {
		return nil, fmt.Errorf("[ERR-POL-026] invalid file type: %s. Allowed: PDF, JPG, PNG, JPEG", req.MimeType)
	}

	// Step 3: Validate file size [VR-PI-022] max 5 MB
	maxFileSize := int64(5 * 1024 * 1024) // 5 MB
	if req.FileSize > maxFileSize {
		return nil, fmt.Errorf("[ERR-POL-027] file size %d bytes exceeds maximum allowed size of 5 MB", req.FileSize)
	}

	// Step 4: Validate document date [VR-PI-023] must not be in the future
	docDate, err := time.Parse("2006-01-02", req.DocumentDate)
	if err != nil {
		return nil, fmt.Errorf("[ERR-POL-028] invalid document_date format: expected YYYY-MM-DD, got %s", req.DocumentDate)
	}
	today := time.Now().Truncate(24 * time.Hour)
	if docDate.After(today) {
		return nil, fmt.Errorf("[ERR-POL-028] document_date %s cannot be in the future", req.DocumentDate)
	}

	// Step 5: Generate DMS document ID (in production, this would come from DMS integration INT-POL-008)
	documentID := fmt.Sprintf("doc_%d_%d_%d", req.ProposalID, time.Now().UnixMilli(), req.UploadedBy)

	// Step 6: Create document reference in database
	// document_date is persisted via migration 002_add_document_date.sql
	docRef := &domain.ProposalDocumentRef{
		ProposalID:    req.ProposalID,
		DocumentID:    documentID,
		DocumentType:  domain.DocumentType(req.DocumentType),
		FileName:      &req.FileName,
		FileSizeBytes: &req.FileSize,
		MimeType:      &req.MimeType,
		DocumentDate:  &docDate,
		UploadedBy:    req.UploadedBy,
	}

	if req.Comments != "" {
		docRef.Comments = &req.Comments
	}

	if err := h.documentRepo.CreateDocumentRef(sctx.Ctx, docRef); err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-002] Error creating document ref for proposal %d: %v", req.ProposalID, err)
		return nil, err
	}

	// Step 7: Build download URL
	downloadURL := fmt.Sprintf("/api/v1/proposals/%d/documents/%d", req.ProposalID, docRef.DocRefID)

	log.Info(sctx.Ctx, "[DOC-POL-002] Document uploaded: proposal=%d, type=%s, doc_ref_id=%d", req.ProposalID, req.DocumentType, docRef.DocRefID)
	// Update section completion status
	if err := h.proposalRepo.UpdateSectionComplete(sctx.Ctx, req.ProposalID, "documents", true); err != nil {
		log.Error(sctx.Ctx, "Error updating document section: %v", err)
		return nil, err
	}
	return &resp.DocumentUploadResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusCreated,
			Message:    "Document uploaded successfully",
		},
		DocRefID:          docRef.DocRefID,
		DocumentID:        documentID,
		DocumentType:      req.DocumentType,
		FileName:          req.FileName,
		UploadedAt:        docRef.UploadedAt,
		DownloadURL:       downloadURL,
		IsMissingNotation: false,
	}, nil
}

// DownloadDocument retrieves document metadata and download info
// [DOC-POL-003] Download document
// [INT-POL-008] DMS Integration
// NOTE: In production, this would proxy/redirect to DMS for actual file content.
// This endpoint returns document metadata and download URL.
func (h *DocumentHandler) DownloadDocument(sctx *serverRoute.Context, req DocumentIDUri) (*resp.DocumentUploadResponse, error) {
	docRef, err := h.documentRepo.GetDocumentByID(sctx.Ctx, req.ProposalID, req.DocumentID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-003] Document %d not found for proposal %d: %v", req.DocumentID, req.ProposalID, err)
		return nil, err
	}

	fileName := ""
	if docRef.FileName != nil {
		fileName = *docRef.FileName
	}

	downloadURL := fmt.Sprintf("/api/v1/proposals/%d/documents/%d", req.ProposalID, docRef.DocRefID)

	return &resp.DocumentUploadResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Document retrieved successfully",
		},
		DocRefID:     docRef.DocRefID,
		DocumentID:   docRef.DocumentID,
		DocumentType: string(docRef.DocumentType),
		FileName:     fileName,
		UploadedAt:   docRef.UploadedAt,
		DownloadURL:  downloadURL,
	}, nil
}

// RemoveDocument soft-deletes a document from a proposal
// [DOC-POL-004] Remove document
// [INT-POL-008] DMS Integration
// Business Rule: Soft-deletes document. Requires maker-checker for certain types.
func (h *DocumentHandler) RemoveDocument(sctx *serverRoute.Context, req DocumentIDUri) (*resp.DeleteResponse, error) {
	// Step 1: Verify document exists
	_, err := h.documentRepo.GetDocumentByID(sctx.Ctx, req.ProposalID, req.DocumentID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-004] Document %d not found for proposal %d: %v", req.DocumentID, req.ProposalID, err)
		return nil, err
	}

	// Step 2: Soft delete
	if err := h.documentRepo.SoftDeleteDocument(sctx.Ctx, req.ProposalID, req.DocumentID); err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-004] Error removing document %d from proposal %d: %v", req.DocumentID, req.ProposalID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "[DOC-POL-004] Document removed: proposal=%d, doc_ref_id=%d", req.ProposalID, req.DocumentID)

	return &resp.DeleteResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusNoContent,
			Message:    "Document removed successfully",
		},
	}, nil
}

// GetMissingDocuments retrieves missing documents for a proposal
// [DOC-POL-005] Get missing documents
// Returns documents marked as missing during QC_REVIEW or APPROVAL stage
func (h *DocumentHandler) GetMissingDocuments(sctx *serverRoute.Context, req MissingDocumentsQuery) (*resp.MissingDocumentsListResponse, error) {
	// Step 1: Get proposal for metadata
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-005] Proposal %d not found: %v", req.ProposalID, err)
		return nil, err
	}

	// Step 2: Query missing documents with optional filters
	missingDocs, err := h.documentRepo.GetMissingDocuments(sctx.Ctx, req.ProposalID, req.Stage, req.Status)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-005] Error fetching missing documents for proposal %d: %v", req.ProposalID, err)
		return nil, err
	}

	// Step 3: Build response with counts
	pendingCount := 0
	uploadedCount := 0
	waivedCount := 0
	items := make([]resp.MissingDocumentItem, len(missingDocs))

	for i, doc := range missingDocs {
		switch doc.Status {
		case domain.MissingDocStatusPending:
			pendingCount++
		case domain.MissingDocStatusUploaded:
			uploadedCount++
		case domain.MissingDocStatusWaived:
			waivedCount++
		}

		items[i] = resp.MissingDocumentItem{
			MissingDocID:        doc.MissingDocID,
			ProposalID:          doc.ProposalID,
			DocumentType:        string(doc.DocumentType),
			DocumentDescription: doc.DocumentDescription,
			ReasonMissing:       extractReasonMissing(doc.Notes),
			Stage:               string(doc.Stage),
			NotedBy:             doc.NotedBy,
			NotedAt:             doc.NotedAt,
			Notes:               doc.Notes,
			Status:              string(doc.Status),
			ResolvedBy:          doc.ResolvedBy,
			ResolvedAt:          doc.ResolvedAt,
			ResolutionNotes:     doc.ResolutionNotes,
			UploadedDocumentID:  doc.UploadedDocumentID,
			Waived:              doc.Waived,
			WaivedBy:            doc.WaivedBy,
			WaivedAt:            doc.WaivedAt,
			WaiverReason:        doc.WaiverReason,
			FollowUpRequired:    doc.Status == domain.MissingDocStatusPending,
		}
	}

	return &resp.MissingDocumentsListResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Missing documents retrieved successfully",
		},
		ProposalID:       req.ProposalID,
		ProposalNumber:   proposal.ProposalNumber,
		TotalMissing:     len(missingDocs),
		PendingCount:     pendingCount,
		UploadedCount:    uploadedCount,
		WaivedCount:      waivedCount,
		MissingDocuments: items,
	}, nil
}

// AddMissingDocument records a document as missing during QC/Approval review
// [DOC-POL-006] Record missing document
// Business Rules:
// - Can only be added during QC_REVIEW or APPROVAL stage
// - Used to track documents that need follow-up
func (h *DocumentHandler) AddMissingDocument(sctx *serverRoute.Context, req MissingDocumentCreateRequest) (*resp.MissingDocumentNotationResponse, error) {
	// Step 1: Validate proposal exists and is in the correct stage
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-006] Proposal %d not found: %v", req.ProposalID, err)
		return nil, err
	}

	// Step 2: Enforce stage constraint — missing documents can only be
	// recorded when the proposal is actually in the matching review stage.
	// QC_REVIEW stage requires proposal status QC_PENDING;
	// APPROVAL stage requires proposal status APPROVAL_PENDING.
	stageStatusMap := map[string]domain.ProposalStatus{
		"QC_REVIEW": domain.ProposalStatusQCPending,
		"APPROVAL":  domain.ProposalStatusApprovalPending,
	}
	requiredStatus, ok := stageStatusMap[req.Stage]
	if !ok {
		return nil, fmt.Errorf("[DOC-POL-006] invalid stage: %s", req.Stage)
	}
	if proposal.Status != requiredStatus {
		return nil, fmt.Errorf("[DOC-POL-006] cannot record missing document for stage %s: proposal is in status %s, expected %s",
			req.Stage, proposal.Status, requiredStatus)
	}

	// Step 3: Create missing document record
	missingDoc := &domain.ProposalMissingDocument{
		ProposalID:   req.ProposalID,
		DocumentType: domain.DocumentType(req.DocumentType),
		Stage:        domain.MissingDocumentStage(req.Stage),
		NotedBy:      req.NotedBy,
	}

	if req.DocumentDescription != "" {
		missingDoc.DocumentDescription = &req.DocumentDescription
	}

	// Merge ReasonMissing into Notes so it is persisted in the database.
	// The proposal_missing_documents table has no dedicated reason_missing column,
	// so we combine reason and notes into the notes field for audit trail.
	combinedNotes := ""
	if req.ReasonMissing != "" {
		combinedNotes = "Reason: " + req.ReasonMissing
	}
	if req.Notes != "" {
		if combinedNotes != "" {
			combinedNotes += " | "
		}
		combinedNotes += req.Notes
	}
	if combinedNotes != "" {
		missingDoc.Notes = &combinedNotes
	}

	if err := h.documentRepo.CreateMissingDocument(sctx.Ctx, missingDoc); err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-006] Error creating missing document for proposal %d: %v", req.ProposalID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "[DOC-POL-006] Missing document recorded: proposal=%d, type=%s, stage=%s, missing_doc_id=%d",
		req.ProposalID, req.DocumentType, req.Stage, missingDoc.MissingDocID)

	return &resp.MissingDocumentNotationResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusCreated,
			Message:    "Missing document recorded successfully",
		},
		MissingDocumentItem: resp.MissingDocumentItem{
			MissingDocID:        missingDoc.MissingDocID,
			ProposalID:          missingDoc.ProposalID,
			DocumentType:        string(missingDoc.DocumentType),
			DocumentDescription: missingDoc.DocumentDescription,
			ReasonMissing:       extractReasonMissing(missingDoc.Notes),
			Stage:               string(missingDoc.Stage),
			NotedBy:             missingDoc.NotedBy,
			NotedAt:             missingDoc.NotedAt,
			Notes:               missingDoc.Notes,
			Status:              string(domain.MissingDocStatusPending),
			FollowUpRequired:    true,
		},
	}, nil
}

// extractReasonMissing extracts the reason from the notes field.
// Notes are stored as "Reason: <reason> | <additional notes>".
// Returns nil if no reason prefix is found.
func extractReasonMissing(notes *string) *string {
	if notes == nil {
		return nil
	}
	s := *notes
	const prefix = "Reason: "
	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		return nil
	}
	reason := s[len(prefix):]
	// If there's a separator, extract only the reason part
	if idx := len(reason); idx > 0 {
		for i := 0; i < len(reason); i++ {
			if i+3 <= len(reason) && reason[i:i+3] == " | " {
				reason = reason[:i]
				break
			}
		}
	}
	return &reason
}

// ResolveMissingDocument marks a missing document as resolved (uploaded or waived)
// [DOC-POL-007] Resolve missing document
// Business Rules:
// - Status can be set to UPLOADED (with document reference) or WAIVED (with reason)
// - Waivers may require additional authorization
func (h *DocumentHandler) ResolveMissingDocument(sctx *serverRoute.Context, req MissingDocumentResolveRequest) (*resp.MissingDocumentNotationResponse, error) {
	// Step 1: Get the existing missing document record
	missingDoc, err := h.documentRepo.GetMissingDocumentByID(sctx.Ctx, req.MissingDocID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-007] Missing document %d not found: %v", req.MissingDocID, err)
		return nil, err
	}

	// Step 2: Validate the record belongs to the correct proposal
	if missingDoc.ProposalID != req.ProposalID {
		return nil, fmt.Errorf("[DOC-POL-007] missing document %d does not belong to proposal %d", req.MissingDocID, req.ProposalID)
	}

	// Step 3: Enforce stage constraint — resolution can only happen when
	// the proposal is in the corresponding review stage (QC_PENDING or APPROVAL_PENDING)
	// or has been returned for resubmission (QC_RETURNED, DATA_ENTRY).
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-007] Error fetching proposal %d for stage check: %v", req.ProposalID, err)
		return nil, err
	}
	allowedStatuses := map[domain.ProposalStatus]bool{
		domain.ProposalStatusQCPending:       true,
		domain.ProposalStatusQCReturned:      true,
		domain.ProposalStatusDataEntry:       true,
		domain.ProposalStatusApprovalPending: true,
	}
	if !allowedStatuses[proposal.Status] {
		return nil, fmt.Errorf("[DOC-POL-007] cannot resolve missing document: proposal is in status %s, expected QC_PENDING, QC_RETURNED, DATA_ENTRY, or APPROVAL_PENDING",
			proposal.Status)
	}

	// Step 4: Validate status-specific fields
	resolveStatus := domain.MissingDocumentStatus(req.Status)
	if resolveStatus == domain.MissingDocStatusUploaded && req.UploadedDocumentID == nil {
		return nil, fmt.Errorf("[DOC-POL-007] uploaded_document_id is required when status is UPLOADED")
	}
	if resolveStatus == domain.MissingDocStatusWaived && req.WaiverReason == "" {
		return nil, fmt.Errorf("[DOC-POL-007] waiver_reason is required when status is WAIVED")
	}

	// Step 5: Resolve the missing document
	var waiverReason *string
	var resolutionNotes *string
	if req.WaiverReason != "" {
		waiverReason = &req.WaiverReason
	}
	if req.ResolutionNotes != "" {
		resolutionNotes = &req.ResolutionNotes
	}

	if err := h.documentRepo.ResolveMissingDocument(sctx.Ctx, req.MissingDocID, resolveStatus, req.ResolvedBy, req.UploadedDocumentID, waiverReason, resolutionNotes); err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-007] Error resolving missing document %d: %v", req.MissingDocID, err)
		return nil, err
	}

	// Step 6: Re-fetch to get the updated record
	updatedDoc, err := h.documentRepo.GetMissingDocumentByID(sctx.Ctx, req.MissingDocID)
	if err != nil {
		log.Error(sctx.Ctx, "[DOC-POL-007] Error re-fetching resolved missing document %d: %v", req.MissingDocID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "[DOC-POL-007] Missing document resolved: missing_doc_id=%d, status=%s", req.MissingDocID, req.Status)

	return &resp.MissingDocumentNotationResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Missing document resolved successfully",
		},
		MissingDocumentItem: resp.MissingDocumentItem{
			MissingDocID:        updatedDoc.MissingDocID,
			ProposalID:          updatedDoc.ProposalID,
			DocumentType:        string(updatedDoc.DocumentType),
			DocumentDescription: updatedDoc.DocumentDescription,
			ReasonMissing:       extractReasonMissing(updatedDoc.Notes),
			Stage:               string(updatedDoc.Stage),
			NotedBy:             updatedDoc.NotedBy,
			NotedAt:             updatedDoc.NotedAt,
			Notes:               updatedDoc.Notes,
			Status:              string(updatedDoc.Status),
			ResolvedBy:          updatedDoc.ResolvedBy,
			ResolvedAt:          updatedDoc.ResolvedAt,
			ResolutionNotes:     updatedDoc.ResolutionNotes,
			UploadedDocumentID:  updatedDoc.UploadedDocumentID,
			Waived:              updatedDoc.Waived,
			WaivedBy:            updatedDoc.WaivedBy,
			WaivedAt:            updatedDoc.WaivedAt,
			WaiverReason:        updatedDoc.WaiverReason,
			FollowUpRequired:    updatedDoc.Status == domain.MissingDocStatusPending,
		},
	}, nil
}
