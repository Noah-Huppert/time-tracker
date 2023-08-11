package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	golangLog "log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TimeEntry records a period of time when work was completed
type TimeEntry struct {
	/// StartTime is the date and time when the period started
	StartTime time.Time

	// EndTime is the date and time when the period ended
	EndTime time.Time

	// Command is an optional comment explaining what work was completed during the period
	Comment string
}

// Hash returns a checksum of the contents of TimeEntry, can be used to find duplicate TimeEntry structs
func (e TimeEntry) Hash() (string, error) {
	hash, err := hashstructure.Hash(e, hashstructure.FormatV2, nil)
	if err != nil {
		return "", fmt.Errorf("failed to hash structure: %s", err)
	}

	return fmt.Sprintf("%d", hash), nil
}

func (e TimeEntry) Duration() time.Duration {
	return e.EndTime.Sub(e.StartTime)
}

type BillingPeriod struct {
	StartTime time.Time
	EndTime   time.Time
	Entries   []TimeEntry
}

func (p BillingPeriod) Duration() time.Duration {
	var total time.Duration = 0

	for _, entry := range p.Entries {
		total += entry.Duration()
	}

	return total
}

// InputTimeFormat represents YYYY-MM-DD HH:MM:SS
const InputTimeFormat = "2006-01-02 15:04:05 MST"

// FilePathTimeFormat is the format used in output file names
const FilePathTimeFormat = "2006-01-02"

// PeriodWeekly is a week duration
const PeriodWeekly = time.Hour * 24 * 7

// PeriodBiWeekly is a 2 week duration
const PeriodBiWeekly = PeriodWeekly * 2

// PeriodMonthly is a month duration
const PeriodMonthly = PeriodBiWeekly * 2

// PeriodWeeklyStr is the value a user must type to refer to PeriodWeekly
const PeriodWeeklyStr = "weekly"

// PeriodBiWeeklyStr is the value a user must type to refer to PeriodBiWeekly
const PeriodBiWeeklyStr = "bi-weekly"

// PeriodMonthlyStr is the value a user must type to refer to PeriodMonthly
const PeriodMonthlyStr = "monthly"

// ValidPeriodStrs is a list of all valid values a user could type to refer to a period
var ValidPeriodStrs = []string{
	PeriodWeeklyStr,
	PeriodBiWeeklyStr,
	PeriodMonthlyStr,
}

// ValidPeriodStrsJoined is a user friendly string listing values from ValidPeriodStrs
var ValidPeriodStrJoined = strings.Join(ValidPeriodStrs, ", ")

// FormatOutputDuration formats a duration for output like: HH:MM:SS
func FormatOutputDuration(dur time.Duration) string {
	hours := int(math.Floor(dur.Hours()))
	dur -= time.Hour * time.Duration(hours)

	minutes := int(math.Floor(dur.Minutes()))
	dur -= time.Minute * time.Duration(minutes)

	seconds := int(math.Floor(dur.Seconds()))

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// ParsePeriod takes a str and converts it to a duration. See Period... constants
func ParsePeriod(str string) (time.Duration, error) {
	if str == PeriodWeeklyStr {
		return PeriodWeekly, nil
	} else if str == PeriodBiWeeklyStr {
		return PeriodBiWeekly, nil
	} else if str == PeriodMonthlyStr {
		return PeriodMonthly, nil
	}

	return time.Hour * 0, fmt.Errorf("invalid period '%s', valid values: %s", str, ValidPeriodStrJoined)
}

const OutputModesHelpStr = "print (Print to console), dir=<DIR> (Output CSVs to <DIR> directory)"

type OutputMode interface {
	Output(periods []BillingPeriod, csvCols TimeTrackerReportCSVColumns) error
}

type TimeTrackerReportCSVColumns struct {
	ColumnStartTime string
	ColumnEndTime   string
	ColumnDuration  string
	ColumnComment   string
}

type PrintOutputMode struct{}

func (m PrintOutputMode) Output(periods []BillingPeriod, csvCols TimeTrackerReportCSVColumns) error {
	for _, period := range periods {
		fmt.Printf("\nPeriod: %s - %s - %s\n", period.StartTime, period.EndTime, period.Duration())
		fmt.Println("============")

		for _, entry := range period.Entries {
			comment := ""
			if len(entry.Comment) > 0 {
				comment = fmt.Sprintf(" (%s)", entry.Comment)
			}

			fmt.Printf("%s - %s%s - %s\n", entry.StartTime, entry.EndTime, comment, entry.Duration())
		}
	}

	return nil
}

type DirOutputMode struct {
	Dir string
}

func (m DirOutputMode) writePeriodReport(period BillingPeriod, csvCols TimeTrackerReportCSVColumns) error {
	// Open file
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %s", err)
	}

	outFile := filepath.Join(wd, m.Dir, fmt.Sprintf("bill-%s-%s.csv", period.StartTime.Format(FilePathTimeFormat), period.EndTime.Format(FilePathTimeFormat)))

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return fmt.Errorf("failed to open report file '%s': %s", outFile, err)
	}
	writer := csv.NewWriter(f)

	defer writer.Flush()

	// Write header
	err = writer.Write([]string{
		csvCols.ColumnStartTime,
		csvCols.ColumnEndTime,
		csvCols.ColumnDuration,
		csvCols.ColumnComment,
	})
	if err != nil {
		return fmt.Errorf("failed to write headers to report file '%s': %s", outFile, err)
	}

	// Write entries
	for entryI, entry := range period.Entries {
		err = writer.Write([]string{
			entry.StartTime.String(),
			entry.EndTime.String(),
			FormatOutputDuration(entry.Duration()),
			entry.Comment,
		})
		if err != nil {
			return fmt.Errorf("failed to write entry %d to report file '%s': %s", entryI, outFile, err)
		}
	}

	return nil
}

