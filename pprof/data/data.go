package data

type Cmd interface {
	Name() string
	Run()
}
