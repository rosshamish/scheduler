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
	MAX_SIMULTANEOUS_CANDIDATES := 300

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
	empties := 0
	for i, component := range components {
		// each component is a list of sections
		sections := component

		// TODO fix existence of empty components
		prevLen := len(candidates)
		if len(component) == 0 {
			empties = empties + 1
			continue
		}
		log.Printf("...Scheduling %v (%d/%d)", sections[0], i+1, len(components)-empties)
		choices := "choices"
		if len(sections) == 1 {
			choices = "choice"
		}
		log.Printf("...%d section %v", len(sections), choices)
		candidates = addComponent(candidates, sections, i)
		log.Printf("...Done. %d candidates (%+d)\n", len(candidates), len(candidates)-prevLen)
		if len(candidates) > 200 {
			log.Printf("...Trimming. %d candidates (%+d)\n", MAX_SIMULTANEOUS_CANDIDATES, -1*(len(candidates)-MAX_SIMULTANEOUS_CANDIDATES))
			candidates = candidates[:MAX_SIMULTANEOUS_CANDIDATES]
		}
		log.Printf("\n")
	}

	return candidates
}

func addComponent(candidates []Schedule, sections []Section, pace int) []Schedule {
	log.Printf("...Finding candidates")
	workReport := ""

	newCandidates := []Schedule{}
	for _, candidate := range candidates {
		if len(candidate.Sections) < pace {
			continue
		}
		for _, s := range sections {
			for _, sCandidate := range candidate.Sections {
				if s.Conflicts(sCandidate) {
					workReport = workReport + "x"
					continue
				}
			}
			candidate = candidate.addSection(s)
			newCandidates = append(newCandidates, candidate)
			workReport = workReport + "O"
		}
	}

	if len(workReport) > 60 {
		workReport = workReport[:60] + " (truncated)"
	}
	log.Printf("...%s", workReport)

	return newCandidates
}

type ByCount [][]Section

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return len(a[i]) < len(a[i]) }

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
