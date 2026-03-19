package main

import (
	"context"
	"policy-management/bootstrap"

	bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
)

func main() {
	app := bootstrapper.New().Options(
		bootstrap.FxRepo,
		bootstrap.FxHandler,
		bootstrap.FxTemporal,
	)
	app.WithContext(context.Background()).Run()
}
