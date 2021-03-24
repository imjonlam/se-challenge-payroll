// Main package
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/imjonlam/waveapi/models"
)

// Global variable App
var app App

func TestMain(m *testing.M) {
	// start app
	app.Initialize(USERNAME, PASSWORD, HOSTNAME, TEST_DBNAME)

	// check if all tables exist
	hasTables()
	code := m.Run()
	resetTables()

	os.Exit(code)
}

// Test Retreiving from empty employee_reports.
// Expected: Status.OK
func TestGetEmptyPayrollReport(t *testing.T) {
	resetTables()

	// create and send request
	request, err := http.NewRequest("GET", "/report", nil)
	if err != nil {
		t.Error(err)
	}

	// check for http.StatusOK
	response := sendRequest(request)
	checkStatusCode(t, http.StatusOK, response.Code)
}

// Test uploading a novel time-report file
// Expected: Status.OK
func TestPostTimeReport(t *testing.T) {
	resetTables()

	fn := "time-report-42.csv"

	// get path to .csv file
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	fp := filepath.Join(wd, "sample_data", fn)

	// open file
	file, err := os.Open(fp)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	// create form file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fn)
	if err != nil {
		t.Error(err)
	}

	io.Copy(part, file)
	writer.Close()

	// create and send request
	request, err := http.NewRequest("POST", "/report", body)
	if err != nil {
		t.Error(err)
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())

	// check for http.StatusOK
	response := sendRequest(request)
	checkStatusCode(t, http.StatusOK, response.Code)

	// check success message
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)

	expected := "Success! Added time report with ID: {id}"
	if !strings.Contains(m["message"], "Success!") {
		t.Errorf("Expected message: \"%s\". Instead got: \"%s\"", expected, m["message"])
	}
}

// Test uploading a previously uploaded test-report file
// Expected: Status.BadRequest
func TestPreviouslyPosted(t *testing.T) {
	fn := "time-report-42.csv"

	// get path to .csv file
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	fp := filepath.Join(wd, "sample_data", fn)

	response := uploadFile(t, fp, "file", "/report")

	// check status code
	checkStatusCode(t, http.StatusBadRequest, response.Code)

	// check error message
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)

	expected := "A record with ID: {} already exists"
	if !strings.Contains(m["error"], "already exists") {
		t.Errorf("Expected error message: \"%s\". Instead got: \"%s\"", expected, m["error"])
	}
}

// Test uploading a non .csv file
// Expected: Status.BadRequest
func TestPostBadFileExt(t *testing.T) {
	resetTables()

	fn := "time-report-42.xlsx"

	// get path to .csv file
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	fp := filepath.Join(wd, "sample_data", fn)

	response := uploadFile(t, fp, "file", "/report")

	// check status code
	checkStatusCode(t, http.StatusBadRequest, response.Code)

	// check error message
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)

	expected := "Expected a CSV (.csv) file"
	if m["error"] != expected {
		t.Errorf("Expected error message: \"%s\". Instead got: \"%s\"", expected, m["error"])
	}
}

// Test uploading a badly named file
// Expected: Status.BadRequest
func TestPostBadFilename(t *testing.T) {
	resetTables()

	fn := "time-report-fortytwo.csv"

	// get path to .csv file
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	fp := filepath.Join(wd, "sample_data", fn)

	response := uploadFile(t, fp, "file", "/report")

	// check status code
	checkStatusCode(t, http.StatusBadRequest, response.Code)

	// check error message
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)

	expected := "The provided filename is incorrectly formated. Expected time-report-{id}.csv"
	if m["error"] != expected {
		t.Errorf("Expected error message: \"%s\". Instead got: \"%s\"", expected, m["error"])
	}
}

// Test uploading a time-report with invalid column headers
// Expected: Status.BadRequest
func TestPostBadColumnHeaders(t *testing.T) {
	resetTables()

	fn := "time-report-404.csv"

	// get path to .csv file
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	fp := filepath.Join(wd, "sample_data", fn)

	response := uploadFile(t, fp, "file", "/report")

	// check status code
	checkStatusCode(t, http.StatusBadRequest, response.Code)

	// check error message
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)

	expected := "Incorrect number of column headers found. Expected: date,hours worked,employee id,job group"
	if m["error"] != expected {
		t.Errorf("Expected error message: \"%s\". Instead got: \"%s\"", expected, m["error"])
	}
}

// Test uploading a time-report with bad datatypes.
// Ex. a record has a hours represented in string format (four vs 4)
// Expected: Status.BadRequest
func TestPostBadDataType(t *testing.T) {
	resetTables()

	fn := "time-report-32.csv"

	// get path to .csv file
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	fp := filepath.Join(wd, "sample_data", fn)

	response := uploadFile(t, fp, "file", "/report")

	// check status code
	checkStatusCode(t, http.StatusBadRequest, response.Code)

	// check error message
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)

	expected := "strconv.ParseFloat: parsing \"four\": invalid syntax"
	if m["error"] != expected {
		t.Errorf("Expected error message: \"%s\". Instead got: \"%s\"", expected, m["error"])
	}
}

// Sends a HTTP request
func sendRequest(r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	app.Router.ServeHTTP(recorder, r)

	return recorder
}

// Checks response status code with expected code
func checkStatusCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// Sends a POST request with an uploaded file
func uploadFile(t *testing.T, fp, filetype, url string) *httptest.ResponseRecorder {
	// open file
	file, err := os.Open(fp)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	// create form file
	fn := filepath.Base(fp)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(filetype, fn)
	if err != nil {
		t.Error(err)
	}

	io.Copy(part, file)
	writer.Close()

	// create and send request
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Error(err)
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())

	// check for http.StatusBadRequest
	return sendRequest(request)
}

// Checks if all requried tables exist, log and terminate otherwise
func hasTables() {
	tables := []string{"time_sheets", "pay_groups", "pay_periods", "employee_reports"}

	for _, table := range tables {
		exists := app.DB.Migrator().HasTable(table)
		if !exists {
			log.Fatalf("Table %s is missing in database %s", table, DBNAME)
		}
	}

	// validate pay_groups table
	checkPayGroups()
}

// Checks to see if pay_groups has correct entries
func checkPayGroups() {
	var groups []models.PayGroup

	result := app.DB.Find(&groups, []string{"A", "B"})
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	if result.RowsAffected != 2 {
		log.Fatal("Table pay_groups is missing entries for Group A/B")
	}
}

// Drops all records from time_sheets, pay_periods and employee_report tables
func resetTables() {
	tables := []string{"time_sheets", "employee_reports", "pay_periods"}

	for _, table := range tables {
		app.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}
