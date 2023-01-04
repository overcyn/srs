package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"
	"strconv"
	"math/rand"
	"github.com/BurntSushi/toml"
	// "github.com/pelletier/go-toml"
)

func main() {
	showAll := flag.Bool("all", false, "")
	flag.Parse()
	args := flag.Args()

	if err := insertNewSupermemo(args[0]); err != nil {
		log.Fatal(err)
	}

	file, err := readFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	if err := file.present(*showAll); err != nil {
		log.Fatal(err)
	}
}

func (f *File) present(showAll bool) error {
	for _, v := range f.cards {
		needsReview := time.Since(v.sm.nextReview) > 0
		if needsReview || showAll {
			fmt.Printf("%v", strings.TrimSpace(v.front))
			var str string
			fmt.Scanln(&str)
			fmt.Printf("---\n%v\n\n", strings.TrimSpace(v.back))
		}
		if needsReview {
			fmt.Printf("Score (0-5): ")
			var ratingStr string
			fmt.Scanln(&ratingStr)
			ratingInt, err := strconv.Atoi(ratingStr)
			if err != nil {
				return err
			}
			if ratingInt < 0 || ratingInt > 5 {
				return errors.New("Invalid rating")
			}
			fmt.Printf("\n")

			var prevSm = *v.sm
			v.sm.Advance(float64(ratingInt))

			fmt.Printf("Easiness: %.2f → %.2f\n", prevSm.easiness, v.sm.easiness)
			fmt.Printf("Repetition: %v✓ → %v✓\n", prevSm.repetition, v.sm.repetition)
			fmt.Printf("Interval: %vd → %vd\n\n", prevSm.interval, v.sm.interval)

			// Write to file
			if err := f.writeCard(v); err != nil {
				return err
			}
		}
	}
	return nil
}

type File struct {
	filename string
	cards    []*Card
}

func (f *File) writeCard(card *Card) error {
	// Read the file
	buf, err := os.ReadFile(f.filename)
	if err != nil {
		return err
	}
	str := string(buf)

	// Marshal the card info
	smStr, err := card.sm.Marshal()
	if err != nil {
		return err
	}

	// Replace the string
	str = strings.Replace(str, card.info, smStr, 1)

	// Write the file
	if err := os.WriteFile(f.filename, []byte(str), fs.ModePerm); err != nil {
		return err
	}
	return nil
}

func insertNewSupermemo(filename string) error {
	// Read the file
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	str := string(bytes)

	// Replace `[""]` with nano id https://zelark.github.io/nano-id-cc/
	for strings.Contains(str, "[\"\"]\n") {
		randStr := randSeq(21)
		str = strings.Replace(str, "[\"\"]\n", "[\"" + randStr + "\"]\n", 1)
	}

	// Replace `i = ""` with default supermemo
	for strings.Contains(str, "i = \"\"") {
		sm := NewSupermemo2()
		smStr, err := sm.Marshal()
		if err != nil {
			return err
		}
		str = strings.Replace(str, "i = \"\"", "i = \"" + smStr + "\"", 1)

		// Sleep for a ms
		time.Sleep(time.Millisecond)
	}

	// Write the file
	if err := os.WriteFile(filename, []byte(str), fs.ModePerm); err != nil {
		return err
	}
	return nil
}

func readFile(filename string) (*File, error) {
	// Read the file
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Unmarshal as TomlCard
	tomlCards := map[string]*TomlCard{}
	err = toml.Unmarshal(bytes, &tomlCards)
	if err != nil {
		return nil, err
	}

	// Convert to Card
	cards := []*Card{}
	for i := range tomlCards {
		tomlCard := tomlCards[i]
		sm := &Supermemo2{}
		if err = sm.Unmarshal(tomlCard.I); err != nil {
			return nil, err
		}
		front := tomlCard.Q
		if front == "" {
			front = i
		}
		c := &Card{
			front: front,
			back: tomlCard.A,
			info: tomlCard.I,
			sm: sm,
		}
		cards = append(cards, c)
	}

	// Return the file
	return &File{
		filename: filename,
		cards:    cards,
	}, nil
}

type TomlCard struct {
	A string `toml: "a"`
	Q string `toml: "q"`
	I string `toml: "i"`
}

type Card struct {
	front string
	back  string
	info  string
	sm    *Supermemo2
}

var letters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz-")

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}
