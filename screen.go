package react

import (
	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
	"sync"
)

type Screen struct {
	sync.Mutex

	Root *ReactElement
	DOM  *DOMNode

	CellOwners [][]*DOMNode
	CursorX    int
	CursorY    int

	TCellScreen tcell.Screen

	QuitCallback func(error) error
}

func NewScreen() (*Screen, error) {
	if tScreen, err := tcell.NewScreen(); err != nil {
		return nil, err
	} else if err = tScreen.Init(); err != nil {
		return nil, err
	} else {
		tScreen.SetStyle(tcell.StyleDefault.
			Background(tcell.ColorBlack).
			Foreground(tcell.ColorWhite))
		// tScreen.EnableMouse()  // TODO do I want this?

		return &Screen{
			CursorX: 0,
			CursorY: 0,

			TCellScreen: tScreen,
		}, nil
	}
}

func (screen *Screen) Init(root *ReactElement, quitCallback func(error) error) error {
	screen.Root = root
	screen.QuitCallback = quitCallback
	return nil
}

func (screen *Screen) SetCursor(x, y int) {
	width, height := screen.Size()

	if x < 0 {
		x = 0
	} else if x >= width {
		x = width - 1
	}

	if y < 0 {
		y = 0
	} else if y >= height {
		y = height - 1
	}

	screen.CursorX = x
	screen.CursorY = y
}

func (screen *Screen) Size() (width, height int) {
	return screen.TCellScreen.Size()
}

func paint(screen *Screen, dom *DOMNode) error {
	if dom.region != nil {
		for Δx := 0; Δx < dom.width; Δx++ {
			for Δy := 0; Δy < dom.height; Δy++ {
				cell := dom.region.cells[Δx][Δy]
				screen.TCellScreen.SetContent(dom.x+Δx, dom.y+Δy, cell.r, nil, cell.style) // draw to screen
				screen.CellOwners[dom.x+Δx][dom.y+Δy] = dom                                // for HandleKey and HandleMouse
			}
		}
	} else if dom.children != nil {
		for _, child := range dom.children {
			if err := paint(screen, child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (screen *Screen) Resize() {
	// TODO this will be for marking the root as needs-to-be-redrawn, if that turns out to be useful
}

func (screen *Screen) Draw() error {
	screen.Lock()
	defer screen.Unlock()

	width, height := screen.TCellScreen.Size()

	// build CellOwners for Handle{Key,Mouse}, if necessary
	if screen.DOM == nil || screen.DOM.width != width || screen.DOM.height != height {
		cellOwners := make([][]*DOMNode, width)
		for x := 0; x < width; x++ {
			cellOwners[x] = make([]*DOMNode, height)
		}
		screen.CellOwners = cellOwners
	}

	if screen.DOM == nil { // first render
		if dom, err := buildDOM(screen.Root, nil, 0, 0, width, height); err == nil {
			screen.DOM = dom
		} else {
			return err
		}
	} else { // subsequent renders
		if err := reconcileDOM(screen.DOM, 0, 0, width, height); err != nil {
			return err
		}
	}

	if err := paint(screen, screen.DOM); err != nil {
		return err
	}

	screen.TCellScreen.ShowCursor(screen.CursorX, screen.CursorY)

	return nil
}

func (screen *Screen) HandleKey(e *tcell.EventKey) error {
	screen.Lock()
	defer screen.Unlock()

	x := screen.CursorX
	y := screen.CursorY
	domNode := screen.CellOwners[x][y]

	var inner func(*DOMNode) error
	inner = func(domNode *DOMNode) error {
		reactElement := domNode.element
		if propagate, err := reactElement.HandleKey(e); err != nil {
			return err
		} else if propagate {
			if domNode.parent != nil { // propagate to the parent
				return inner(domNode.parent)
			} else { // propagated over the top so use the defaults
				// TODO
				return screen.defaultKeyHandler(e)
			}
		} else { // propagate no further
			return nil
		}
	}

	return inner(domNode)
}

func (screen *Screen) defaultKeyHandler(e *tcell.EventKey) error {
	switch e.Key() {
	case tcell.KeyEsc: // Can always quit with ESC
		if err := screen.QuitCallback(nil); err != nil {
			return errors.Wrap(err, "Failed when executing QuitCallback")
		}
	case tcell.KeyUp:
		screen.SetCursor(screen.CursorX, screen.CursorY-1)
	case tcell.KeyDown:
		screen.SetCursor(screen.CursorX, screen.CursorY+1)
	case tcell.KeyLeft:
		screen.SetCursor(screen.CursorX-1, screen.CursorY)
	case tcell.KeyRight:
		screen.SetCursor(screen.CursorX+1, screen.CursorY)
	default:
		return screen.HandleKey(e)
	}

	return nil
}
