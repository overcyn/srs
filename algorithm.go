package main

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

// Supermemo2 calculates review intervals using SM2 algorithm
type Supermemo2 struct {
	NextReview time.Time
	Repetition int
	Interval   int
	Easiness   float64
}

// NewSupermemo2 returns a new Supermemo2 instance
func NewSupermemo2() *Supermemo2 {
	return &Supermemo2{
		NextReview: time.Now(),
		Repetition: 0,
		Easiness:   2.5,
		Interval:   0,
	}
}

// Advance advances supermemo state for a card.
func (sm *Supermemo2) Advance(rating float64) {
	if rating >= 3 {
		if sm.Repetition == 0 {
			sm.Interval = 1
		} else if sm.Repetition == 2 {
			sm.Interval = 2
		} else {
			sm.Interval = int(math.Round(float64(sm.Interval) * sm.Easiness))
		}
		sm.Repetition += 1
	} else {
		sm.Repetition = 0
		sm.Interval = 1
	}
	sm.Easiness = sm.Easiness + (0.1 - (5-rating)*(0.08+(5-rating)*0.02))
	if sm.Easiness < 1.3 {
		sm.Easiness = 1.3
	}
	sm.NextReview = time.Now().Add(time.Hour * time.Duration(24*sm.Interval))
}

// MarshalJSON implements json.Marshaller for Supermemo2
func (sm *Supermemo2) Marshal() (string, error) {
	return sm.NextReview.Format("2006-01-02T15:04:05Z") + "|" + strconv.Itoa(sm.Repetition) + "|" + strconv.Itoa(sm.Interval) + "|" + strconv.Itoa(int(sm.Easiness*100)), nil
}

// UnmarshalJSON implements json.Unmarshaller for Supermemo2
func (sm *Supermemo2) Unmarshal(s string) error {
	cs := strings.Split(s, "|")
	if len(cs) != 4 {
		return errors.New("Unexpected number of commas")
	}

	nextReview, err := time.Parse("2006-01-02T15:04:05Z", cs[0])
	if err != nil {
		return err
	}
	sm.NextReview = nextReview

	repetition, err := strconv.Atoi(cs[1])
	if err != nil {
		return err
	}
	sm.Repetition = repetition

	interval, err := strconv.Atoi(cs[2])
	if err != nil {
		return err
	}
	sm.Interval = interval

	easiness, err := strconv.Atoi(cs[2])
	if err != nil {
		return err
	}
	sm.Easiness = float64(easiness) / 100
	return nil
}
