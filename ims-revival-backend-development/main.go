package main

import (
	"context"
	"plirevival/bootstrap"

	bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
)


func main() {
	app := bootstrapper.New().Options(
		// bootstrap.Fxvalidator,
		bootstrap.FxRepo,
		bootstrap.FxHandler,
		bootstrap.FxActivities,
		bootstrap.Fxtemporal,
		//bootstrap.FxCache,
		//bootstrap.FxMinIO,
	)
	app.WithContext(context.Background()).Run()
}
