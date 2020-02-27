package cpanel

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var i int = 0
var insecure = true

// Cpaneldata structure for backup file data
type Cpaneldata struct {
	FileName   string
	FileHandle *os.File
}

// ConfigcPanel defines the config variables and types for cPanel instance.
type ConfigcPanel struct {
	HostName string `json:"hostname"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

var ResponseSizeLimit = (20 * 1024 * 1024) + 1337

func (c *JSONAPIGateway) api(req CpanelAPIRequest, out interface{}) error {
	vals := req.Arguments.Values(req.APIVersion)
	reqURL := fmt.Sprintf("https://%s:2083/", c.Hostname)
	switch req.APIVersion {
	case "uapi":
		reqURL += fmt.Sprintf("execute/%s/%s?%s", req.Module, req.Function, vals.Encode())
	case "2":
		fallthrough
	case "1":
		vals.Add("cpanel_jsonapi_user", c.Username)
		vals.Add("cpanel_jsonapi_apiversion", req.APIVersion)
		vals.Add("cpanel_jsonapi_module", req.Module)
		vals.Add("cpanel_jsonapi_func", req.Function)
		reqURL += fmt.Sprintf("json-api/cpanel?%s", vals.Encode())
	default:
		return fmt.Errorf("Unknown api version: %s", req.APIVersion)
	}

	httpReq, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	httpReq.SetBasicAuth(c.Username, c.Password)

	if c.cl == nil {
		c.cl = &http.Client{}
		c.cl.Transport = &http.Transport{
			DisableKeepAlives:   true,
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.Insecure,
			},
		}
	}

	resp, err := c.cl.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}

	// limit maximum response size
	lReader := io.LimitReader(resp.Body, int64(ResponseSizeLimit))

	bytes, err := ioutil.ReadAll(lReader)
	if err != nil {
		return err
	}

	if os.Getenv("DEBUG_CPANEL_RESPONSES") == "1" {
		log.Println(reqURL)
		log.Println(resp.Status)
		log.Println(req.Function)
		log.Println(req.Arguments)
		log.Println(vals)
		log.Println(string(bytes))
	}

	if len(bytes) == ResponseSizeLimit {
		return errors.New("API response maximum size exceeded")
	}

	return json.Unmarshal(bytes, out)
}

func (r BaseResult) Error() error {
	if r.ErrorString == "" {
		return nil
	}
	return errors.New(r.ErrorString)
}

// UAPI function creates a UAPI client for cPanel
func (c *JSONAPIGateway) UAPI(module, function string, arguments Args, out interface{}) error {
	req := CpanelAPIRequest{
		APIVersion: "uapi",
		Module:     module,
		Function:   function,
		Arguments:  arguments,
	}

	return c.api(req, out)
}

// API2 function creates API2 client
func (c *JSONAPIGateway) API2(module, function string, arguments Args, out interface{}) error {
	req := CpanelAPIRequest{
		APIVersion: "2",
		Module:     module,
		Function:   function,
		Arguments:  arguments,
	}

	var result API2Result
	err := c.api(req, &result)
	if err == nil {
		err = result.Error()
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(result.Result, out)
}

// JSONAPIGateway defines the properties of the client
type JSONAPIGateway struct {
	Hostname string
	Username string
	Password string
	Insecure bool
	cl       *http.Client
}

// APIGateway consitutes the client of UAPI and API1
type APIGateway interface {
	UAPI(module, function string, arguments Args, out interface{}) error
	API2(module, function string, arguments Args, out interface{}) error
}

// Api contains ApiGateway value
type Api struct {
	Gateway APIGateway
}

// CpanelAPI is used to access the Cpanel features
type CpanelAPI struct {
	Api
}

// CpanelAPIRequest consists all information of function to be used
type CpanelAPIRequest struct {
	Module      string `json:"module"`
	RequestType string `json:"reqtype"`
	Function    string `json:"func"`
	APIVersion  string `json:"apiversion"`
	Arguments   Args   `json:"args"`
}

// Args defines the arguments taken by the UAPI functions
type Args map[string]interface{}

func (a Args) Values(apiVersion string) url.Values {
	vals := url.Values{}
	for k, v := range a {
		if apiVersion == "1" {
			kv := strings.SplitN(k, "=", 2)
			if len(kv) == 1 {
				vals.Add(kv[0], "")
			} else if len(kv) == 2 {
				vals.Add(kv[0], kv[1])
			}
		} else {
			vals.Add(k, fmt.Sprintf("%v", v))
		}
	}
	return vals
}

//NewAPI generates New Api
func NewAPI(gw APIGateway) Api {
	return Api{
		Gateway: gw,
	}
}

//Close is implemented to use Gateway
func (c *JSONAPIGateway) Close() error {
	return nil
}

//NewJSONAPI returns the client to be used for accessing cPanel features
func NewJSONAPI(hostname, username, password string, insecure bool) (CpanelAPI, error) {
	c := &JSONAPIGateway{
		Hostname: hostname,
		Username: username,
		Password: password,
		Insecure: insecure,
	}
	// todo: a way to check the username/password here
	return CpanelAPI{NewAPI(c)}, nil
}

type API2Result struct {
	BaseResult
	Result json.RawMessage `json:"cpanelresult"`
}

type BaseResult struct {
	ErrorString string `json:"error"`
}

type BaseUAPIResponse struct {
	BaseResult
	StatusCode int      `json:"status"`
	Errors     []string `json:"errors"`
	Messages   []string `json:"messages"`
}

//FullBackuptoHomeDirAPIResponse is the type of response returned by fullbackup_to_homedir
type FullBackuptoHomeDirAPIResponse struct {
	BaseUAPIResponse
	Data struct {
		PID string `json:"pid"`
	} `json:"data"`
}

type BaseAPI2Response struct {
	BaseResult
	Event struct {
		Result int    `json:"result"`
		Reason string `json:"reason"`
	} `json:"event"`
}

type ListfullbackupsApiResponse struct {
	BaseAPI2Response
	Data []struct {
		Status    string `json:"status"`
		Localtime string `json:"localtime"`
		File      string `json:"file"`
		Time      int    `json:"time"`
		Reason    string `json:"reason"`
		Result    bool   `json:"result"`
	} `json:"data"`
}

// LoadcPanelProperty reads and parses the JSON file.
// that contains a cPanel instance's property.
// and returns all the properties as an object.
func LoadcPanelProperty(fullFileName string) (ConfigcPanel, error) { // fullFileName for fetching cPanel credentials from given JSON filename.
	var configcPanel ConfigcPanel

	// Open and read the json file
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configcPanel, err
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configcPanel)

	// Display read information.
	fmt.Println("\nReading cPanel configuration from file: ", fullFileName)

	// Display read information.
	fmt.Println("\nReading cPanel configuration from file: ", fullFileName)
	fmt.Println("Host Name\t: ", configcPanel.HostName)
	fmt.Println("User Name\t: ", configcPanel.UserName)
	fmt.Println("Password\t: ", configcPanel.Password)
	return configcPanel, nil
}

// ConnectToCpanel will connect to a cPanel instance,
// based on the read property from an external file.
// It returns a reference to an io.Reader with cPanel instance information.
func ConnectToCpanel(fullFileName string) (*Cpaneldata, error) {

	// Read cPanel instance's properties from an external file.
	configcPanel, err := LoadcPanelProperty(fullFileName)

	if err != nil {
		log.Fatal("Load cPanel Property:", err)
	}

	// Create connection with cPanel
	fmt.Println("\nConnecting to cPanel...")
	client, err := NewJSONAPI(configcPanel.HostName, configcPanel.UserName, configcPanel.Password, insecure)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(1 * time.Second)
	_, err = net.DialTimeout("tcp", configcPanel.HostName+":2083", timeout)
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to cPanel!")

	var list ListfullbackupsApiResponse
	err = client.Gateway.API2("Backups", "listfullbackups", Args{}, &list)
	prevLen := len(list.Data)

	// Creates a full backup to the user's home directory
	fmt.Println("Creating Full Backup...")
	var out FullBackuptoHomeDirAPIResponse
	err = client.Gateway.UAPI("Backup", "fullbackup_to_homedir", Args{
		"email": "",
	}, &out)

	if err != nil {
		log.Fatal("Full Backup Error : ", err)
	}

	time.Sleep(10 * time.Second) //Wait for backup file to be created

	var status string //status of the backup file "inprogress or complete"
	var fileName string

	for status != "complete" {
		// Lists the account's backup files.
		var list ListfullbackupsApiResponse
		err = client.Gateway.API2("Backups", "listfullbackups", Args{}, &list)
		if err != nil {
			log.Fatal(err)
		}

		currLen := len(list.Data)
		if currLen > prevLen {
			status = list.Data[currLen-1].Status
			fileName = list.Data[currLen-1].File
		}

	}

	fmt.Printf("Completed Full Backup:\t%s", fileName)

	// Created file handle for backup file
	file, err := os.Open("/home/" + configcPanel.UserName + "/" + fileName)

	if err != nil {
		log.Fatal(err)
	}

	return &Cpaneldata{FileHandle: file, FileName: fileName}, nil

}
