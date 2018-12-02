package react

import (
	"fmt"
	"github.com/gdamore/tcell"
)

func Label() *ReactElement {
	return &ReactElement{
		Type: "Label",
		DrawFn: func(r *ReactElement, maxWidth, maxHeight int) (*DrawResult, error) {
			label := r.Props["label"].(string)

			width := len(label)
			if width > maxWidth {
				width = maxWidth
			}

			result := DrawResult{
				Region: NewRegion(0, 0, maxWidth, maxHeight),
			}

			for x := 0; x < width; x++ {
				result.Region.cells[x][0] = Cell{
					r:     rune(label[x]),
					style: tcell.StyleDefault,
				}
			}
			return &result, nil
		},
	}
}

func Line() *ReactElement {
	return &ReactElement{
		Type: "Line",
		DrawFn: func(r *ReactElement, maxWidth, maxHeight int) (*DrawResult, error) {
			char := r.Props["char"].(rune)
			length := r.Props["length"].(int)

			if length > maxWidth { // as long as possible
				length = maxWidth
			} else if length == 0 { // default to max width
				length = maxWidth
			}

			result := DrawResult{
				// TODO use length instead of maxWidth once "Let ReactElements take up less space" is resolved
				Region: NewRegion(0, 0, maxWidth, 1),
			}

			for x := 0; x < length; x++ {
				result.Region.cells[x][0] = Cell{
					r:     char,
					style: tcell.StyleDefault,
				}
			}

			return &result, nil
		},
	}
}

func TextEntry() *ReactElement {
	return &ReactElement{
		Type: "TextEntry",
		State: State{
			"value":    "",
			"finished": false,
		},
		DrawFn: func(r *ReactElement, maxWidth, maxHeight int) (*DrawResult, error) {
			label := r.Props["label"].(string)
			value := r.State["value"].(string)

			s := fmt.Sprintf("%s: %s", label, value)
			length := len(s)

			if length > maxWidth { // as long as possible
				length = maxWidth
			}

			result := DrawResult{
				Region: NewRegion(0, 0, maxWidth, maxHeight),
			}

			for x := 0; x < length; x++ {
				result.Region.cells[x][0] = Cell{
					r:     rune(s[x]),
					style: tcell.StyleDefault,
				}
			}

			return &result, nil
		},
		HandleKeyFn: func(r *ReactElement, e *tcell.EventKey) (bool, error) {
			// TODO handle backspace
			// TODO handle more commands like emacs-mode (ctrl-e, ctrl-a, pasting?)
			whenFinished := r.Props["whenFinished"].(func(string) error)
			value := r.State["value"].(string)
			finished := r.State["finished"].(bool)

			if finished { // trigger only once
				return true, nil
			} else if e.Key() == tcell.KeyEnter { // trigger callback
				r.State["finished"] = true
				return false, whenFinished(value)
			} else { // keep going
				r.State["value"] = value + string(e.Rune())
				return false, nil
			}
		},
	}
}

func Notification() *ReactElement {
	return &ReactElement{
		Type: "Notification",
		State: State{
			"finished": false,
		},
		DrawFn: func(r *ReactElement, maxWidth, maxHeight int) (*DrawResult, error) {
			label := r.Props["label"].(string)

			return &DrawResult{
				Elements: []Child{
					*NewChild(Label(), label, maxWidth, maxHeight, Properties{"label": label}),
				}}, nil
		},
		HandleKeyFn: func(r *ReactElement, e *tcell.EventKey) (bool, error) {
			whenFinished := r.Props["whenFinished"].(func() error)
			finished := r.State["finished"].(bool)

			if finished { // trigger only once
				return true, nil
			} else { // trigger callback
				r.State["finished"] = true
				return false, whenFinished()
			}
		},
	}
}

// TODO style: padding, reaction to smaller space (horizontal and also maybe vertical)
func HorizontalLayout() *ReactElement {
	return &ReactElement{
		Type: "HorizontalLayout",
		DrawFn: func(r *ReactElement, maxWidth, maxHeight int) (*DrawResult, error) {
			children := r.Props["children"].([]*Child)

			result := DrawResult{
				Elements: make([]Child, len(children)),
			}

			for i, child := range children {
				if i >= maxHeight { // just stop printing
					break
				}

				result.Elements[i] = Child{
					Element: child.Element,
					Key:     child.Key,
					Props:   child.Props,
					X:       0,
					Y:       i,
					Width:   maxWidth,
					Height:  1,
				}
			}

			return &result, nil
		},
	}
}
