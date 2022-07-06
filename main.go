package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"
	"fmt"
	"errors"
	// "os"
	"github.com/pkg/term"
)

func main() {
	flag.Parse()
	args := flag.Args()

	content, err := ioutil.ReadFile(args[0])
    if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(content),"\n")
	flashCards := map[int]FlashCard{}
	for idx, i := range lines {
		flashCard, err := parse(i)
		if err == nil {
			flashCards[idx] = flashCard
		}
	}
	for _, v := range flashCards {
		fmt.Printf("%v\n", v.front)
		fmt.Printf("%v\n\n", v.back)
	}

	for {
		ascii, keycode, err := getChar()
		fmt.Printf("%v %v %v", ascii, keycode, err)
	}
}

type FlashCard struct {
	front string
	back string
	comment string
}

func parse(str string) (FlashCard, error) {
	a, b, found := stringsCut(str, "<!--srs:")
	if !found {
		return FlashCard{}, errors.New("Missing comment start")
	}
	c, d, found := stringsCut(b, "-->")
	if !found {
		return FlashCard{}, errors.New("Missing comment end")
	}
	e, f, found := stringsCut(a, ":")
	if !found {
		return FlashCard{}, errors.New("Missing colon separator")
	}
	_ = d
	return FlashCard{
		front: e,
		back: f,
		comment: c,
	}, nil
}

func stringsCut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

// Returns either an ascii code, or (if input is an arrow) a Javascript key code.
func getChar() (ascii int, keyCode int, err error) {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	} else {
		// Two characters read??
	}
	t.Restore()
	t.Close()
	return
}