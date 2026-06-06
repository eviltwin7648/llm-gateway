package router

type Router struct {
}

func NewRouter() Router {
	return &Router{}
}

type Provider string

func (r *Router) Route(model string) Provider {
	//logic to be filled
	var provider string
	return provider
}
