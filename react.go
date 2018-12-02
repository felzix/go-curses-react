package react

import (
	"github.com/gdamore/tcell"
)

type ReactElement struct {
	Type string
	Key  string

	Props Properties
	State State

	OnMountFn    func(r *ReactElement) error
	OnDismountFn func(r *ReactElement) error
	DrawFn       func(*ReactElement, int, int) (*DrawResult, error)
	HandleKeyFn  func(*ReactElement, *tcell.EventKey) (bool, error)
}
type Properties map[string]interface{}
type State map[string]interface{}

// treat this like a union
type DrawResult struct {
	Elements []Child
	Region   *Region
}
type Child struct {
	Element *ReactElement
	Props   Properties
	Key     string
	X       int
	Y       int
	Width   int
	Height  int
}

func NewChild(r *ReactElement, key string, width, height int, props Properties) *Child {
	return &Child{
		Element: r,
		Props:   props,
		Key:     key,
		Width:   width,
		Height:  height,
	}
}

func ManagedChild(r *ReactElement, key string, props Properties) *Child {
	return &Child{
		Element: r,
		Props:   props,
		Key:     key,
	}
}

func (r *ReactElement) OnMount() error {
	if r.OnMountFn != nil {
		return r.OnMountFn(r)
	}
	return nil
}

func (r *ReactElement) OnDismount() error {
	if r.OnDismountFn != nil {
		return r.OnDismountFn(r)
	}
	return nil
}

func (r *ReactElement) Draw(width, height int) (*DrawResult, error) {
	if r.DrawFn != nil {
		return r.DrawFn(r, width, height)
	}
	return nil, nil
}

func (r *ReactElement) HandleKey(e *tcell.EventKey) (bool, error) {
	if r.HandleKeyFn != nil {
		return r.HandleKeyFn(r, e)
	}
	return true, nil // keep propagating
}
