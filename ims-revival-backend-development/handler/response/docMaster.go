package response

import (
	"plirevival/core/domain"
	"plirevival/core/port"
)

type DocumentResponse struct {
	Message  string          `json:"message"`
	Filename string          `json:"filename"`
	Document domain.Document `json:"document"`
}

// Function to create a new leave document response
func NewDocumentResponse(doc domain.Document) DocumentResponse {
	return DocumentResponse{
		Message:  "document uploaded successfully",
		Filename: doc.DocumentName,
		Document: doc,
	}
}

type DocUploadResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DocumentResponse `json:"data"`
}
