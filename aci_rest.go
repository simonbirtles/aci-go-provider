package aci

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"strings"
	"time"
	"log"
)

/*
* Implements:
* APIC Login
*
* Returns:
* string:APIC-cookie
* error
*
 */
func Aci_login(host, username, password string) (string, error) {

	url := fmt.Sprintf("https://%s/api/aaaLogin.json", host)

	payload_string := fmt.Sprintf(`{"aaaUser": {"attributes": {"name": "%s", "pwd" : "%s"}}} `, username, password)
	var payload = []byte(payload_string)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr, Timeout: time.Second * 10}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	fmt.Println("\nresponse Status Code:", resp.StatusCode)
	fmt.Println("\nresponse Headers:", resp.Header)

	// 2XX Response
	if resp.StatusCode > 199 && resp.StatusCode < 300 {
		// get the body of the response
		// grab and save APIC cookie
		cookies := resp.Cookies()
		var apic_cookie string
		for _, cookie := range cookies {
			if cookie.Name == "APIC-cookie" {
				apic_cookie = cookie.Value
				break
			}
		}

		if len(apic_cookie) == 0 {
			return "", errors.New("APIC login returned 2XX but APIC did not return a valid APIC-cookie.")
		}

		//fmt.Println("")
		//fmt.Println("\nBody Response")
		//fmt.Println(string(body))
		return apic_cookie, nil

		// 4XX Response
	} else if resp.StatusCode > 399 && resp.StatusCode < 500 {

		// e.g. {"totalCount":"1","imdata":[{"error":{"attributes":{"code":"401","text":"Username or password is incorrect - FAILED local authentication"}}}]}
		if HasContentType(&resp.Header, "application/json") == true {

			// non 2XX response has JSON payload, read out the error string
			var json_resp interface{}
			body, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &json_resp)

			if err != nil {
				fmt.Printf("[DEBUG] ACI Login - Unmarshall Error")
				return "", err

			} else {

				// TODO requires error checking on keys
				imdata, ok := json_resp.(map[string]interface{})["imdata"]
				if !ok {
					return "", errors.New(fmt.Sprintf("[DEBUG] APIC login failed with status code: %d with malformed response payload ", resp.StatusCode) )
				}
				error, ok := imdata.([]interface{})[0].(map[string]interface{})["error"]
				if !ok {
					return "", errors.New(fmt.Sprintf("[DEBUG] APIC login failed with status code: %d with malformed response payload ", resp.StatusCode) )
				}
				attributes, ok := error.(map[string]interface{})["attributes"]
				if !ok {
					return "", errors.New(fmt.Sprintf("[DEBUG] APIC login failed with status code: %d with malformed response payload ", resp.StatusCode) )
				}
				errText, ok  := attributes.(map[string]interface{})
				if !ok {
					return "", errors.New(fmt.Sprintf("[DEBUG] APIC login failed with status code: %d with malformed response payload ", resp.StatusCode) )
				}

				return "", errors.New(fmt.Sprintf("%s", errText["text"]))
			}

		// no json payload in non 2XX response
		} else {
			errText := fmt.Sprintf("\n\n[DEBUG] ACI Login - %s HTTP return code. No JSON Payload", resp.Status)
			return "", errors.New(errText)
		}

	// 1XX, 3XX, 5XX Response
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		errText := fmt.Sprintf("\n\n[DEBUG] ACI Login - HTTP POST failed with status: %s\n\n[BODY]: %s\n\n[URL]: %s", resp.Status, string(body), url )
		return "", errors.New(errText)
	}

	return "", errors.New("[DEBUG]  ACI Login - Unknown Error")

}

