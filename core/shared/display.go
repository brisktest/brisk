// Copyright 2024 Brisk, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shared

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/moby/term"
	"github.com/morikuni/aec"
	"github.com/spf13/viper"
)

// returns a new display
func NewDisplay(fd uintptr, screens []*ScreenSect) OutputWriter {
	//TODO make it so we can default to just printing out when not a terminal (for CI and straight to file)
	if !term.IsTerminal(fd) {
		log.Print("Should be a terminal this probably won't work too good")
	}

	d := Display{Screens: screens, Fd: fd, Writer: os.Stdout}
	lock := &sync.Mutex{}
	for _, v := range screens {
		v.displayLock = lock
		v.display = &d
	}
	return &d
}

type OutputWriter interface {
	Println(output string)
	GetStatusScreen() *ScreenSect
}

type Display struct {
	Screens []*ScreenSect
	Fd      uintptr
	Writer  io.Writer
	Mux     sync.Mutex
}

func (d *Display) height() uint16 {
	var height uint16
	for _, v := range d.Screens {
		height += v.Height()
	}

	return height
}

func moveCursorDown(fd uintptr, height uint16) {
	down := aec.EmptyBuilder.Down(uint(height)).ANSI
	fmt.Print(down)

}
func moveCursorUp(fd uintptr, height uint16) {
	up := aec.EmptyBuilder.Up(uint(height)).ANSI
	fmt.Print(up)

}
func (d *Display) GetStatusScreen() *ScreenSect {
	return d.Screens[1]
}

func NewStatusScreen() *ScreenSect {
	return &ScreenSect{Top: 0, Bottom: 1, Name: "Status", Type: STATUS_SCREEN}

}

func NewScrollingScreen() *ScreenSect {
	return &ScreenSect{Top: 2, Bottom: 3, Name: "Scrolling", Type: SCROLLING_SCREEN, repeated: true}

}

func (d *Display) needsRedraw() bool {
	return true
}

type ScreenSect struct {
	//these are the column numbers
	Top             uint16
	Bottom          uint16
	Name            string
	Content         []string
	dirty           bool
	mux             sync.Mutex
	BackgroundColor aec.RGB8Bit
	Color           aec.RGB8Bit
	Type            int
	// UpdateFunc  func()
	repeated    bool
	lineLength  int
	displayLock *sync.Mutex
	display     *Display
}

func (s *ScreenSect) SetBackgroundColor(R uint8, G uint8, B uint8) {
	s.BackgroundColor = aec.NewRGB8Bit(uint8(R), uint8(G), uint8(B))
}

func (s *ScreenSect) SetColor(R uint8, G uint8, B uint8) {
	s.Color = aec.NewRGB8Bit(uint8(R), uint8(G), uint8(B))
}

func (s *ScreenSect) Height() uint16 {

	return uint16(s.lineLength)
}

func (s *ScreenSect) needsRedraw() bool {

	return s.Type == STATUS_SCREEN || s.dirty
}

func (d *Display) Println(output string) {
	d.Mux.Lock()
	defer d.Mux.Unlock()

	s := d.Screens[0]

	builder := aec.EmptyBuilder
	col := aec.Column(0)

	// label := builder.Color8BitB(s.BackgroundColor).Color8BitF(s.Color).With(col).ANSI
	label := builder.EraseLine(2).Color8BitF(s.Color).With(col).ANSI
	line := strings.ReplaceAll(output, "\n", " ")

	fmt.Println(label.Apply(line))
	d.Refresh()
}

func (d *Display) Refresh() {
	s2 := d.Screens[1]
	if len(s2.Content) > 0 {
		statusBuilder := aec.EmptyBuilder
		statusCol := aec.Column(0)
		eraseLine := aec.EraseLine(aec.EraseModes.All)
		fmt.Print(aec.Column(0))

		statusLabel := statusBuilder.Color8BitF(s2.Color).With(statusCol).With(eraseLine).ANSI

		fmt.Print(statusLabel.Apply(s2.Content[0]))

	}
}

func truncateString(s string, maxLength int) string {
	if os.Getenv("NO_TRUNCATE_FOR_SCREEN") == "true" {
		return s
	}

	if len(s) <= maxLength {
		return s
	} else {
		return s[:maxLength-1]
	}
}

// func (s *ScreenSect) Update(content []string) {
// 	s.mux.Lock()

// 	s.Content = content
// 	s.dirty = true
// 	s.mux.Unlock()
// 	s.UpdateFunc()

// }

func (s *ScreenSect) Set(str string) {

	if os.Getenv("PLAIN") == "true" || viper.GetBool("CI") || !term.IsTerminal(os.Stdout.Fd()) {
		return
	}

	if s.display == nil {
		return
	}

	s.Content = make([]string, 1)
	s.Content[0] = str
	s.display.Mux.Lock()
	defer s.display.Mux.Unlock()

	s.display.Refresh()
}

func ProgressBar(size int) {

	fd := os.Stdout.Fd()
	if !term.IsTerminal(fd) {
		log.Print("Should be a terminal this probably won't work too good")
	}

	builder := aec.EmptyBuilder
	n := GetWidth(fd) - 6
	up2 := aec.Up(2)
	col := aec.Column(n)
	bar := aec.Color8BitF(aec.NewRGB8Bit(64, 255, 64))
	label := builder.LightRedF().Underline().With(col).ANSI

	for i := 0; i <= int(n-2); i++ {
		fmt.Print(up2)
		fmt.Println(label.Apply(fmt.Sprint(i, "/", n-2)))
		fmt.Print("[")
		fmt.Print(bar.Apply(strings.Repeat("=", i)))
		fmt.Println(col.Apply("]"))
	}
}

func GetWidth(fd uintptr) uint {
	size, err := term.GetWinsize(fd)
	if err != nil {
		log.Panic(err)
	}
	return uint(size.Width)
}

func GetHeight(fd uintptr) uint {
	size, err := term.GetWinsize(fd)
	if err != nil {
		log.Panic(err)
	}
	return uint(size.Height)
}

const SCROLLING_SCREEN = 0
const STATUS_SCREEN = 1
