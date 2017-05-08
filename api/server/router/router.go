package router

import "github.com/julienschmidt/httprouter"

// Route defines a route
type Route struct {
	Method string
	Path   string
	Handle httprouter.Handle
}

// NewRouter creates a new HTTP router for routes
func NewRouter(routes []Route) *httprouter.Router {
	router := httprouter.New()

	for _, r := range routes {
		router.Handle(r.Method, r.Path, r.Handle)
	}

	return router
}
