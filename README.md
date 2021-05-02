# Wave Software Development Challenge
## Project Description

Imagine that this is the early days of Wave's history, and that we are prototyping a new payroll system API. A front end (that hasn't been developed yet, but will likely be a single page application) is going to use our API to achieve two goals:

1. Upload a CSV file containing data on the number of hours worked per day per employee
1. Retrieve a report detailing how much each employee should be paid in each _pay period_

All employees are paid by the hour (there are no salaried employees.) Employees belong to one of two _job groups_ which determine their wages; job group A is paid $20/hr, and job group B is paid $30/hr. Each employee is identified by a string called an "employee id" that is globally unique in our system.

Hours are tracked per employee, per day in comma-separated value files (CSV).
Each individual CSV file is known as a "time report", and will contain:

1. A header, denoting the columns in the sheet (`date`, `hours worked`,
   `employee id`, `job group`)
1. 0 or more data rows

In addition, the file name should be of the format `time-report-x.csv`,
where `x` is the ID of the time report represented as an integer. For example, `time-report-42.csv` would represent a report with an ID of `42`.

You can assume that:

1. Columns will always be in that order.
1. There will always be data in each column and the number of hours worked will always be greater than 0.
1. There will always be a well-formed header line.
1. There will always be a well-formed file name.

A sample input file named `time-report-42.csv` is included in this repo.

### What your API must do:

We've agreed to build an API with the following endpoints to serve HTTP requests:

1. An endpoint for uploading a file.

   - This file will conform to the CSV specifications outlined in the previous section.
   - Upon upload, the timekeeping information within the file must be stored to a database for archival purposes.
   - If an attempt is made to upload a file with the same report ID as a previously uploaded file, this upload should fail with an error message indicating that this is not allowed.

2. An endpoint for retrieving a payroll report structured in the following way:

   _NOTE:_ It is not the responsibility of the API to return HTML, as we will delegate the visual layout and redering to the front end. The expectation is that this API will only return JSON data.

   - Return a JSON object `payrollReport`.
   - `payrollReport` will have a single field, `employeeReports`, containing a list of objects with fields `employeeId`, `payPeriod`, and `amountPaid`.
   - The `payPeriod` field is an object containing a date interval that is roughly biweekly. Each month has two pay periods; the _first half_ is from the 1st to the 15th inclusive, and the _second half_ is from the 16th to the end of the month, inclusive. `payPeriod` will have two fields to represent this interval: `startDate` and `endDate`.
   - Each employee should have a single object in `employeeReports` for each pay period that they have recorded hours worked. The `amountPaid` field should contain the sum of the hours worked in that pay period multiplied by the hourly rate for their job group.
   - If an employee was not paid in a specific pay period, there should not be an object in `employeeReports` for that employee + pay period combination.
   - The report should be sorted in some sensical order (e.g. sorted by employee id and then pay period start.)
   - The report should be based on all _of the data_ across _all of the uploaded time reports_, for all time.

As an example, given the upload of a sample file with the following data:

   | date       | hours worked | employee id | job group |
   | ---------- | ------------ | ----------- | --------- |
   | 2020-01-04 | 10           | 1           | A         |
   | 2020-01-14 | 5            | 1           | A         |
   | 2020-01-20 | 3            | 2           | B         |
   | 2020-01-20 | 4            | 1           | A         |

A request to the report endpoint should return the following JSON response:

   ```json
   {
     "payrollReport": {
       "employeeReports": [
         {
           "employeeId": "1",
           "payPeriod": {
             "startDate": "2020-01-01",
             "endDate": "2020-01-15"
           },
           "amountPaid": "$300.00"
         },
         {
           "employeeId": "1",
           "payPeriod": {
             "startDate": "2020-01-16",
             "endDate": "2020-01-31"
           },
           "amountPaid": "$80.00"
         },
         {
           "employeeId": "2",
           "payPeriod": {
             "startDate": "2020-01-16",
             "endDate": "2020-01-31"
           },
           "amountPaid": "$90.00"
         }
       ]
     }
   }
   ```

