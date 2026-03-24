package routes

import (
	router "gitlab.cept.gov.in/it-2.0-common/n-api-server"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	"go.uber.org/fx"
)

// RoutesParams defines the Fx dependencies for route registration
type RoutesParams struct {
	fx.In
	Router   *router.Router
	Handlers []serverHandler.Handler `group:"servercontrollers"`
}

// Routes registers all handler routes with the router.
// The actual route registration is handled by the n-api-server framework
// which automatically registers handlers via the servercontrollers group.
func Routes(p RoutesParams) {
	// Handler registration is performed automatically by the framework
	// through the servercontrollers group tag. This function exists
	// to trigger Fx dependency resolution and ensure handlers are initialized.
}
