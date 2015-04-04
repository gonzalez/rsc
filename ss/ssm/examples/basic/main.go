// This basic example illustrates how to setup a Self-Service client to make a simple
// API call. The reference for the API can be found at
// https://s3.amazonaws.com/rs_api_docs/selfservice/manager/index.html#/
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/rightscale/rsc/rsapi"
	"github.com/rightscale/rsc/ss"
	"github.com/rightscale/rsc/ss/ssm"
)

func main() {
	// 1. Retrieve login and endpoint information
	email := flag.String("e", "", "Login email")
	pwd := flag.String("p", "", "Login password")
	account := flag.Int("a", 0, "Account id")
	host := flag.String("h", "us-3.rightscale.com", "RightScale API host")
	flag.Parse()
	if *email == "" {
		fail("Login email required")
	}
	if *pwd == "" {
		fail("Login password required")
	}
	if *account == 0 {
		fail("Account id required")
	}
	if *host == "" {
		fail("Host required")
	}

	// 2. Setup client using basic auth
	logger := log.New(os.Stdout, "", 0)
	auth := rsapi.LoginAuthenticator{Username: *email, Password: *pwd, Client: http.DefaultClient}
	client, err := ssm.New(*account, *host, &auth, logger, nil)
	if err != nil {
		fail("failed to create client: %s", err)
	}

	// 3. Make execution index call using expanded view
	l := client.ExecutionLocator(fmt.Sprintf("/projects/%d/executions", *account))
	execs, err := l.Index(rsapi.ApiParams{})
	if err != nil {
		fail("failed to list executions: %s", err)
	}
	sort.Sort(ByName(execs))

	// 4. Print executions launch from
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 5, 0, 1, ' ', 0)
	fmt.Fprintln(w, "Name\tState\tBy\tLink")
	fmt.Fprintln(w, "-----\t-----\t-----\t-----")
	for _, e := range execs {
		link := fmt.Sprintf("https://%s/manager/#/exe/%s", ss.HostFromLogin(client.Host),
			e.Id)
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s\t%s", e.Name, e.Status, e.CreatedBy.Email, link))
	}
	w.Flush()
}

type ByName []*ssm.Execution

func (b ByName) Len() int           { return len(b) }
func (b ByName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByName) Less(i, j int) bool { return b[i].Name < b[j].Name }

// Print error message and exit with code 1
func fail(format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Println(fmt.Sprintf(format, v...))
	os.Exit(1)
}