/*
* Implements:
* APIC REST GET
*
* Returns:
* []byte : Response Payload
* error
*
*/
func Get(info *ApicGetInfo) ([]byte, error) {

	if len((*info).ApicClient.Cookie) == 0 {
		return nil, errors.New("No APIC cookie provided.")
	}

	if len((*info).Path) == 0 {
		return nil, errors.New("No URI path provided.")
	}

	queryfilter := formatQueryFilter(&((*info).Filter))
	url := fmt.Sprintf("https://%s/api/%s", (*info).ApicClient.ApicHosts[0], info.Path )

	if strings.HasSuffix(url, ".xml") {
         return nil, errors.New(fmt.Sprintf("Error: XML format requested, only JSON supported.") )
    }

	if !strings.HasSuffix(url, ".json") {
		url += ".json"
	}

	url += queryfilter

	req, err := http.NewRequest("GET", url, nil)
	apiccookie := new(http.Cookie)
	apiccookie.Name = "APIC-Cookie"
	apiccookie.Value = (*info).ApicClient.Cookie
	req.AddCookie(apiccookie)

	//fmt.Println("Passing APIC-Cookie: ", (*info).Cookie)

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr, Timeout: time.Second * 10}	
	
	// Make GET Request
	time.Sleep( time.Duration((*info).Delay) * time.Millisecond)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[DEBUG} acirest: error %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check Return HTTP Status Codes
	if resp.StatusCode == 401 {
		// 401 Unauthorised
		log.Printf("[DEBUG} acirest: APIC rejected credentials for this request. [401 Unauthorised]")
		return nil, errors.New("APIC rejected credentials for this request. [401 Unauthorised]")
		
	} else if resp.StatusCode > 399 && resp.StatusCode < 500  {
		// 4XX Client Error
		log.Printf("[DEBUG} acirest: APIC Response with Client Error: [%s]", resp.Status)
		return nil, errors.New(fmt.Sprintf("APIC Response with Client Error: [%s]", resp.Status) )

	} else if resp.StatusCode == 504 {
		// 504 Gateway Timeout
		log.Printf("[DEBUG} acirest: APIC connection Gateway Timeout: [%s]", resp.Status)
		return nil, errors.New(fmt.Sprintf("APIC connection Gateway Timeout: [%s]", resp.Status) )

	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// <200 or >299 Catch All Others (1-199, 300-399, 500-503, 505-599)
		log.Printf("[DEBUG} acirest: APIC REST error: [%s]", resp.Status)
		return nil, errors.New(fmt.Sprintf("APIC REST error: [%s]", resp.Status) )
	}
	
	// 2XX success 
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG} acirest: response Status Code:", resp.Status)
	//log.Printf("[DEBUG} acirest: response Headers:", resp.Header)
	//log.Printf("[DEBUG} acirest: response Body", string(body))
	return body, nil
}

