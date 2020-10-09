package display

import (
	"log"

	"github.com/go-vgo/robotgo"
)

func (d *Display) handleKeyEventsLoop() {
	for {
		select {
		case ev, ok := <-d.keyEvQueue:
			if !ok {
				// Client disconnected.
				return
			}
			if ev.IsDown() {
				d.appendDownKeyIfMissing(ev.Key)
				d.dispatchDownKeys()
			} else {
				d.removeDownKey(ev.Key)
			}
		}
	}
}

func (d *Display) dispatchDownKeys() {
	if len(d.downKeys) == 1 {
		ks, ok := robotASCIMap[d.downKeys[0]]
		if !ok {
			log.Println("Unhandled keysym:", d.downKeys[0])
			return
		}
		robotgo.KeyTap(ks)
		return
	}
	args := make([]interface{}, len(d.downKeys))
	for idx, key := range d.downKeys {
		ks, ok := robotASCIMap[key]
		if !ok {
			log.Println("Unhandled keysym:", d.downKeys[0])
			return
		}
		args[len(d.downKeys)-1-idx] = ks
	}
	robotgo.KeyTap(args[0].(string), args[1:]...)
}

func (d *Display) appendDownKeyIfMissing(downKey uint32) {
	for _, k := range d.downKeys {
		if k == downKey {
			return
		}
	}
	d.downKeys = append(d.downKeys, downKey)
}

func (d *Display) removeDownKey(downKey uint32) {
	newDownKeys := make([]uint32, 0)
	for _, k := range d.downKeys {
		if k != downKey {
			newDownKeys = append(newDownKeys, k)
		}
	}
	d.downKeys = newDownKeys
}

var robotASCIMap = map[uint32]string{
	uint32('a'): "a", uint32('A'): "A",
	uint32('b'): "b", uint32('B'): "B",
	uint32('c'): "c", uint32('C'): "C",
	uint32('d'): "d", uint32('D'): "D",
	uint32('e'): "e", uint32('E'): "E",
	uint32('f'): "f", uint32('F'): "F",
	uint32('g'): "g", uint32('G'): "G",
	uint32('h'): "h", uint32('H'): "H",
	uint32('i'): "i", uint32('I'): "I",
	uint32('j'): "j", uint32('J'): "J",
	uint32('k'): "k", uint32('K'): "K",
	uint32('l'): "l", uint32('L'): "L",
	uint32('m'): "m", uint32('M'): "M",
	uint32('n'): "n", uint32('N'): "N",
	uint32('o'): "o", uint32('O'): "O",
	uint32('p'): "p", uint32('P'): "P",
	uint32('q'): "q", uint32('Q'): "Q",
	uint32('r'): "r", uint32('R'): "R",
	uint32('s'): "s", uint32('S'): "S",
	uint32('t'): "t", uint32('T'): "T",
	uint32('u'): "u", uint32('U'): "U",
	uint32('v'): "v", uint32('V'): "V",
	uint32('w'): "w", uint32('W'): "W",
	uint32('x'): "x", uint32('X'): "X",
	uint32('y'): "y", uint32('Y'): "Y",
	uint32('z'): "z", uint32('Z'): "Z",

	uint32(','): ",", uint32('.'): ".", uint32('/'): "/",
	uint32(';'): ";", uint32('\''): "'",
	uint32('['): "[", uint32(']'): "]", uint32('\\'): "\\",
	uint32('-'): "-", uint32('+'): "+",

	0xff08: "backspace",
	0xffff: "delete",
	0xff0d: "enter",
	0xff09: "tab",
	0xff1b: "esc",
	0xff52: "up",    // Up arrow key
	0xff54: "down",  // Down arrow key
	0xff53: "right", // Right arrow key
	0xff51: "left",  // Left arrow key
	0xff50: "home",
	0xff57: "end",
	0xff55: "pageup",
	0xff56: "pagedown",

	0xffbe: "f1",
	0xffbf: "f2",
	0xffc0: "f3",
	0xffc1: "f4",
	0xffc2: "f5",
	0xffc3: "f6",
	0xffc4: "f7",
	0xffc5: "f8",
	0xffc6: "f9",
	0xffc7: "f10",
	0xffc8: "f11",
	0xffc9: "f12",

	0xffe7: "lcmd",   // left command
	0xffe8: "rcmd",   // right command
	0xffe9: "lalt",   // left alt
	0xffea: "ralt",   // right alt
	0xffe3: "lctrl",  // left ctrl
	0xffe4: "rctrl",  // right ctrl
	0xffe1: "lshift", // left shift
	0xffe2: "rshift", // right shift
	0xffe5: "capslock",
	0xff80: "space",
	0xff61: "print",
	0xfd1d: "printscreen", // No Mac support
	0xff9e: "insert",
	0xff67: "menu", //	Windows only

	0xffb0: "0",
	0xffb1: "1",
	0xffb2: "2",
	0xffb3: "3",
	0xffb4: "4",
	0xffb5: "5",
	0xffb6: "6",
	0xffb7: "7",
	0xffb8: "8",
	0xffb9: "9",
}