We consider ourselves to be language agnostic here at Wave, so feel free to use any combination of technologies you see fit to both meet the requirements and showcase your skills. We only ask that your submission:

- Is easy to set up
- Can run on either a Linux or Mac OS X developer machine
- Does not require any non open-source software

### Documentation:

Please commit the following to this `README.md`:

1. Instructions on how to build/run your application

    **WHERE TO CLONE TO?**
      - Clone repostory to: *%GOPATH%/src/github.com/{username}/se-challenge-payroll*

    **API ROUTES**
     - **GET**: `/report`
     - **POST**:  `/report`
        - form-data with key = `'file`'

    **REQUIREMENTS**:
      - **Language**:
        - GO (https://golang.org/dl/)
      - **Libraries used**
        - *run `go build` in repository to resolve all dependencies*
        1. gorm (https://gorm.io/):
            - `go get -u -v gorm.io/gorm`
        2. gorm mysql: 
            - `go get -u -v gorm.io/driver/mysql`
        3. gorilla/mux (https://github.com/gorilla/mux):
            - `go get -u -v github.com/gorilla/mux`


    **CONFIGURATION** (**important**):
      - *configuration file provided as exemplar only. In production, should be a dotenv and ignored from repository*
      - In the file [./env.go](env.go), modify the variables per your database configurations
        - `DBNAME` used for main application database
        - `TEST_DBNAME` used for running tests
      - **NOTE**:
        1. Application **does not** create database in schema for you. However, tables are setup for you.
        2. Tests **do not** create database for you
            - **DO NOT USE SAME DATABASE** - database table records are wiped during and after tests are performed.
            - *If for mock purposes, ignore above*

   **TO CREATE APPLICATION** (*optional*):
      - In repository (*%GOPATH%/src/github.com/{username}/se-challenge-payroll*)  : 
        - `go install`
      - Executable stored in: *%GOPATH%/bin/wave_payroll_api*
  
    **TO RUN APPLICATION (*setup database, start mux server*)**:
      - If an executable is created, simply run it. 
        - I.E.: `./wave_payroll_api`
      - Otherwise:
        - In repository use: `go run .`

    **Testing**
      - In repostory run:  `go test -v`

1. Answers to the following questions:
   1. How did you test that your implementation was correct?
    
      - Testing was performed in two ways:
        1. Postman was used for manual testing
        2. `go test` was used to simulate a series of request scenarios.
            - HTTP responses and database tables were validated
            - All tests shown in [./main_test.go](main_test.go)
      - Correctness determined by meeting expected results (outlined above) and passing all test cases

   2. If this application was destined for a production environment, what would you add or change?
      * **Changes**
        1. As mentioned earlier, I would add a dotenv file for database configuration and ignore from repository
        2. Relate all EmployeeID model fields as foreign keys to an actual Employee Table   
        3. Providing more accurate error messages to all cases (currently returns default error message for some instances)
      
      * **Additions**
        1. Add verbose logging for all API requests
        2. Add Update and Delete (to complete CRUD) requests

   3. What compromises did you have to make as a result of the time constraints of this challenge?
      - As this was the first time I've made an API as well as using GO, likely not all the best practices were used.
      - I wish I could've implemented a middleware for logging API requests
      - Relational database could have been normalized better, pay_groups table should not be hardcoded with preset groups/rates

**Special Note**
 - Per felixc, quote *"let's simplify by saying that an employee can only be in one job group at a time"*,
    - My implementation stores the group (and rate) outside of the time_sheets table. This means allows for more flexility
    - Instead of one employee being tied to a specific group, it is dependant on the uploaded time-report
      - This means that the application works as intended if said employee changes groups during the same pay period

 - Thank you for giving me the opportunity to learn to develop in something entirely outside of my scope!