func (m DirOutputMode) writeRollup(periods []BillingPeriod, csvCols TimeTrackerReportCSVColumns) error {
	// Open file
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %s", err)
	}

	outFile := filepath.Join(wd, m.Dir, "billing-periods.csv")

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return fmt.Errorf("failed to open report file '%s': %s", outFile, err)
	}
	writer := csv.NewWriter(f)

	defer writer.Flush()

	// Write header
	err = writer.Write([]string{
		csvCols.ColumnStartTime,
		csvCols.ColumnEndTime,
		csvCols.ColumnDuration,
	})
	if err != nil {
		return fmt.Errorf("failed to write headers to report file '%s': %s", outFile, err)
	}

	for _, period := range periods {
		err = writer.Write([]string{
			period.StartTime.String(),
			period.EndTime.String(),
			FormatOutputDuration(period.Duration()),
		})
		if err != nil {
			return fmt.Errorf("failed to write period %s-%s to report file '%s': %s", period.StartTime.String(), period.EndTime.String(), outFile, err)
		}
	}

	return nil
}

func (m DirOutputMode) Output(periods []BillingPeriod, csvCols TimeTrackerReportCSVColumns) error {
	// Record billing periods
	for _, period := range periods {
		if err := m.writePeriodReport(period, csvCols); err != nil {
			return err
		}
	}

	// Record rollup of periods
	if err := m.writeRollup(periods, csvCols); err != nil {
		return err
	}

	return nil
}

func ParseOutputMode(str string) (OutputMode, error) {
	dirRegex, err := regexp.Compile("^dir=(.*)$")
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex to detect dir mode: %s", err)
	}

	if str == "print" {
		return PrintOutputMode{}, nil
	} else if matches := dirRegex.FindStringSubmatch(str); len(matches) > 0 {
		return DirOutputMode{
			Dir: matches[1],
		}, nil
	}

	return nil, fmt.Errorf("string '%s' did not match any output modes, valid output modes: %s", str, OutputModesHelpStr)
}

type TimeTracker struct {
	logger *zap.Logger
}

type TimeTrackerReportOpts struct {
	InDir           string
	BillingPeriod   time.Duration
	Timezone        string
	ColumnStartTime string
	ColumnEndTime   string
	ColumnComment   string
	OutputMode      OutputMode
}

func NewTimeTracker() *TimeTracker {
	return &TimeTracker{}
}

func (t *TimeTracker) Init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		golangLog.Fatalf("failed to initialize logger: %s\n", err)
	}

	t.logger = logger
}

func (t *TimeTracker) Cleanup() {
	if err := t.logger.Sync(); err != nil && !strings.Contains(err.Error(), "handle is invalid") {
		golangLog.Fatalf("failed flush logger: %s\n", err)
	}
}

