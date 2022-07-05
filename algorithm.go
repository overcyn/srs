package main

import (
	"time"
	"math"
)

// https://github.com/ap4y/leaf/blob/master/supermemo2_plus.go
struct SM2PlusCustom {
	LastReview time.Time
	Difficulty     float64
	Interval       float64
}

// NextReviewAt returns next review timestamp for a card.
func (sm *SM2PlusCustom) NextReviewAt() time.Time {
	return sm.LastReview.Add(time.Duration(24*sm.Interval) * time.Hour)
}

// PercentOverdue returns corresponding SM2+ value for a Card.
func (sm *SM2PlusCustom) PercentOverdue() float64 {
	percentOverdue := time.Since(sm.LastReview).Hours() / float64(24*sm.Interval)
	return math.Min(2, percentOverdue)
}

// Advance advances supermemo state for a card.
func (sm *SM2PlusCustom) Advance(rating float64) float64 {
	success := rating >= ratingSuccess
	percentOverdue := float64(1)
	if success {
		percentOverdue = sm.PercentOverdue()
	}

	sm.Difficulty += percentOverdue / 35 * (8 - 9*rating)
	sm.Difficulty = math.Max(0, math.Min(1, sm.Difficulty))
	difficultyWeight := 3.5 - 1.7*sm.Difficulty

	minInterval := math.Min(1.0, sm.Interval)
	factor := minInterval / math.Pow(difficultyWeight, 2)
	if success {
		minInterval = 0.2
		factor = minInterval + (difficultyWeight-1)*percentOverdue
	}

	sm.LastReview = time.Now()
	sm.Interval = math.Max(minInterval, math.Min(sm.Interval*factor, 300))
	return sm.Interval
}