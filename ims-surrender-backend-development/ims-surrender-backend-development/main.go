package main

import (
	"context"

	bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/bootstrap"
)

func main() {
	app := bootstrapper.New().Options(
		bootstrap.FxRepo,
		bootstrap.FxHandler,
		bootstrap.FxTemporal,
	)
	app.WithContext(context.Background()).Run()
}
