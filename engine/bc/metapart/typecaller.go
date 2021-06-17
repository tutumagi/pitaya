package metapart

// // Register register a component with options
// func (desc *TypeDesc) Register(c component.Component, options ...component.Option) {
// 	// handlerComp = append(handlerComp, regComp{c, options})
// 	desc.routers.Register(c, options)
// }

// // RegisterRemote register a remote component with options
// func (desc *TypeDesc) RegisterRemote(c component.Component, options ...component.Option) {
// 	// remoteComp = append(remoteComp, regComp{c, options})
// 	desc.routers.Register(c, options)
// }

func startupComponents() {
	// // component initialize hooks
	// for _, c := range handlerComp {
	// 	c.comp.Init()
	// }

	// // component after initialize hooks
	// for _, c := range handlerComp {
	// 	c.comp.AfterInit()
	// }

	// // register all components
	// for _, c := range handlerComp {
	// 	if err := handlerService.Register(c.comp, c.opts); err != nil {
	// 		logger.Log.Errorf("Failed to register handler: %s", err.Error())
	// 	}
	// }

	// // component initialize hooks
	// for _, c := range remoteComp {
	// 	c.comp.Init()
	// }

	// // component after initialize hooks
	// for _, c := range remoteComp {
	// 	c.comp.AfterInit()
	// }

	// // register all remote components
	// for _, c := range remoteComp {
	// 	if remoteService == nil {
	// 		logger.Log.Warn("registered a remote component but remoteService is not running! skipping...")
	// 	} else {
	// 		if err := remoteService.Register(c.comp, c.opts); err != nil {
	// 			logger.Log.Errorf("Failed to register remote: %s", err.Error())
	// 		}
	// 	}
	// }

	// handlerService.DumpServices()
	// if remoteService != nil {
	// 	remoteService.DumpServices()
	// }
}

func shutdownComponents() {
	// // reverse call `BeforeShutdown` hooks
	// length := len(handlerComp)
	// for i := length - 1; i >= 0; i-- {
	// 	handlerComp[i].comp.BeforeShutdown()
	// }

	// // reverse call `Shutdown` hooks
	// for i := length - 1; i >= 0; i-- {
	// 	handlerComp[i].comp.Shutdown()
	// }

	// length = len(remoteComp)
	// for i := length - 1; i >= 0; i-- {
	// 	remoteComp[i].comp.BeforeShutdown()
	// }

	// // reverse call `Shutdown` hooks
	// for i := length - 1; i >= 0; i-- {
	// 	remoteComp[i].comp.Shutdown()
	// }
}
