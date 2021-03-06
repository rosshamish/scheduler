package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
)

type Schedule struct {
	Sections  []Section `json:"sections"`
	Conflicts []Conflict
}

func (sch Schedule) String() string {
	repr := "["
	for i, section := range sch.Sections {
		if i > 0 {
			repr += ", "
		}
		repr += fmt.Sprintf("%v", section)
	}
	repr += "]"
	return repr
}

func (sch Schedule) addSection(sec Section) Schedule {
	sch.Sections = append(sch.Sections, sec)
	return sch
}

type ByNumConflicts []Schedule

func (a ByNumConflicts) Len() int           { return len(a) }
func (a ByNumConflicts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNumConflicts) Less(i, j int) bool { return len(a[i].Conflicts) < len(a[i].Conflicts) }

type Section struct {
	AsString            sql.NullString `json:"asString"`
	AutoEnroll          sql.NullString `json:"autoEnroll"`
	AutoEnrollComponent sql.NullString `json:"autoEnrollComponent"`
	Career              sql.NullString `json:"career"`
	Catalog             int            `json:"catalog"`
	Course              sql.NullString `json:"course"`
	CourseDescription   sql.NullString `json:"courseDescription"`
	CourseTitle         sql.NullString `json:"courseTitle"`
	Department          sql.NullString `json:"department"`
	DepartmentCode      sql.NullString `json:"departmentCode"`
	Faculty             sql.NullString `json:"faculty"`
	FacultyCode         sql.NullString `json:"facultyCode"`
	Subject             sql.NullString `json:"subject"`
	SubjectTitle        sql.NullString `json:"subjectTitle"`
	Term                sql.NullString `json:"term"`
	Units               int            `json:"units"`
	Class               sql.NullString `json:"class_"`
	Component           sql.NullString `json:"component"`
	Day                 sql.NullString `json:"day"`
	StartTime           sql.NullString `json:"startTime"`
	EndTime             sql.NullString `json:"endTime"`
	Section             sql.NullString `json:"section"`
	Campus              sql.NullString `json:"campus"`
	Capacity            int            `json:"capacity"`
	InstructorUid       sql.NullString `json:"instructorUid"`
	Location            sql.NullString `json:"location"`
	TimetableRange      TimetableRange `json:"-"`
}

func (s Section) String() string {
	return "<" + s.AsString.String + ">"
}

func (s Section) Conflicts(o Section) bool {
	if s.isSameCourseAndComponent(o) {
		return true
	}

	if s.hasTimeConflict(o) {
		return true
	}

	if s.hasDependencyConflict(o) {
		return true
	}

	return false
}

func (s Section) isSameCourseAndComponent(o Section) bool {
	if s.Course.String == o.Course.String {
		if s.Component.String == o.Component.String {
			return true
		}
	}
	return false
}

func (s Section) hasTimeConflict(o Section) bool {
	if s.TimetableRange == nil {
		s.TimetableRange = TimetableRangeFrom(s)
	}
	if o.TimetableRange == nil {
		o.TimetableRange = TimetableRangeFrom(o)
	}
	return s.TimetableRange.Overlaps(o.TimetableRange)
}

func (s Section) hasDependencyConflict(o Section) bool {
	if s.Course.String != o.Course.String {
		// They're not even the same course. No conflict
		return false
	}
	if s.AutoEnroll.String == "" && o.AutoEnroll.String == "" {
		// Neither section has a dependency! No conflict
		return false
	}
	if s.Section.String != o.AutoEnroll.String &&
		s.AutoEnroll.String != o.Section.String {
		// Neither section depends on the other section's identifier. No conflict.
		return false
	}
	if s.Section.String == o.AutoEnroll.String &&
		s.Component.String == o.AutoEnrollComponent.String {
		// Dependency exists, and is satisfied, from o->s. No conflict
		return false
	} else if s.AutoEnroll.String == o.Section.String &&
		s.AutoEnrollComponent.String == o.Component.String {
		// Dependency exists, and is satisfied, from s->o. No conflict
		return false
	}

	// There exists a dependency between s-o, and it is not satisfied. Conflict.
	return true
}

// type TimetableRange maps a Day to an availability bitmap where
// 0=free, 1=busy.
type TimetableRange map[Day]uint64

type Day string
type Days string

func (tr TimetableRange) Overlaps(otr TimetableRange) bool {
	for day, blocks := range tr {
		if blocks&otr[day] != 0 {
			return true
		}
	}
	return false
}

func TimetableRangeFromTimes(days Days, a AmPmTime, b AmPmTime) TimetableRange {
	startBlock := a.AsTimetableBlockNum()
	endBlock := b.AsTimetableBlockNum()
	tr := make(TimetableRange)

	for _, day := range days {
		for i := startBlock; i <= endBlock; i++ {
			tr[Day(day)] |= 1 << i
		}
	}
	return tr
}

func TimetableRangeFrom(s Section) TimetableRange {
	return TimetableRangeFromTimes(Days(s.Day.String), AmPmTime(s.StartTime.String), AmPmTime(s.EndTime.String))
}

type AmPmTime string

var AmPmRegex string = `(?P<hour>\d\d):(?P<minute>\d\d) (?P<ampm>[AP]M)`

func (time AmPmTime) AsTimetableBlockNum() uint64 {
	re, _ := regexp.Compile(AmPmRegex)

	matches := re.FindStringSubmatch(string(time))
	if matches == nil {
		log.Fatalf("Time '%s' does not match regex '%s'", string(time), AmPmRegex)
	}

	hour, _ := strconv.ParseUint(matches[1], 10, 0)
	minute, _ := strconv.ParseUint(matches[2], 10, 0)
	var amPmOffset uint64 = 0
	if hour != 12 && matches[3] == "PM" {
		amPmOffset = 12
	}

	blockNum := (hour+amPmOffset)*2 + minute/30
	return blockNum
}
