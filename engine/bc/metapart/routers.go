package metapart

import (
	"fmt"

	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/component"
	"github.com/tutumagi/pitaya/docgenerator"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/route"
)

// type RouteKind = string

// const (
// 	EntityKind  = "entity"
// 	ServiceKind = "service"
// 	SpaceKind   = "space"
// )

// type routeManager map[string]*Routers

// var rtManager routeManager = map[string]*Routers{}

// func (r routeManager) getRoute(typName string) *Routers {
// 	if routers, ok := r[typName]; ok {
// 		return routers
// 	}
// 	routers := NewRouters(nil)
// 	r[typName] = routers

// 	return routers
// }

type (
	// Routers service
	Routers struct {
		handlerServices map[string]*component.Service // all registered hanlder service
		remoteServices  map[string]*component.Service // all registered remote service

		handlers map[string]*component.Handler // all handler method
		remotes  map[string]*component.Remote  // all remote method

		server *cluster.Server
	}
)

// NewRouters creates and returns a new handler service
func NewRouters(server *cluster.Server) *Routers {
	h := &Routers{
		server: server,

		handlerServices: make(map[string]*component.Service),
		remoteServices:  make(map[string]*component.Service),

		handlers: make(map[string]*component.Handler),
		remotes:  make(map[string]*component.Remote),
	}

	return h
}

// Register registers components
func (h *Routers) Register(comp component.Component, opts ...component.Option) error {
	s := component.NewService(comp, opts)

	if _, ok := h.handlerServices[s.Name]; ok {
		return fmt.Errorf("handler: service already defined: %s", s.Name)
	}

	if err := s.ExtractHandler(); err != nil {
		return err
	}

	// register all handlers
	h.handlerServices[s.Name] = s
	for name, handler := range s.Handlers {
		h.handlers[fmt.Sprintf("%s.%s", s.Name, name)] = handler
	}
	return nil
}

func (h *Routers) getHandler(rt *route.Route) (*component.Handler, error) {
	handler, ok := h.handlers[rt.Short()]
	if !ok {
		e := fmt.Errorf("pitaya/handler: %s not found", rt.String())
		return nil, e
	}
	return handler, nil
}

func (h *Routers) getRemote(rt *route.Route) (*component.Remote, error) {
	remote, ok := h.remotes[rt.Short()]
	if !ok {
		e := fmt.Errorf("pitaya/remote: %s not found", rt.String())
		return nil, e
	}
	return remote, nil
}

// Register registers components
func (h *Routers) RegisterRemote(comp component.Component, opts ...component.Option) error {
	s := component.NewService(comp, opts)

	if _, ok := h.remoteServices[s.Name]; ok {
		return fmt.Errorf("remote: service already defined: %s", s.Name)
	}

	if err := s.ExtractRemote(); err != nil {
		return err
	}

	h.remoteServices[s.Name] = s
	// register all remotes
	for name, remote := range s.Remotes {
		h.remotes[fmt.Sprintf("%s.%s", s.Name, name)] = remote
	}

	return nil
}

// DumpServices outputs all registered services
func (h *Routers) DumpServices() {
	for name, hh := range h.handlers {
		logger.Log.Infof("registered handler %s, isRawArg: %t, type: %v", name, hh.IsRawArg, hh.MessageType)
	}
	for name := range h.remotes {
		logger.Log.Infof("registered remote %s", name)
	}
}

// Docs returns documentation for handlers
func (h *Routers) docsHandler(getPtrNames bool) (map[string]interface{}, error) {
	if h == nil {
		return map[string]interface{}{}, nil
	}
	return docgenerator.HandlersDocs(h.server.Type, h.handlerServices, getPtrNames)
}

// Docs returns documentation for remotes
func (h *Routers) docsRemote(getPtrNames bool) (map[string]interface{}, error) {
	if h == nil {
		return map[string]interface{}{}, nil
	}
	return docgenerator.RemotesDocs(h.server.Type, h.remoteServices, getPtrNames)
}

func Documents(getPtrNames bool) (map[string]interface{}, error) {
	handlerDocs := make(map[string]interface{})
	remoteDocs := make(map[string]interface{})
	for typName, desc := range registerEntityTypes {
		_ = typName
		tmp, err := desc.Routers.docsHandler(getPtrNames)
		if err != nil {
			return nil, err
		}
		for k, v := range tmp {
			handlerDocs[k] = v
		}
		tmp, err = desc.Routers.docsRemote(getPtrNames)
		if err != nil {
			return nil, err
		}
		for k, v := range tmp {
			remoteDocs[k] = v
		}
	}

	return map[string]interface{}{
		"handlers": handlerDocs,
		"remotes":  remoteDocs,
	}, nil
}
