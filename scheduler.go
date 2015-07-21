package scheduler

import ()

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

func GenerateSchedule(req ScheduleRequest) Schedule {
	return Schedule{}
}
