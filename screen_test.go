package react

import (
	"github.com/gdamore/tcell"
	"testing"
)

func Test_simple(t *testing.T) {
	screen, err := NewScreen()
	if err != nil {
		t.Fatal(err)
	}
	defer screen.TCellScreen.Fini()

	screen.Root = Label()
	screen.Root.Props = Properties{"label": "perfect sky"}

	result, err := screen.Root.Draw(screen.TCellScreen.Size())
	if err != nil {
		t.Fatal(err)
	}

	if result.Region.Cells[0][0].R != 'p' {
		t.Errorf(
			`Drawn cell should be "p" but was "%v"`,
			result.Region.Cells[0][0].R,
		)
	}
}

func Test_children(t *testing.T) {
	screen, err := NewScreen()
	if err != nil {
		t.Fatal(err)
	}
	defer screen.TCellScreen.Fini()

	screen.Root = HorizontalLayout()
	screen.Root.Props = Properties{
		"children": []*Child{
			ManagedChild(Label(), "0", Properties{"label": "Wizard of Oz"}),
			ManagedChild(Label(), "1", Properties{"label": "God's Plan"}),
		},
	}

	width, height := screen.TCellScreen.Size()

	if err := screen.Draw(); err != nil { // populates DOM and such
		t.Fatal(err)
	}
	dr, err := screen.Root.Draw(width, height)
	if err != nil {
		t.Fatal(err)
	}

	if len(dr.Elements) != 2 {
		t.Errorf(`Expected "2" children but got "%d"`, len(dr.Elements))
	}

	e0dr, err := dr.Elements[0].Element.Draw(dr.Elements[0].Width, dr.Elements[0].Height)
	if err != nil {
		t.Fatal(err)
	}
	e1dr, err := dr.Elements[1].Element.Draw(dr.Elements[0].Width, dr.Elements[0].Height)
	if err != nil {
		t.Fatal(err)
	}

	e0drRune := e0dr.Region.Cells[0][0].R
	if e0drRune != 'W' {
		t.Errorf(`Expected "W" but got "%v"`, e0drRune)
	}

	e1drRune := e1dr.Region.Cells[0][0].R
	if e1drRune != 'G' {
		t.Errorf(`Expected "G" but got "%v"`, e1drRune)
	}
}

func Test_toDOM(t *testing.T) {
	r := HorizontalLayout()
	r.Props = Properties{
		"children": []*Child{
			ManagedChild(Label(), "0", Properties{"label": "Wizard of Oz"}),
			ManagedChild(Label(), "1", Properties{"label": "God's Plan"}),
		},
	}

	dom, err := buildDOM(r, nil, 0, 0, 20, 5)
	if err != nil {
		t.Fatal(err)
	}

	if len(dom.children) != 2 {
		t.Errorf(`Expected "2" children but has "%d" children`, len(dom.children))
	}

	if expected := dom.children[1].region.Cells[1][0].R; expected != 'o' {
		t.Errorf(`Expected "o" but got "%v"`, expected)
	}
}

func Test_paint(t *testing.T) {
	screen, err := NewScreen()
	if err != nil {
		t.Fatal(err)
	}
	defer screen.TCellScreen.Fini()

	screen.Root = HorizontalLayout()
	screen.Root.Props = Properties{
		"children": []*Child{
			ManagedChild(Label(), "0", Properties{"label": "Wizard of Oz"}),
			ManagedChild(Label(), "1", Properties{"label": "God's Plan"}),
		},
	}

	if err := screen.Draw(); err != nil {
		t.Fatal(err)
	}

	r, _, style, _ := screen.TCellScreen.GetContent(1, 0)
	if r != 'i' {
		t.Errorf(`Expected "i" but got "%s"`, string(r))
	}
	if style != tcell.StyleDefault {
		t.Errorf(`Expected default style (0) but got "%d"`, style)
	}

	width, height := screen.TCellScreen.Size()
	runes := make([][]rune, width)

	for x := 0; x < width; x++ {
		runes[x] = make([]rune, height)
		for y := 0; y < height; y++ {
			r, _, _, _ = screen.TCellScreen.GetContent(x, y)
			runes[x][y] = r
		}
	}

	r, _, style, _ = screen.TCellScreen.GetContent(0, 1)
	if r != 'G' {
		t.Errorf(`Expected "G" but got "%s"`, string(r))
	}
	if style != tcell.StyleDefault {
		t.Errorf(`Expected default style (0) but got "%d"`, style)
	}
}