func (t *TimeTracker) CLI() {
	// Parse options
	var inDir string
	flag.StringVar(&inDir, "in-dir", "times", "Directory containing time tracking CSV files")

	var billingPeriod string
	flag.StringVar(&billingPeriod, "billing-period", PeriodBiWeeklyStr, fmt.Sprintf("How often bills will be issued, valid values: %s", ValidPeriodStrJoined))

	var timeZone string
	flag.StringVar(&timeZone, "timezone", "EST", "timezone in which input and output times are in")

	var columnStartTime string
	flag.StringVar(&columnStartTime, "column-start-time", "time started", "Name of column which contains start time in format (24 hour time): YYYY-MM-DD HH:MM:SS")

	var columnEndTime string
	flag.StringVar(&columnEndTime, "column-end-time", "time ended", "Name of column which contains end time in format (24 hour time): YYYY-MM-DD HH:MM:SS")

	var columnComment string
	flag.StringVar(&columnComment, "column-comment", "comment", "Name of column which contains an optional description of what happened during the time period")

	var outputModeStr string
	flag.StringVar(&outputModeStr, "output", "dir=reports", fmt.Sprintf("Output mode, valid values: %s", OutputModesHelpStr))

	flag.Parse()

	// Convert options into programmatic values
	billingPeriodDuration, err := ParsePeriod(billingPeriod)
	if err != nil {
		t.logger.Fatal("failed to parse billing-period option", zap.Error(err))
	}

	outputMode, err := ParseOutputMode(outputModeStr)
	if err != nil {
		t.logger.Fatal("failed to parse output-mode option", zap.Error(err))
	}

	// Run logic
	opts := TimeTrackerReportOpts{
		InDir:           inDir,
		BillingPeriod:   billingPeriodDuration,
		Timezone:        timeZone,
		ColumnStartTime: columnStartTime,
		ColumnEndTime:   columnEndTime,
		ColumnComment:   columnComment,
		OutputMode:      outputMode,
	}

	if err := t.Report(opts); err != nil {
		t.logger.Fatal("failed to run report", zap.Any("error", err))
	}
}

type InputDirectoryError struct {
	InDir string
	Err   error
}

func (e InputDirectoryError) MarshallLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("in-dir", e.InDir)
	enc.AddString("error", e.Err.Error())
	return nil
}

func (e InputDirectoryError) Error() string {
	return e.Err.Error()
}

type InputFileError struct {
	FilePath string
	Err      error
}

func (e InputFileError) MarshallLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file-path", e.FilePath)
	enc.AddString("error", e.Err.Error())

	return nil
}

func (e InputFileError) Error() string {
	return e.Err.Error()
}

type MissingInputCSVColumnError struct {
	FilePath   string
	ColumnName string
}

func (e MissingInputCSVColumnError) MarshallLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file-path", e.FilePath)
	enc.AddString("error", e.Error())

	return nil
}

func (e MissingInputCSVColumnError) Error() string {
	return fmt.Sprintf("'%s' column not found in CSV file", e.ColumnName)
}

type InputCSVRowParseError struct {
	FilePath      string
	RowNum        int
	ColumnName    string
	ParseRawValue string
	Err           error
}

func (e InputCSVRowParseError) MarshallLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file-path", e.FilePath)
	enc.AddInt("row", e.RowNum)
	enc.AddString(e.ColumnName, e.ParseRawValue)
	enc.AddString("error", e.Err.Error())

	return nil
}

func (e InputCSVRowParseError) Error() string {
	return e.Err.Error()
}

type InputCSVRowError struct {
	FilePath string
	RowNum   int
	Err      error
}

func (e InputCSVRowError) MarshallLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file-path", e.FilePath)
	enc.AddInt("row", e.RowNum)
	enc.AddString("error", e.Err.Error())

	return nil
}

func (e InputCSVRowError) Error() string {
	return e.Err.Error()
}

