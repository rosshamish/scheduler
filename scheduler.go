package scheduler

import (
	"database/sql"
	"errors"
	"log"
	"sort"

	"github.com/kisielk/sqlstruct"
	_ "github.com/lib/pq"
)

type ScheduleRequest struct {
	Institution     string           `json:"institution"`
	Term            string           `json:"term"`
	Courses         []string         `json:"courses"`
	ElectivesGroups []ElectivesGroup `json:"electives,omitempty"`
	Preferences     *Preferences     `json:"preferences,omitempty"`
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
	MAX_SIMULTANEOUS_CANDIDATES := 10

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

	// Components with fewer options should be scheduled first
	sort.Sort(ByCount(components))

	candidates := []Schedule{}
	candidates = append(candidates, Schedule{})
	for i, component := range components {
		// each component is a list of sections
		sections := component

		prevLen := len(candidates)
		log.Printf("...Adding %v, finding schedules (%d/%d)", sections[0], i+1, len(components))

		choices := "choices"
		if len(sections) == 1 {
			choices = "choice"
		}
		log.Printf("...%d section %v", len(sections), choices)

		candidates = addComponent(candidates, sections, i)

		possibilities := prevLen * len(sections)
		var pct float64 = 0
		if possibilities != 0 {
			pct = float64(len(candidates)) / float64(possibilities) * 100.0
		}
		log.Printf("...Found %d from %d possibilities (%2.0f%%)\n", len(candidates), possibilities, pct)

		if len(candidates) > MAX_SIMULTANEOUS_CANDIDATES {
			log.Printf("...Keeping %d, killing worst %d\n", MAX_SIMULTANEOUS_CANDIDATES, len(candidates)-MAX_SIMULTANEOUS_CANDIDATES)
			candidates = candidates[:MAX_SIMULTANEOUS_CANDIDATES]
		}
		log.Printf("\n")
	}

	return candidates
}

func addComponent(candidates []Schedule, sections []Section, pace int) []Schedule {
	workReport := ""

	newCandidates := []Schedule{}
	for _, candidate := range candidates {
		if len(candidate.Sections) < pace {
			continue
		}
		for _, s := range sections {
			conflict := false
			for _, sCandidate := range candidate.Sections {
				if s.Conflicts(sCandidate) {
					conflict = true
					break
				}
			}
			if conflict {
				workReport = workReport + "x"
				continue
			}

			workReport = workReport + "O"
			newCandidates = append(newCandidates, candidate.addSection(s))
		}
	}
	log.Printf("new candidates %d", len(newCandidates))

	if len(workReport) > 60 {
		workReport = workReport[:60] + " (truncated)"
	}
	log.Printf("...%s", workReport)

	return newCandidates
}

type ByCount [][]Section

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return len(a[i]) < len(a[j]) }

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
				AND institution=$4`
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
			sections = append(sections, section)
		}
		if len(sections) == 0 {
			continue
		}
		components = append(components, sections)
	}

	return components, nil
}
