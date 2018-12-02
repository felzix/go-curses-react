package react

import (
	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
)

type DOMNode struct {
	element  *ReactElement
	parent   *DOMNode
	children []*DOMNode
	region   *Region
	x        int
	y        int
	width    int
	height   int
	dirty    bool
}

type Region struct {
	cells  [][]Cell
	x      int
	y      int
	width  int
	height int
}
type Cell struct {
	r     rune
	style tcell.Style
}

func NewRegion(x, y, width, height int) *Region {
	cells := make([][]Cell, width)
	for x := 0; x < width; x++ {
		cells[x] = make([]Cell, height)
	}

	return &Region{
		cells:  cells,
		x:      x,
		y:      y,
		width:  width,
		height: height,
	}
}

/* Mimicks ReactJS's reconcilation algorithm
   See: https://reactjs.org/docs/reconciliation.html
*/
func reconcileDOM(dom *DOMNode, x, y, width, height int) error {
	result, err := dom.element.Draw(width, height)
	if err != nil {
		return err
	}

	if dom.x != x || dom.y != y || dom.width != width || dom.height != height { // resize: rebuild new node
		if newDOM, err := buildDOM(dom.element, dom.parent, x, y, width, height); err == nil {
			*dom = *newDOM
		} else {
			return err
		}
	} else if result.Region != nil { // no children: redraw
		dom.region = result.Region
	} else {
		marked := make(map[string]*DOMNode, len(dom.children))
		for _, node := range dom.children {
			marked[node.element.Key] = node
		}

		newChildren := make([]*DOMNode, len(result.Elements))
		for i := 0; i < len(result.Elements); i++ {
			child := result.Elements[i]
			if oldNode, ok := marked[child.Key]; ok && child.Element.Type == oldNode.element.Type {
				// keep old element
				oldNode.element.Props = child.Props
				if err := reconcileDOM(oldNode, child.X, child.Y, child.Width, child.Height); err != nil {
					return err
				}
				delete(marked, child.Key)
				newChildren[i] = oldNode
			} else {
				// create new element
				child.Element.Props = child.Props
				child.Element.Key = child.Key
				if newNode, err := buildDOM(child.Element, dom, child.X, child.Y, child.Width, child.Height); err == nil {
					newChildren[i] = newNode
				} else {
					return err
				}
			}
		}

		for _, deletedNode := range marked {
			if err := deletedNode.element.OnDismount(); err != nil {
				return err
			}
		}

		dom.children = newChildren
	}

	return nil
}

func buildDOM(element *ReactElement, parent *DOMNode, x, y, width, height int) (*DOMNode, error) {
	node := DOMNode{
		element: element,
		parent:  parent,
		x:       x,
		y:       y,
		width:   width,
		height:  height,
	}

	if err := element.OnMount(); err != nil {
		return nil, err
	}

	result, err := element.Draw(width, height)
	if err != nil {
		return nil, err
	}

	if result.Region != nil {
		node.region = result.Region
	} else if result.Elements != nil {
		node.children = make([]*DOMNode, len(result.Elements))
		for i, child := range result.Elements {
			child.Element.Key = child.Key
			child.Element.Props = child.Props
			if childNode, err := buildDOM(child.Element, &node, x+child.X, y+child.Y, child.Width, child.Height); err == nil {
				node.children[i] = childNode
			} else {
				return nil, err
			}
		}
	} else {
		return nil, errors.New("ReactElement's DrawFn method returned an empty DrawResult instance.")
	}

	return &node, nil
}
