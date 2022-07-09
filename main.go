package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"
	"strconv"
)

func main() {
	flag.Parse()
	args := flag.Args()

	file, err := readFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range file.cards {
		fmt.Printf("%v", strings.TrimSpace(v.front))
		var str string
		fmt.Scanln(&str)
		fmt.Printf("---\n%v\n\n", strings.TrimSpace(v.back))

		if time.Since(v.sm.nextReview) > 0 {
			fmt.Printf("Score (0-5): ")
			var ratingStr string
			fmt.Scanln(&ratingStr)
			ratingInt, err := strconv.Atoi(ratingStr)
			if err != nil {
				return
			}
			if ratingInt < 0 || ratingInt > 5 {
				return
			}
			fmt.Printf("\n")
			var prevSm = *v.sm

			v.sm.Advance(float64(ratingInt))

			fmt.Printf("Easiness: %.2f → %.2f\n", prevSm.easiness, v.sm.easiness)
			fmt.Printf("Repetition: %v✓ → %v✓\n", prevSm.repetition, v.sm.repetition)
			fmt.Printf("Interval: %vd → %vd\n\n", prevSm.interval, v.sm.interval)
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
	str := c.front + ":" + c.back + "<!--srs" + comment + "-->"
	return str, nil
}

func (c *Card) UnmarshalString(str string) error {
	a, b, found := stringsCut(str, "<!--srs")
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
