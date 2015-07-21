package scheduler

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Schedule struct {
	Sections []Section `json:"sections"`
}

type Section struct {
	AsString          string   `json:"asString"`
	Career            string   `json:"career"`
	Catalog           int      `json:"catalog"`
	Course            string   `json:"course"`
	CourseDescription string   `json:"courseDescription"`
	CourseTitle       string   `json:"courseTitle"`
	Department        string   `json:"department"`
	DepartmentCode    string   `json:"departmentCode"`
	Faculty           string   `json:"faculty"`
	FacultyCode       string   `json:"facultyCode"`
	Subject           string   `json:"subject"`
	SubjectTitle      string   `json:"subjectTitle"`
	Term              string   `json:"term"`
	Units             int      `json:"units"`
	Class             string   `json:"class_"`
	Component         string   `json:"component"`
	Day               Day      `json:"day"`
	StartTime         AmPmTime `json:"startTime"`
	EndTime           AmPmTime `json:"endTime"`
	Section           string   `json:"section"`
	Campus            string   `json:"campus"`
	Capacity          int      `json:"capacity"`
	InstructorUid     string   `json:"instructorUid"`
	Location          string   `json:"location"`
}

type Day string

var Days string = "MWTRF"

func (day Day) AsDayNum() (dayNum int, err error) {
	dayNum = strings.Index(Days, string(day))
	if dayNum == -1 {
		return -1, errors.New(fmt.Sprintf("Day \"%q\" is not a valid Day", day))
	}

	return dayNum, nil
}

type AmPmTime string

var AmPmRegex string = `(?P<hour>\d\d):(?P<minute>\d\d) (?P<ampm>[AP]M)`

func (time AmPmTime) AsTimetableBlockNum() (blockNum int, err error) {
	re, _ := regexp.Compile(AmPmRegex)

	matches := re.FindStringSubmatch(string(time))
	if matches == nil {
		return -1, errors.New(fmt.Sprintf("Time \"%q\" does not match regex %q", time, amPmRegex))
	}

	hour, _ := strconv.Atoi(matches[1])
	minute, _ := strconv.Atoi(matches[2])
	amPmOffset := 0
	if hour != 12 && matches[3] == "PM" {
		amPmOffset = 12
	}

	blockNum = (hour+amPmOffset)*2 + minute/30
	return blockNum, nil
}
