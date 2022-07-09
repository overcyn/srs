package main

import (
	"errors"
	"math"
	"time"
	"fmt"
)

// Supermemo2 calculates review intervals using SM2 algorithm
type Supermemo2 struct {
	nextReview time.Time
	repetition int
	interval   int
	easiness   float64
}

// NewSupermemo2 returns a new Supermemo2 instance
func NewSupermemo2() *Supermemo2 {
	return &Supermemo2{
		nextReview: time.Now(),
		repetition: 0,
		easiness:   2.5,
		interval:   0,
	}
}

// Advance advances supermemo state for a card.
func (sm *Supermemo2) Advance(rating float64) {
	if rating >= 3 {
		if sm.repetition == 0 {
			sm.interval = 1
		} else if sm.repetition == 2 {
			sm.interval = 6
		} else {
			sm.interval = int(math.Round(float64(sm.interval) * sm.easiness))
		}
		sm.repetition += 1
	} else {
		sm.repetition = 0
		sm.interval = 1
	}
	sm.easiness = sm.easiness + (0.1 - (5-rating)*(0.08+(5-rating)*0.02))
	if sm.easiness < 1.3 {
		sm.easiness = 1.3
	}
	sm.nextReview = time.Now().Add(time.Hour * time.Duration(24*sm.interval))
}

// MarshalJSON implements json.Marshaller for Supermemo2
func (sm *Supermemo2) Marshal() (string, error) {
	str := fmt.Sprintf("%.2f|%d✓|%dd|%s", sm.easiness, sm.repetition, sm.interval, sm.nextReview.Format("2006-01-02T15:04:05Z"))
	return str, nil
}

// UnmarshalJSON implements json.Unmarshaller for Supermemo2
func (sm *Supermemo2) Unmarshal(s string) error {
	var nextReviewStr string
	count, err := fmt.Sscanf(s, "%f|%d✓|%dd|%s", &sm.easiness, &sm.repetition, &sm.interval, &nextReviewStr)
	if err != nil {
		return err
	}
	if count != 4 {
		return errors.New("Invalid string")
	}

	nextReview, err := time.Parse("2006-01-02T15:04:05Z", nextReviewStr)
	if err != nil {
		return err
	}
	sm.nextReview = nextReview
	
	return nil
}
