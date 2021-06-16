package kvserver

import (
	"fmt"
	"regexp"
	"strings"
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

type ImpCount struct {
	Total int
	Today int
	Slot  int
}

type LineItem struct {
	ID        int       `json:"id,omitempty"`
	Type      int       `json:"type,omitempty"`
	Price     float64   `json:"price,omitempty"`
	Source    string    `json:"source,omitempty"`
	FCap      string    `json:"fcap,omitempty"`
	StartDate time.Time `json:"startdate,omitempty"`
	EndDate   time.Time `json:"enddate,omitempty"`
	Goal      int       `json:"goal,omitempty"`
	Device    string    `json:"device,omitempty"`
	OS        string    `json:"os,omitempty"`
	IG        string    `json:"ig,omitempty"`
	RegExKey  string    `json:"targetings,omitempty"`
	Pacing    int       `json:"pacing,omitempty"`
	Creatives []int     `json:"creatives,omitempty"`

	//pacing calculations
	RegExpression []*regexp.Regexp `json:"-"`
	CurrentDate   time.Time        `json:"-"`
	DailyGoal     int              `json:"-"`
	Impressions   ImpCount         `json:"-"`
	Winning       ImpCount         `json:"-"`
	Status        LineItemStatus   `json:"-"`
}

func NewLineItem(
	id int,
	litype int,
	price float64,
	source string,
	device string,
	os string,
	ig string,
	fcap string,
	startdate string,
	enddate string,
	goal int) (*LineItem, error) {
	obj := &LineItem{
		ID:          id,
		Type:        litype,
		Price:       price,
		Source:      source,
		Device:      device,
		OS:          os,
		IG:          ig,
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

	obj.formTargetingKey()
	obj.getPacingRate(1)
	obj.updateStatus()

	return obj, nil
}

func (l *LineItem) updateStatus() {
	l.Status = Delivering
	now := time.Now().UTC().Unix()
	if l.StartDate.Unix() > now ||
		now > l.EndDate.Unix() {
		l.Status = InActive
	}

	//lifetime goal
	if l.Winning.Total >= l.Goal {
		l.Status = Completed
	}

	//today's goal
	if l.Winning.Today >= l.DailyGoal {
		l.Status = PartialCompleted
	}
}

func (l *LineItem) formTargetingKey() {
	//(samsung|mi|,):(apple|android|,):(sports|music|,)+
	//(samsung|mi|,):(.*):(sports|music|,)+
	keys := []string{
		getkey(l.Device),
		getkey(l.OS),
		getkey(l.IG) + "+"}
	l.RegExKey = strings.Join(keys, ":")

	l.RegExpression = make([]*regexp.Regexp, len(keys))
	for index, key := range keys {
		l.RegExpression[index], _ = regexp.Compile(key)
	}
}

func getkey(value string) string {
	if len(value) == 0 {
		return "(.*)"
	}
	return fmt.Sprintf("(%s)", strings.Replace(value, ",", "|", -1))
}

func (l *LineItem) calculateGoal() {
	daysRemaining := int(l.EndDate.Sub(l.CurrentDate).Hours()/24) + 1
	remainingGoal := l.Goal - l.Impressions.Total
	l.DailyGoal = remainingGoal / daysRemaining
}

func (l *LineItem) caluclatePacingRate() {
	totalDiff := l.EndDate.Sub(l.CurrentDate)
	daysRemaining := int(totalDiff.Hours()/24) + 1
	todaySlotsRemaining := 0

	if daysRemaining == 1 {
		todaySlotsRemaining = int(totalDiff.Minutes()/15) + 1
	} else {
		startTime := time.Date(l.CurrentDate.Year(), l.CurrentDate.Month(), l.CurrentDate.Day(), 0, 0, 0, 0, l.CurrentDate.Location())
		todaySlotsRemaining = int(l.CurrentDate.Sub(startTime).Minutes()/15) + 1
	}

	l.calculateGoal()
	slotGoal := (l.DailyGoal - l.Impressions.Today) / todaySlotsRemaining

	l.Pacing = 1
	if slotGoal > 0 && l.Impressions.Slot > 0 {
		l.Pacing = int(l.Impressions.Slot / slotGoal)
	}

	if l.Pacing <= 0 {
		l.Pacing = 1
	}
}

func (l *LineItem) getPacingRate(csigSlotImpression int) int {
	totalDiff := l.EndDate.Sub(l.CurrentDate)
	daysRemaining := int(totalDiff.Hours()/24) + 1
	todaySlotsRemaining := 0

	if daysRemaining == 1 {
		todaySlotsRemaining = int(totalDiff.Minutes()/15) + 1
	} else {
		startTime := time.Date(l.CurrentDate.Year(), l.CurrentDate.Month(), l.CurrentDate.Day(), 0, 0, 0, 0, l.CurrentDate.Location())
		todaySlotsRemaining = int(l.CurrentDate.Sub(startTime).Minutes()/15) + 1
	}

	l.calculateGoal()
	slotGoal := (l.DailyGoal - l.Impressions.Today) / todaySlotsRemaining
	pacingrate := 1
	if slotGoal > 0 && csigSlotImpression > 0 {
		pacingrate = int(csigSlotImpression / slotGoal)
	}

	if pacingrate <= 0 {
		pacingrate = 1
	}
	return pacingrate
}

func (l *LineItem) updateImpressions(impCount int, winningCount int) {
	now := time.Now().UTC()
	if l.CurrentDate.Day() == now.Day() &&
		l.CurrentDate.Month() == now.Month() &&
		l.CurrentDate.Year() == now.Year() {

		l.Impressions.Today = l.Impressions.Today + impCount
		l.Winning.Today = l.Winning.Today + winningCount
	} else {
		//new day
		l.calculateGoal()
		l.Impressions.Today = impCount
	}

	l.Impressions.Total = l.Impressions.Total + impCount
	l.Winning.Total = l.Winning.Total + winningCount

	if slotNumber(&now) == slotNumber(&l.CurrentDate) {
		l.Impressions.Slot = l.Impressions.Slot + impCount
		l.Winning.Slot = l.Winning.Slot + winningCount
	} else {
		l.Impressions.Slot = impCount
		l.Winning.Slot = winningCount
	}

	l.CurrentDate = now
	l.updateStatus()
}

func slotNumber(date *time.Time) int {
	return int((date.Hour()*60+date.Minute())/15) + 1
}
