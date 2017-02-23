package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/cloudfoundry-community/go-cfclient"
)

var myClient *cfclient.Client
var myAppName string

/*
 * Handler that reads the request and parses the response
 *
 * Will look at the URL.Path and find an app that matches this name.  If one is found, will
 * load the environment variables from the cfclient.App object, then build a request and get
 * all environment variables using the REST API request.
 *
 * @param resp - http.ResponseWriter used to output to the browser
 * @param req - http.Request object
 *
 * @return <none>
 *
 * @see #getApp
 * @see #getAllEnv
 */
func handler(resp http.ResponseWriter, req *http.Request) {
	appName := req.URL.Path[1:]
	fmt.Fprintf(resp, "Looking for app: %q ...\n\n", appName)
	if app, err := getApp(appName); nil == err {
		fmt.Fprintf(resp, "Found the below environment variables for %q \n\n", app.Name)
		for k, v := range app.Environment {
			fmt.Fprintf(resp, "  %q  :  %q\n", k, v)
		}

		fmt.Fprintf(resp, "\nAdditional environment variables:\n\n")
		getAllEnv(app.Guid, resp)
	} else {
		fmt.Fprintf(resp, "Error: %q \n\n", err)
	}
}

/*
 * Builds a new request to get all environment variables based on App GUID
 *
 * @param guid - App GUID, found by searching for App.Name and calling App.Guid
 * @param parentResp - http.ResponseWriter that can be used to output information
 *
 * @return <none>
 *
 * @see #getApp
 */
func getAllEnv(guid string, parentResp http.ResponseWriter) {
	r := myClient.NewRequest("GET", "/v2/apps/"+guid+"/env")
	if resp, reqErr := myClient.DoRequest(r); nil == reqErr {
		resBody, readErr := ioutil.ReadAll(resp.Body)
		if nil == readErr {
			fmt.Fprintf(parentResp, "\n\n%s\n", string(resBody[:]))
		} else {
			log.Printf("Error reading app request %v", resBody)
		}
	} else {
		fmt.Fprintf(parentResp, "Error requesting apps: %v", reqErr)
	}
}

/*
 * Gets an app using the given name.  Must iterate through all apps
 *
 * @param name - string name of app
 * @return
 *		cfclient.App - app we found
 *		tmpErr - error if we can't match the given name to an app
 */
func getApp(name string) (tmpApp cfclient.App, tmpErr error) {
	if myAppName != name {
		apps, _ := myClient.ListApps()
		for _, tmpApp := range apps {
			if name == tmpApp.Name {
				return tmpApp, nil
			}
		}
	}
	return tmpApp, fmt.Errorf("Could not find app with name %q", name)
}

/*
 * Main entry point
 *
 * @return <none>
 */
func main() {
	myAppName = os.Getenv("APP_NAME")
	appPort := os.Getenv("PORT")
	skipSsl, boolErr := strconv.ParseBool(os.Getenv("SKIP_SSL_VALIDATION"))
	if nil != boolErr {
		skipSsl = false
	}

	c := &cfclient.Config{
		ApiAddress:        os.Getenv("API_ADDRESS"),
		Username:          os.Getenv("API_USERNAME"),
		Password:          os.Getenv("API_PASSWORD"),
		SkipSslValidation: skipSsl,
	}
	tmpClient, err := cfclient.NewClient(c)
	if nil == err {
		myClient = tmpClient
	} else {
		fmt.Println(err)
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+appPort, nil)
}
