package kvserver

import (
	"fmt"
	"time"
)

const (
	DateTimeLayout = "2006-01-02 15:04:05"
)

type LineItemStatus int

const (
	InActive LineItemStatus = iota
	Delivering
	PartialCompleted
	Completed
)

type LineItem struct {
	ID        int       `json:"id,omitempty"`
	Type      int       `json:"type,omitempty"`
	Price     float64   `json:"price,omitempty"`
	FCap      string    `json:"fcap,omitempty"`
	StartDate time.Time `json:"startdate,omitempty"`
	EndDate   time.Time `json:"enddate,omitempty"`
	Goal      int       `json:"goal,omitempty"`
	//pacing calculations
	CurrentDate           time.Time      `json:"-"`
	DailyGoal             int            `json:"-"`
	TotalImpressionServed int            `json:"-"`
	TodayImpressionServed int            `json:"-"`
	Status                LineItemStatus `json:"-"`
}

func NewLineItem(
	id int,
	litype int,
	price float64,
	fcap string,
	startdate string,
	enddate string,
	goal int) (*LineItem, error) {
	obj := &LineItem{
		ID:          id,
		Type:        litype,
		Price:       price,
		FCap:        fcap,
		Goal:        goal,
		CurrentDate: time.Now(),
	}

	if id <= 0 {
		return nil, fmt.Errorf("invalid lineitem id")
	}

	if price <= 0 {
		return nil, fmt.Errorf("invalid lineitem price value")
	}

	if goal <= 0 {
		return nil, fmt.Errorf("invalid lineitem goal")
	}

	var err error
	if obj.StartDate, err = time.ParseInLocation(DateTimeLayout, startdate, time.UTC); err != nil {
		return nil, err
	}

	if obj.EndDate, err = time.ParseInLocation(DateTimeLayout, enddate, time.UTC); err != nil {
		return nil, err
	}

	if obj.StartDate.Unix() >= obj.EndDate.Unix() {
		return nil, fmt.Errorf("invalid lineitem startdate > enddate")
	}

	obj.calculateGoal()
	obj.updateStatus()

	return obj, nil
}

func (l *LineItem) UpdateImpressions(impCount int) {
	now := time.Now().UTC()
	if l.CurrentDate.Day() == now.Day() &&
		l.CurrentDate.Month() == now.Month() &&
		l.CurrentDate.Year() == now.Year() {
		l.TodayImpressionServed = l.TodayImpressionServed + impCount
	} else {
		//new day
		l.calculateGoal()
		l.TodayImpressionServed = impCount
	}

	l.TotalImpressionServed = l.TotalImpressionServed + impCount
	l.CurrentDate = now
	l.updateStatus()
}

func (l *LineItem) updateStatus() {
	l.Status = Delivering
	now := time.Now().UTC().Unix()
	if l.StartDate.Unix() > now ||
		now > l.EndDate.Unix() {
		l.Status = InActive
	}

	//lifetime goal
	if l.TotalImpressionServed >= l.Goal {
		l.Status = Completed
	}

	//today's goal
	if l.TodayImpressionServed >= l.DailyGoal {
		l.Status = PartialCompleted
	}
}

func (l *LineItem) calculateGoal() {
	daysRemaining := int(l.EndDate.Sub(l.CurrentDate).Hours()/24) + 1
	remainingGoal := l.Goal - l.TotalImpressionServed
	l.DailyGoal = remainingGoal / daysRemaining
}

func (l *LineItem) GetPacingRate(csigSlotImpression int) int {
	totalDiff := l.EndDate.Sub(l.CurrentDate)
	daysRemaining := int(totalDiff.Hours()/24) + 1
	todaySlotsRemaining := 0

	if daysRemaining == 1 {
		todaySlotsRemaining = int(totalDiff.Minutes()/15) + 1
	} else {
		startTime := time.Date(l.CurrentDate.Year(), l.CurrentDate.Month(), l.CurrentDate.Day(), 0, 0, 0, 0, l.CurrentDate.Location())
		todaySlotsRemaining = int(l.CurrentDate.Sub(startTime).Minutes()/15) + 1
	}

	slotGoal := (l.DailyGoal - l.TodayImpressionServed) / todaySlotsRemaining
	pacingrate := 1
	if slotGoal > 0 && csigSlotImpression > 0 {
		pacingrate = int(csigSlotImpression / slotGoal)
	}

	if pacingrate <= 0 {
		pacingrate = 1
	}
	return pacingrate
}


