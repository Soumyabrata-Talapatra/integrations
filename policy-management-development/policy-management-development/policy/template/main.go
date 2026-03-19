package main

import (
	"context"
	"pisapi/bootstrap"

	bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
)


func main() {
	app := bootstrapper.New().Options(
		// bootstrap.Fxvalidator,
		bootstrap.FxHandler,
		bootstrap.FxRepo,
	)
	app.WithContext(context.Background()).Run()
}
