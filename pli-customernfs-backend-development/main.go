package main

import (
	"context"

	"customer-nfs-service/bootstrap"

	bootstrapper "gitlab.cept.gov.in/it-2.0-common/n-api-bootstrapper"
)

// Customer NFS Service — Postal Life Insurance
// Module: Customer Non-Financial Service (Address Change, Name Change)
// Technology: Golang, Temporal.io, PostgreSQL, Kafka
// Version: 1.0.0

func main() {
	app := bootstrapper.New().Options(
		// Register all handler modules (HTTP layer)
		bootstrap.FxHandler,
		// Register all repository modules (data access layer)
		bootstrap.FxRepo,
		// Register Temporal workflow client module
		bootstrap.FxTemporal,
		// Register Temporal worker module (polls customer-nfs-tq)
		bootstrap.FxWorker,
	)
	app.WithContext(context.Background()).Run()
}
