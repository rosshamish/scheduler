package scheduler

import (
	"database/sql"
	"errors"
	"log"
	"math"

	"github.com/wkschwartz/pigosat"

	"github.com/kisielk/sqlstruct"
	_ "github.com/lib/pq"
)

type ScheduleRequest struct {
	Institution     string           `json:"institution"`
	Term            string           `json:"term"`
	Courses         []string         `json:"courses"`
	ElectivesGroups []ElectivesGroup `json:"electives,omitempty"`
	BusyTimes       []BusyTime       `json:"busy-times,omitempty"`
	Preferences     *Preferences     `json:"preferences,omitempty"`
}

type BusyTime struct {
	Days      Days     `json:"day"`
	StartTime AmPmTime `json:"startTime"`
	EndTime   AmPmTime `json:"endTime"`
}

func (s ScheduleRequest) GetBusyTimetableRanges() []TimetableRange {
	ttRanges := make([]TimetableRange, 0)
	for _, b := range s.BusyTimes {
		ttRanges = append(ttRanges, TimetableRangeFromTimes(b.Days, b.StartTime, b.EndTime))
	}
	return ttRanges
}

type ElectivesGroup struct {
	Courses []string `json:"courses"`
}

type Preferences struct {
	StartEarly    float32 `json:"start-early,omitempty"`
	NoMarathons   float32 `json:"no-marathons,omitempty"`
	CurrentStatus bool    `json:"current-status,omitempty"`
	ObeyStatus    bool    `json:"obey-status,omitempty"`
}

func Generate(req ScheduleRequest) []Schedule {
	components := [][]Section{}
	for _, course := range req.Courses {
		courseComponents, err := getComponents(course, req.Term, req.Institution)
		if err != nil {
			log.Fatal(err)
		}
		if len(courseComponents) == 0 {
			continue
		}
		components = append(components, courseComponents...)
	}

	opts := new(pigosat.Options)
	opts.PropagationLimit = uint64(math.Pow10(5))
	p, err := pigosat.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	var clauses []pigosat.Clause
	indexToSection, sectionToIndex := buildSectionIndex(components)

	/*
	 * Constraint: MUST schedule one of each component
	 */
	clauses = make([]pigosat.Clause, 0)
	clauseByC := make(map[string][]pigosat.Literal)
	for _, section := range indexToSection {
		// log.Printf("[%d] %s", sectionToIndex[section.String()], section.String())
		c := section.Course.String + section.Component.String
		clauseByC[c] = append(clauseByC[c], pigosat.Literal(sectionToIndex[section.String()]))
	}
	for _, clause := range clauseByC {
		clauses = append(clauses, pigosat.Clause(clause))
	}
	p.AddClauses(pigosat.Formula(clauses))
	// log.Printf("Clauses: %v\n", clauses)

	/*
	 * Constraint: MUST NOT schedule conflicting sections together
	 *         AND MUST NOT schedule sections during any busy time
	 *   Note: sections in the same component conflict.
	 *   Note: recall (A' + B') == (AB)'
	 */
	busyTimes := req.GetBusyTimetableRanges()
	log.Printf("%v\n", busyTimes)
	conflicts := getConflictsExt(components, busyTimes)
	clauses = make([]pigosat.Clause, 0)
	for _, conflict := range conflicts {
		var clause []pigosat.Literal
		if conflict.a.String() != conflict.b.String() {
			clause = []pigosat.Literal{
				pigosat.Literal(-1 * sectionToIndex[conflict.a.String()]),
				pigosat.Literal(-1 * sectionToIndex[conflict.b.String()]),
			}
		} else {
			clause = []pigosat.Literal{
				pigosat.Literal(-1 * sectionToIndex[conflict.a.String()])}
		}
		clauses = append(clauses, pigosat.Clause(clause))
	}
	p.AddClauses(pigosat.Formula(clauses))
	// log.Printf("Clauses: %v\n", clauses)

	count := 0
	schedules := make([]Schedule, 0)
	for status, solution := p.Solve(); status == pigosat.Satisfiable; status, solution = p.Solve() {
		solnSections := make([]Section, 0)
		for i, val := range solution {
			if val == false {
				continue
			}
			if section, ok := indexToSection[int32(i)]; ok {
				solnSections = append(solnSections, section)
			}
		}
		if count < 200 {
			schedules = append(schedules, Schedule{solnSections, []Conflict{}})
		}
		count = count + 1
	}
	log.Printf("Found %d schedules\n", count)

	return schedules
}

