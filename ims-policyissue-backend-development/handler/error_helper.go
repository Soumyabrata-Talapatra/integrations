package handler

import (
	"errors"

	"github.com/jackc/pgx/v5"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
)

// func handleRepoError(err error, notFoundMsg string, internalMsg string) error {

// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return apierrors.HandleErrorWithStatusCodeAndMessage(
// 			apierrors.HTTPErrorNotFound,
// 			notFoundMsg,
// 			nil,
// 		)
// 	}

// 	return apierrors.HandleErrorWithStatusCodeAndMessage(
// 		apierrors.HTTPErrorServerError,
// 		internalMsg,
// 		err,
// 	)
// }
func handleRepoError(err error, notFoundMsg string, internalMsg string) error {

	if errors.Is(err, pgx.ErrNoRows) {
		return notFound(notFoundMsg)
	}

	return serverError(internalMsg, err)
}
func badRequest(msg string) error {
	return apierrors.HandleErrorWithStatusCodeAndMessage(
		apierrors.HTTPErrorBadRequest,
		msg,
		nil,
	)
}

func notFound(msg string) error {
	return apierrors.HandleErrorWithStatusCodeAndMessage(
		apierrors.HTTPErrorNotFound,
		msg,
		nil,
	)
}

func serverError(msg string, err error) error {
	return apierrors.HandleErrorWithStatusCodeAndMessage(
		apierrors.HTTPErrorServerError,
		msg,
		err,
	)
}
func invalidDate(field string) error {
	return badRequest("Invalid " + field + " format. Expected YYYY-MM-DD")
}
