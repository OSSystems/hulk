package service

import (
	"encoding/json"
	"net/http"

	"github.com/OSSystems/hulk/api/server/router"
	"github.com/OSSystems/hulk/hulk"
	"github.com/OSSystems/hulk/log"
	"github.com/julienschmidt/httprouter"
)

type serviceRouter struct {
	hulk *hulk.Hulk
}

// Routes returns a route list for /services endpoint
func Routes(hulk *hulk.Hulk) []router.Route {
	r := &serviceRouter{
		hulk: hulk,
	}

	return []router.Route{
		{Method: "GET", Path: "/services", Handle: r.getServices},
		{Method: "GET", Path: "/services/:service", Handle: r.getService},
	}
}

func (sr *serviceRouter) getServices(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	output, _ := json.Marshal(sr.hulk.Services())

	_, err := w.Write(output)
	if err != nil {
		log.Error(err)
	}
}

func (sr *serviceRouter) getService(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	for _, service := range sr.hulk.Services() {
		if service.Name == p.ByName("service") {
			output, _ := json.Marshal(service)

			_, err := w.Write(output)
			if err != nil {
				log.Error(err)
			}

			return
		}
	}
}
