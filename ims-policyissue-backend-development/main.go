package main

import (
	"context"

	"policy-issue-service/bootstrap"

	bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
)

// @title Policy Issue API
// @version 1.0.0
// @description Comprehensive API for policy issue, approvals, quote, calculation for Postal Life Insurance.
// @termsOfService https://www.pli.gov.in/terms

// @contact.name Policy Issue API Support
// @contact.email api-support@pli.gov.in

// @license.name Proprietary
// @license.url https://www.pli.gov.in/terms

// @host api.pli.gov.in
// @BasePath /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer: ` prefix, e.g. "Bearer abcde12345"
func main() {
	app := bootstrapper.New().Options(
		bootstrap.FxHandler,
		bootstrap.FxRepo,
		bootstrap.FxTemporal,
	)
	app.WithContext(context.Background()).Run()
}
