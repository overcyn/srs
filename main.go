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

	file, err := readFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range file.cards {
		fmt.Printf("%v\n", strings.TrimSpace(v.front))

		_, _, err := getChar()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n\n", strings.TrimSpace(v.back))

		fmt.Printf("Score (1-5): ")
		ascii, _, err := getChar()
		if err != nil {
			log.Fatal(err)
		}
		found := false
		for _, i := range []string{"0", "1", "2", "3", "4", "5"} {
			if string(ascii) == i {
				found = true
			}
		}
		if !found {
			return
		}
		fmt.Printf("\n\n")
	}
}

type File struct {
	filename string
	lines []string
	cards map[int]*Card
}

func readFile(filename string) (*File, error) {
	content, err := ioutil.ReadFile(filename)
    if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content),"\n")
	cards := map[int]*Card{}
	for idx, i := range lines {
		flashCard, err := parseCard(i)
		if err == nil {
			cards[idx] = flashCard
		}
	}

	return &File{
		filename: filename,
		lines: lines,
		cards: cards,
	}, nil
}

type Card struct {
	front string
	back string
	comment string
}

func parseCard(str string) (*Card, error) {
	a, b, found := stringsCut(str, "<!--srs:")
	if !found {
		return nil, errors.New("Missing comment start")
	}
	c, d, found := stringsCut(b, "-->")
	if !found {
		return nil, errors.New("Missing comment end")
	}
	e, f, found := stringsCut(a, ":")
	if !found {
		return nil, errors.New("Missing colon separator")
	}
	_ = d
	return &Card{
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