package es

type IsProjector interface {
	IsProjector()
}

type BaseProjector struct {
}

func (BaseProjector) IsProjector() {}