func (t *TimeTracker) Report(opts TimeTrackerReportOpts) error {
	// Read in data
	inFileNames, err := os.ReadDir(opts.InDir)
	if err != nil {
		return InputDirectoryError{
			InDir: opts.InDir,
			Err:   fmt.Errorf("failed to read input directory: %s", err),
		}
	}

	timeEntries := make(map[string]TimeEntry) // keys are hashes of the values
	for _, fileEntry := range inFileNames {
		filePath := filepath.Join(opts.InDir, fileEntry.Name())
		// Open file
		f, err := os.Open(filePath)
		if err != nil {
			return InputFileError{
				FilePath: filePath,
				Err:      fmt.Errorf("failed to open input file: %s", err),
			}
		}

		reader := csv.NewReader(f)

		// Read headers
		headers, err := reader.Read()
		if err != nil {
			return InputFileError{
				FilePath: filePath,
				Err:      fmt.Errorf("failed to read CSV headers: %s", err),
			}
		}
		headerMap := make(map[string]int)
		for i, key := range headers {
			headerMap[key] = i
		}

		// Check for required columns
		if _, ok := headerMap[opts.ColumnStartTime]; !ok {
			return MissingInputCSVColumnError{
				FilePath:   filePath,
				ColumnName: opts.ColumnStartTime,
			}
		}

		if _, ok := headerMap[opts.ColumnEndTime]; !ok {
			return MissingInputCSVColumnError{
				FilePath:   filePath,
				ColumnName: opts.ColumnEndTime,
			}
		}

		if _, ok := headerMap[opts.ColumnComment]; !ok {
			return MissingInputCSVColumnError{
				FilePath:   filePath,
				ColumnName: opts.ColumnComment,
			}
		}

		// Parse rows into TimeEntry structs
		rows, err := reader.ReadAll()
		if err != nil {
			return InputFileError{
				FilePath: filePath,
				Err:      fmt.Errorf("failed to read rows of CSV: %s", err),
			}
		}

		for rowI, row := range rows {
			// Parse date times
			startTimeStr := fmt.Sprintf("%s %s", row[headerMap[opts.ColumnStartTime]], opts.Timezone)
			startTime, err := time.Parse(InputTimeFormat, startTimeStr)
			if err != nil {
				return InputCSVRowParseError{
					FilePath:      filePath,
					RowNum:        rowI,
					ColumnName:    opts.ColumnStartTime,
					ParseRawValue: startTimeStr,
				}
			}

			endTimeStr := fmt.Sprintf("%s %s", row[headerMap[opts.ColumnEndTime]], opts.Timezone)
			endTime, err := time.Parse(InputTimeFormat, endTimeStr)
			if err != nil {
				return InputCSVRowParseError{
					FilePath:      filePath,
					RowNum:        rowI,
					ColumnName:    opts.ColumnEndTime,
					ParseRawValue: endTimeStr,
				}
			}

			// Create entry
			entry := TimeEntry{
				StartTime: startTime,
				EndTime:   endTime,
				Comment:   row[headerMap[opts.ColumnComment]],
			}
			entryHash, err := entry.Hash()
			if err != nil {
				return InputCSVRowError{
					FilePath: filePath,
					RowNum:   rowI,
					Err:      fmt.Errorf("failed to hash time entry: %s", err),
				}
			}
			timeEntries[entryHash] = entry
		}
	}

	timeEntriesList := []TimeEntry{}
	for _, entry := range timeEntries {
		timeEntriesList = append(timeEntriesList, entry)
	}

	sort.Slice(timeEntriesList, func(i, j int) bool {
		return timeEntriesList[i].StartTime.Before(timeEntriesList[j].StartTime)
	})

	if len(timeEntriesList) == 0 {
		return fmt.Errorf("no time entries")
	}

	// Roll-up into reports
	firstStartDate := timeEntriesList[0].StartTime

	periodStart := time.Date(firstStartDate.Year(), firstStartDate.Month(), 1, 0, 0, 0, 0, firstStartDate.Location())
	periodEnd := periodStart.Add(opts.BillingPeriod)

	currentBillingPeriod := BillingPeriod{
		StartTime: periodStart,
		EndTime:   periodEnd,
		Entries:   []TimeEntry{},
	}
	billingPeriods := []BillingPeriod{}

	for _, entry := range timeEntriesList {
		// Check if starting a new period
		if entry.StartTime.After(periodEnd) {
			billingPeriods = append(billingPeriods, currentBillingPeriod)

			periodStart = periodStart.Add(opts.BillingPeriod)
			periodEnd = periodEnd.Add(opts.BillingPeriod)

			currentBillingPeriod = BillingPeriod{
				StartTime: periodStart,
				EndTime:   periodEnd,
				Entries:   []TimeEntry{},
			}
		}

		// Add time entry
		currentBillingPeriod.Entries = append(currentBillingPeriod.Entries, entry)
	}

	billingPeriods = append(billingPeriods, currentBillingPeriod)

	// Output
	reportCols := TimeTrackerReportCSVColumns{
		ColumnStartTime: opts.ColumnStartTime,
		ColumnEndTime:   opts.ColumnEndTime,
		ColumnComment:   opts.ColumnComment,
		ColumnDuration:  "duration",
	}
	if err := opts.OutputMode.Output(billingPeriods, reportCols); err != nil {
		return fmt.Errorf("failed to output reports: %s", err)
	}

	return nil
}

func main() {
	tracker := NewTimeTracker()
	tracker.Init()

	defer tracker.Cleanup()

	tracker.CLI()
}
