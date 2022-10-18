package es

// CommandHandlerMiddleware is a function that middlewares can implement to be
// able to chain.
type CommandHandlerMiddleware func(CommandHandler) CommandHandler

// UseCommandHandlerMiddleware wraps a CommandHandler in one or more middleware.
func UseCommandHandlerMiddleware(h CommandHandler, middleware ...CommandHandlerMiddleware) CommandHandler {
	// Apply in reverse order.
	for i := len(middleware) - 1; i >= 0; i-- {
		m := middleware[i]
		h = m(h)
	}

	return h
}