/*
* Implements:
* APIC REST POST 
*
* Returns:
* []byte : Response Payload
* error
*
*/
func Post(params *ApicPostInfo) ([]byte, error) {

	if len((*params).ApicClient.Cookie) == 0 {
		return nil, errors.New("No APIC cookie provided.")
	}

	if len((*params).Path) == 0 {
		return nil, errors.New("No URI path provided.")
	}

	if len((*params).Payload) == 0 {
		return nil, errors.New("No payload provided.")
	}

	queryfilter := formatQueryFilter(&((*params).Filter))
	url := fmt.Sprintf("https://%s/api/%s", (*params).ApicClient.ApicHosts[0], params.Path)

	if strings.HasSuffix(url, ".xml") {
         return nil, errors.New(fmt.Sprintf("Error: XML format requested, only JSON supported.") )
    }

	if !strings.HasSuffix(url, ".json") {
		url += ".json"
	}

	url += queryfilter

	fmt.Println(url)
	fmt.Println(bytes.NewBuffer((*params).Payload))

	// Build POST Request 
	req, err := http.NewRequest("POST", url, bytes.NewBuffer((*params).Payload))
	if err != nil {
		return nil, err
	}
	// content type
	req.Header.Set("Content-Type", "application/json")
	// add cookie
	apiccookie := new(http.Cookie)
	apiccookie.Name = "APIC-Cookie"
	apiccookie.Value = (*params).ApicClient.Cookie
	req.AddCookie(apiccookie)
	// transport options - no verify SSL
	tr := &http.Transport{ TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	// Do POST
	resp, err := client.Do(req)	
	if err != nil {
		return nil, err
	}	
	defer resp.Body.Close()

	// Check Return HTTP Status Codes
	if resp.StatusCode == 400 {
		// 400 Bad Request
		errorText := getApic4XXErrorText(resp)
		return nil, errors.New(fmt.Sprintf("APIC reported this POST as a Bad Request. [400 Bad Request] - [%s]", errorText) )
		
	} else if resp.StatusCode == 401 {
		// 401 Unauthorised
		return nil, errors.New("APIC rejected credentials for this request. [401 Unauthorised]")
		
	} else if resp.StatusCode > 399 && resp.StatusCode < 500  {
		// 4XX Client Error
		return nil, errors.New(fmt.Sprintf("APIC Response with Client Error: [%s]", resp.Status) )

	} else if resp.StatusCode == 504 {
		// 504 Gateway Timeout
		return nil, errors.New(fmt.Sprintf("APIC connection Gateway Timeout: [%s]", resp.Status) )

	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// <200 or >299 Catch All Others (1-199, 300-399, 500-503, 505-599)
		return nil, errors.New(fmt.Sprintf("APIC REST error: [%s]", resp.Status) )
	}

	// 2XX success 
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("\nresponse Status Code:", resp.Status)
	fmt.Println("\nresponse Headers:", resp.Header)
	fmt.Println("\nresponse Body", string(body))
	time.Sleep( time.Duration((*params).Delay) * time.Millisecond)
	return body, nil
}

/*
* Implements:
* APIC REST DELETE
* Should be provided ApicDeleteInfo struct
*
* Returns:
* error
*
*/
func Delete(info *ApicDeleteInfo) error {

	dn := (*info).Path
	if len(dn) == 0 {
		return errors.New(fmt.Sprintf("Error: Empty DN") )
	}
	
	if strings.HasSuffix(dn, ".xml") {
		return errors.New(fmt.Sprintf("Error: XML format requested, only JSON supported.") )
   	}
   
    if !strings.HasSuffix(dn, ".json") {
		dn += ".json"
	}

	url := fmt.Sprintf( "https://%s/api/%s", (*info).ApicClient.ApicHosts[0], dn ) 
	
	// Build POST Request 
	req, err := http.NewRequest("DELETE", url, nil)
	apiccookie := new(http.Cookie)
	apiccookie.Name = "APIC-Cookie"
	apiccookie.Value = (*info).ApicClient.Cookie
	req.AddCookie(apiccookie)
	// transport options - no verify SSL
	tr := &http.Transport{ TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	// Do POST
	resp, err := client.Do(req)
	time.Sleep(1000 * time.Millisecond)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	return nil

}

/*
* Implements:
* Creates a formatted APIC URL Query String from ApicQueryFilter stuct
*
* Returns:
* string : URL query string ?xxx=aaa&...
*
*/
func formatQueryFilter(queryfilter *ApicQueryFilter) string {

	var querystring string
	// returns a new Value initialized to the concrete value stored in the interface ApicQueryFilter
	s := reflect.ValueOf(queryfilter).Elem()
	// Value is the reflection interface to a Go value.
	typeOfT := s.Type()
	// loop through fields in the struct
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		//fmt.Printf("%d: %s %s = %v\n", i, typeOfT.Field(i).Tag.Get("json"), f.Type(), f.Interface())
		if len(f.Interface().(string)) > 0 {
			querystring += fmt.Sprintf("&%s=%v", typeOfT.Field(i).Tag.Get("json"), f.Interface())
		}
	}
	querystring = strings.Replace(querystring, "&", "?", 1)
	return querystring
}

/*
* Implements:
* Checks Response Header for content-type string against given mimetype string
*
* Returns:
* bool : true if given mimetype string is header content type
*
*/
func HasContentType(r *http.Header, mimetype string) bool {
	contentType := r.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}
	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}

/*
* Implements:
* On an APIC 4XX error, extract the error text from the APIC response
*
* Returns:
* string: The APIC response error text
*
*/
func getApic4XXErrorText(resp *http.Response) (string) {

	// e.g. {"totalCount":"1","imdata":[{"error":{"attributes":{"code":"401","text":"Username or password is incorrect - FAILED local authentication"}}}]}
	if HasContentType(&resp.Header, "application/json") == true {

		// non 2XX response has JSON payload, read out the error string
		var json_resp interface{}
		body, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(body, &json_resp)

		if err != nil {
			fmt.Printf("Unmarshall Error")
			return fmt.Sprintf("APIC request failed with status code: %d with malformed response payload [Unmarshall Error]", resp.StatusCode)

		} else {

			// TODO requires error checking on keys
			imdata, ok := json_resp.(map[string]interface{})["imdata"]
			if !ok {
				return fmt.Sprintf("APIC request failed with status code: %d with malformed response payload", resp.StatusCode) 
			}
			error, ok := imdata.([]interface{})[0].(map[string]interface{})["error"]
			if !ok {
				return fmt.Sprintf("APIC request failed with status code: %d with malformed response payload", resp.StatusCode) 
			}
			attributes, ok := error.(map[string]interface{})["attributes"]
			if !ok {
				return fmt.Sprintf("APIC request failed with status code: %d with malformed response payload", resp.StatusCode) 
			}
			errText, ok  := attributes.(map[string]interface{})
			if !ok {
				return fmt.Sprintf("APIC request failed with status code: %d with malformed response payload", resp.StatusCode) 
			}

			return fmt.Sprintf("%s", errText["text"])
		}

	// no json payload in non 2XX response
	} else {
		errText := fmt.Sprintf("\n\n%s HTTP return code. No JSON Payload", resp.Status)
		return errText
	}
}