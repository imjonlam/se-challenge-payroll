// Main package
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/imjonlam/wave_payroll_api/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	csvDateFormat = "2/1/2006"
	dbDateFormat  = "2006-01-02"
)

// Application struct
type App struct {
	Router *mux.Router
	DB     *gorm.DB
}

// A Response struct
type Response struct {
	StatusCode int    `json:"statusCode"`
	StatusText string `json:"statusText"`
	Error      string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
}

// Initialize Application
func (a *App) Initialize(user, password, hostname, dbname string) {
	// get connection string
	dsn := dsn(user, password, hostname, dbname)

	// connect to schema
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Could not connect to the database, %s", err)
	}

	// set App variables
	a.DB = db
	a.Router = mux.NewRouter()

	// auto migrate required tables from models
	db.AutoMigrate(&models.TimeSheet{},
		&models.EmployeeReport{})

	// insert starting records
	if err := a.seed(); err != nil {
		log.Fatalf("Unable to pre-populate tables, %s", err)
	}

	// add handlers
	a.addHandlers()
}

// Add handlers to mux
func (a *App) addHandlers() {
	a.Router.HandleFunc("/report", a.getPayrollReport).Methods("GET")
	a.Router.HandleFunc("/report", a.insertTimeSheet).Methods("POST")
}

// Starts the server
func (a *App) Run(address string) {
	log.Fatal(http.ListenAndServe(address, a.Router))
}

// Insert records into pay_groups table
func (a *App) seed() error {
	groupA := models.PayGroup{ID: "A", Rate: 20.00}
	if err := groupA.CreatePayGroup(a.DB); err != nil {
		return err
	}

	groupB := models.PayGroup{ID: "B", Rate: 30.00}
	if err := groupB.CreatePayGroup(a.DB); err != nil {
		return err
	}

	return nil
}

// Returns all entries in employee_reports as JSON
func (a *App) getPayrollReport(w http.ResponseWriter, r *http.Request) {
	var report models.PayrollReport
	employeeReports := &report.PayrollReport

	if err := employeeReports.GetEmployeeReports(a.DB); err != nil {
		sendBadRequest(w, err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	}
}

// Retrieves a CSV from a HTTP POST request
// Further uploads contents to the datbase
func (a *App) insertTimeSheet(w http.ResponseWriter, r *http.Request) {
	// get file using key = 'file'
	file, handler, err := r.FormFile("file")
	if err != nil {
		sendBadRequest(w, err.Error())
		return
	}

	// get file information
	fn, ext := splitExt(handler.Filename)
	if ext != ".csv" {
		sendBadRequest(w, "Expected a CSV (.csv) file")
		return
	}

	// validate filename
	id, err := strconv.ParseInt(strings.Split(fn, "-")[2], 10, 32)
	if err != nil {
		sendBadRequest(w, "The provided filename is incorrectly formated. Expected time-report-{id}.csv")
		return
	}

	// test the existence of time report
	var testRecord models.TimeSheet
	if err := a.DB.First(&testRecord, id).Error; err != nil {
	} else {
		sendBadRequest(w, fmt.Sprintf("A record with ID: %d already exists", id))
		return
	}

	// read csv
	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		sendBadRequest(w, "Unable to read csv")
	}

	// verify column headers
	header := lines[0]
	if len(header) != 4 ||
		header[0] != "date" || header[1] != "hours worked" ||
		header[2] != "employee id" || header[3] != "job group" {
		sendBadRequest(w,
			"Incorrect number of column headers found. Expected: date,hours worked,employee id,job group")
		return
	}

	// start transaction
	// if at anytime an error occurs, entire transaction is rollbacked
	if err := a.DB.Transaction(func(tx *gorm.DB) error {
		for _, record := range lines[1:] {
			// parse record into proper datatypes
			date, err := time.Parse(csvDateFormat, record[0])
			if err != nil {
				sendBadRequest(w, err.Error())
				return err
			}

			hours, err := strconv.ParseFloat(record[1], 32)
			if err != nil {
				sendBadRequest(w, err.Error())
				return err
			}

			employeeID, err := strconv.Atoi(record[2])
			if err != nil {
				sendBadRequest(w, err.Error())
				return err
			}

			// create a new time_sheets entry
			timesheet := models.TimeSheet{
				ID:         uint(id),
				Date:       date,
				Hours:      float32(hours),
				EmployeeID: uint(employeeID),
				PayGroup:   models.PayGroup{ID: record[3]},
			}

			// insert into time_sheets
			if err := timesheet.CreatetTimeSheet(tx); err != nil {
				sendBadRequest(w, err.Error())
				return err
			}

			// determine pay_periods dates
			startDate := getStartDate(timesheet.Date)
			endDate := getEndDate(timesheet.Date)

			if timesheet.Date.Day() < 16 {
				endDate = startDate.AddDate(0, 0, 14)
			} else {
				startDate = startDate.AddDate(0, 0, 15)
			}

			// calculate total amount earned
			var payGroup models.PayGroup
			tx.Where("ID = ?", timesheet.PayGroupID).First(&payGroup)
			amount := payGroup.Rate * timesheet.Hours

			// create a new employee_reports entry
			employeeReport := models.EmployeeReport{
				Amount:     amount,
				EmployeeID: timesheet.EmployeeID,
				PayPeriod: models.PayPeriod{
					StartDate: startDate,
					EndDate:   endDate,
				},
			}

			// insert into employee_reports
			if err := employeeReport.CreateEmpoyeeReport(tx, amount); err != nil {
				sendBadRequest(w, err.Error())
				return err
			}
		}

		// commit
		return nil
	}); err == nil {
		// return to caller with StatusOK
		sendSuccess(w, fmt.Sprintf("Success! Added time report with ID: %d", id))
	}
}

// Splits a basename
// Returns the filename and extension separated
func splitExt(basename string) (string, string) {
	ext := filepath.Ext(basename)
	fn := basename[:len(basename)-len(ext)]

	return fn, ext
}

// Returns the stating date of the month
func getStartDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

// Returns the ending date of the month
func getEndDate(date time.Time) time.Time {
	return getStartDate(date).AddDate(0, 1, 0).Add(-time.Second)
}

// Returns a MYSQL connection string
func dsn(username, password, hostname, dbname string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True",
		username, password, hostname, dbname)
}

// Reply to HTTP request with StatusOK
func sendSuccess(w http.ResponseWriter, message string) {
	response := Response{
		StatusCode: http.StatusOK,
		StatusText: http.StatusText(http.StatusOK),
		Message:    message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Reply to HTTP request with BadRequest
func sendBadRequest(w http.ResponseWriter, err string) {
	response := Response{
		StatusCode: http.StatusBadRequest,
		StatusText: http.StatusText(http.StatusBadRequest),
		Error:      err,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}
