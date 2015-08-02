package scheduler

import (
	"encoding/json"
	"log"
	"regexp"
	"testing"
)

var testCases = [...]string{
	// {  # Fall 2015 ECE 304
	// -- has 3 components and has a dependency
	// github.com/rosshamish/classtime/issues/98
	`{
		"institution": "ualberta",
		"term": "1530",
		"courses": ["105005"]
	}`,
	// 1st year engineering Fall Term 2014
	`{
		"institution": "ualberta",
		"term": "1490",
		"courses": [
		    "001343",
		    "004093",
		    "004096",
		    "006768",
		    "009019"
		],
		"electives": [
		    {
		        "courses": [
		            "000268",
		            "000269",
		            "000270"
		        ]
		    }
		],
		"preferences": {
		    "start-early": -10,
		    "no-marathons": -10,
		    "current-status": true,
		    "obey-status": true
		}
	}`,
}

func removeWhitespace(s string) string {
	re, _ := regexp.Compile(`[\s\t\r\n]`)
	s = string(re.ReplaceAll([]byte(s), []byte("")))
	return s
}

func TestJSONMarshalling(t *testing.T) {
	for _, testCase := range testCases {
		schedReq := new(ScheduleRequest)

		json.Unmarshal([]byte(testCase), schedReq)
		ret, _ := json.Marshal(schedReq)

		had := removeWhitespace(testCase)
		got := removeWhitespace(string(ret))
		if had != got {
			t.Errorf("had: \n%q\ngot: \n%q", had, got)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	log.Printf("--")
	for i, testCase := range testCases {
		log.Printf("Test case #%d", i)
		schedReq := new(ScheduleRequest)

		json.Unmarshal([]byte(testCase), schedReq)
		log.Printf("Schedule request: %+v", *schedReq)
		schedules := Generate(*schedReq)
		log.Printf("Generated %d schedules", len(schedules))
		if len(schedules) > 0 {
			log.Printf("First schedule: \n%q", schedules[0])
		}
		log.Printf("------------")
	}
}

func TestGenerate(t *testing.T) {
	log.Printf("------------")
	for i, testCase := range testCases {
		log.Printf("Test case #%d", i)
		schedReq := new(ScheduleRequest)

		json.Unmarshal([]byte(testCase), schedReq)
		log.Printf("Schedule request: %+v", *schedReq)
		schedules := Generate(*schedReq)
		log.Printf("Generated %d schedules", len(schedules))
		if len(schedules) > 0 {
			log.Printf("First schedule: \n%q", schedules[0])
		}
		log.Printf("------------")
	}
}
