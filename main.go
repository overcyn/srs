package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/pkg/term"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
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
		ratingInt, err := strconv.Atoi(string(ascii))
		if err != nil {
			return
		}
		if ratingInt < 0 || ratingInt > 5 {
			return
		}
		rating := float64(ratingInt)
		fmt.Printf("\n\n")

		if time.Since(v.sm.NextReview) > 0 {
			v.sm.Advance(rating)
		}

		// Write to file
		if err := file.write(); err != nil {
			log.Fatal(err)
		}
	}
}

type File struct {
	filename string
	lines    []string
	cards    map[int]*Card
}

func (f *File) write() error {
	buf := &bytes.Buffer{}
	for i, line := range f.lines {
		toWrite := line
		if card, ok := f.cards[i]; ok {
			var err error
			if toWrite, err = card.MarshalString(); err != nil {
				return err
			}
		}
		if _, err := buf.WriteString(toWrite); err != nil {
			return err
		}
		if i != len(f.lines)-1 {
			if _, err := buf.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	if err := os.WriteFile(f.filename, buf.Bytes(), fs.ModePerm); err != nil {
		return err
	}
	return nil
}

func readFile(filename string) (*File, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	cards := map[int]*Card{}
	for idx, i := range lines {
		c := Card{}
		if err := c.UnmarshalString(i); err != nil {
			continue
		}
		cards[idx] = &c
	}

	return &File{
		filename: filename,
		lines:    lines,
		cards:    cards,
	}, nil
}

type Card struct {
	front string
	back  string
	sm    *Supermemo2
}

func (c *Card) MarshalString() (string, error) {
	comment, err := c.sm.Marshal()
	if err != nil {
		return "", err
	}
	str := c.front + ":" + c.back + "<!--srs:" + comment + "-->"
	return str, nil
}

func (c *Card) UnmarshalString(str string) error {
	a, b, found := stringsCut(str, "<!--srs:")
	if !found {
		return errors.New("Missing comment start")
	}
	c1, _, found := stringsCut(b, "-->")
	if !found {
		return errors.New("Missing comment end")
	}
	front, back, found := stringsCut(a, ":")
	if !found {
		return errors.New("Missing colon separator")
	}
	c.front = front
	c.back = back
	c.sm = NewSupermemo2()
	if c1 == "" {
		return nil
	}
	err := c.sm.Unmarshal(c1)
	if err != nil {
		return err
	}
	return nil
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
