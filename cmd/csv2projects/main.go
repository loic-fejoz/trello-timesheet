package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// type taskEntry struct {
// 	datetime      time.Time
// 	durationInDay float64
// 	label         string
// 	tags          []string
// }

// MonthEntries encapsulates a map from year-month to days counter
type MonthEntries map[string]float64

// Summary is a map from name of project to MonthEntries
type Summary map[string]MonthEntries

// Add some effort at given year-month
func (me *MonthEntries) Add(year int, month time.Month, effort float64) {
	yearmonth := fmt.Sprintf("%d-%02d", year, month)
	e := (*me)[yearmonth]
	(*me)[yearmonth] = e + effort
}

// Collect all year-month enties in a record
func (me MonthEntries) Collect(initial []string, allYearMonths []string) []string {
	record := initial
	for _, yearmonth := range allYearMonths {
		entry := me[yearmonth]
		record = append(record, fmt.Sprintf("%.02f", entry))
	}
	return record
}

// Get month entries for project with given name
func (s *Summary) Get(projectname string) *MonthEntries {
	me := (*s)[projectname]
	if me == nil {
		me = make(MonthEntries)
	}
	(*s)[projectname] = me
	return &me
}

// Append a CSV record to a Summary
func (s *Summary) Append(record []string) {
	datetime, err := time.Parse("2006-01-02T", record[0]+"T")
	if err != nil {
		log.Print("Skip an entry because of malformed date:" + record[0])
		return
	}
	projectname := "divers"
	if len(record[3]) > 0 {
		tags := strings.Split(record[3], ",")
		if len(tags) > 0 {
			for _, prj := range tags {
				projectname = prj
				break
			}
		}
	}
	effort, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		log.Print("Skip an entry because of malformed effort, not a double:" + record[1])
		return
	}
	s.Get(projectname).Add(datetime.Year(), datetime.Month(), effort)
}

func (s *Summary) AllMonths() []string {
	allmonths := make(map[string]string)
	for _, me := range *s {
		for yearmonth, _ := range me {
			allmonths[yearmonth] = yearmonth
		}
	}
	keys := make([]string, 0, len(allmonths))
	for i, _ := range allmonths {
		keys = append(keys, i)
	}
	sort.Strings(keys)
	return keys
}

// Load from given file
func (s *Summary) Load(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if len(record) < 4 {
			log.Print("Skip a malformed entry", record)
			continue
		}
		s.Append(record)
	}
}

// WriteCSV dump summary content to given file
func (s Summary) WriteCSV(output *os.File) {
	o := csv.NewWriter(output)
	months := s.AllMonths()
	headers := append([]string{"Project"}, months...)
	o.Write(headers)
	for projectname, me := range s {
		record := []string{projectname}
		record = me.Collect(record, months)
		o.Write(record)
	}
	o.Flush()
}

var filename = flag.String("filename", "", "input file")
var outputfilename = flag.String("output", "", "output file")

func main() {
	flag.Parse()

	timesheet := make(Summary)
	timesheet.Load(*filename)
	var output *os.File
	if len(*outputfilename) > 0 {
		o, err := os.Create(*outputfilename)
		if err != nil {
			log.Panic(err)
		}
		output = o
		defer output.Close()
	} else {
		output = os.Stdout
	}
	timesheet.WriteCSV(output)
}
