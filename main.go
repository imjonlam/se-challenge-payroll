// Main package
package main

import (
	"fmt"
)

func main() {
	app := App{}

	// start app
	app.Initialize(USERNAME, PASSWORD, HOSTNAME, DBNAME)

	// listen and serve
	app.Run(fmt.Sprintf(":%d", PORT))
}