type Conflict struct {
	a, b Section
}

func (c Conflict) String() string {
	return c.a.String() + "x" + c.b.String()
}

func getConflicts(components [][]Section) []Conflict {
	conflicts := make([]Conflict, 0)
	for i, ac := range components {
		for j, bc := range components {
			if j < i {
				continue
			}
			// A component is a list of sections
			aSections, bSections := ac, bc
			// Conflict if:
			//   - time conflict
			//   - same component (i == j, here)
			//   - dependency not satisfied ie LEC A1->LAB A2,A3
			for ii, a := range aSections {
				for jj, b := range bSections {
					if ii <= jj {
						continue
					}
					if a.Conflicts(b) || ii == jj {
						conflicts = append(conflicts, Conflict{a, b})
					}
				}
			}
		}
	}
	return conflicts
}

func getConflictsExt(components [][]Section, busyTimes []TimetableRange) []Conflict {
	conflicts := getConflicts(components)
	for _, sections := range components {
		for _, section := range sections {
			if section.TimetableRange == nil {
				section.TimetableRange = TimetableRangeFrom(section)
			}
			for _, busyTime := range busyTimes {
				if section.TimetableRange.Overlaps(busyTime) {
					conflicts = append(conflicts, Conflict{section, section})
				}
			}
		}
	}
	return conflicts
}

func buildSectionIndex(components [][]Section) (map[int32]Section, map[string]int32) {
	indexToSection := make(map[int32]Section)
	sectionToIndex := make(map[string]int32)
	var idx int32 = 1
	for _, sections := range components {
		for _, section := range sections {
			indexToSection[idx] = section
			sectionToIndex[section.String()] = idx
			idx = idx + 1
		}
	}
	return indexToSection, sectionToIndex
}

var componentTypes = [...]string{"LEC", "LAB", "SEM", "LBL"}

// getComponents returns a course's sections in a given term at
// a given institution. The sections are values of a map whose
// keys are component strings ("LAB", "LEC", ...)
func getComponents(course, term, institution string) ([][]Section, error) {
	db, err := sql.Open("postgres", "postgres://localhost:5432/classtime?sslmode=disable")
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, errors.New("DB connection failed")
	}
	defer db.Close()

	var components [][]Section
	query := `SELECT * FROM section
			WHERE component=$1
				AND course=$2
				AND term=$3
				AND institution=$4
			ORDER BY course, component`
	courseAndSectionIdToSection := make(map[string]Section)
	for _, c := range componentTypes {
		rows, err := db.Query(query, c, course, term, institution)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		sections := []Section{}
		for rows.Next() {
			var section Section
			if err := sqlstruct.Scan(&section, rows); err != nil {
				log.Fatal(err)
				continue
			}
			courseAndSectionIdToSection[section.Course.String+section.Section.String] = section
			sections = append(sections, section)
		}
		if len(sections) == 0 {
			continue
		}
		components = append(components, sections)
	}

	for _, sections := range components {
		for _, section := range sections {
			if section.AutoEnroll.String != "" {
				autoEnrollSection, ok := courseAndSectionIdToSection[section.Course.String+section.AutoEnroll.String]
				if ok {
					section.AutoEnrollComponent.Scan(autoEnrollSection.Component.String)
				}
			}
		}
	}

	return components, nil
}
