package smartthings

// TODO:  Organize this with an event/subscriber model
// https://flaviocopes.com/golang-event-listeners/

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/allthingscode/ErcotRtLmpWatcher/notify"
)

var lastTempUsed uint64

// TODO:  Setup a log file so that history can be easily reviewed.

// ThermostatMode determines if we're doing heating or cooling
const ThermostatMode = "cooling"

type configuration struct {
	// SmartThingsAPIToken is used to authenticate when calling the SmartThings API
	APIToken string

	// ThermostatCommandURL is the full SmartThings API URL to send commands to my home thermostat
	ThermostatCommandURL string
}

var config = configuration{}

// Configure initializes settings for this package
func Configure(
	StAPIToken string,
	StThermostatCommandURL string,
	gmOauth2ConfigClientID string,
	gmOauth2ConfigClientSecret string,
	gmOauth2TokenAccessToken string,
	gmOauth2TokenRefreshToken string,
	gmTo string,
	gmSubjectPrefix string,
) {
	config.APIToken = StAPIToken
	config.ThermostatCommandURL = StThermostatCommandURL

	notify.Configure(
		gmOauth2ConfigClientID,
		gmOauth2ConfigClientSecret,
		gmOauth2TokenAccessToken,
		gmOauth2TokenRefreshToken,
		gmTo,
		gmSubjectPrefix,
	)
}

// HandlePrice will adjust the termostat based on price
func HandlePrice(price float32) error {

	if ThermostatMode == "cooling" {
		temp, getTempErr := getThermostatTempForCooling(price)
		if getTempErr != nil {
			return getTempErr
		}
		setTempErr := setThermostat(temp, ThermostatMode)
		if setTempErr != nil {
			return setTempErr
		}
	}

	return nil
}

// TODO:  Make the mapping come from a configuration file
// TODO:  Create named presets that can quickly change the mapping
// TODO:  Create different algorithms/profiles.  E.g. MaxCoolPerPrice
// getThermostatTempForCooling determines the ideal thermostat setting, based on price
func getThermostatTempForCooling(price float32) (uint64, error) {

	// TODO:  Put this in a config file so it can be easily adjusted
	const MinTemp = uint64(72)

	type priceRange struct {
		low  float32
		high float32
	}
	var priceRanges = map[uint64]priceRange{
		uint64(72): priceRange{float32(-100.0), float32(1.0)},
		uint64(73): priceRange{float32(1.0), float32(2.0)},
		uint64(74): priceRange{float32(2.0), float32(2.5)},
		uint64(75): priceRange{float32(2.5), float32(4.0)},
		uint64(76): priceRange{float32(4.0), float32(5.0)},
		uint64(77): priceRange{float32(5.0), float32(6.0)},
		uint64(78): priceRange{float32(6.0), float32(10.0)},
		uint64(79): priceRange{float32(10.0), float32(11.0)},
		uint64(80): priceRange{float32(11.0), float32(100.0)},
	}
	//fmt.Printf("Price Ranges:  %v\n", priceRanges)

	temp := uint64(80)
	for i := 72; i <= 80; i++ {
		priceRangeV := priceRanges[uint64(i)]
		//fmt.Printf("Checking price range for %d:  %v\n", i, priceRangeV)
		if price > priceRangeV.low && price <= priceRangeV.high {
			temp = uint64(i)
			break
		}
	}

	// Enforce a minimum temp
	if temp < MinTemp {
		temp = MinTemp
	}

	return temp, nil
}

// getThermostatTempForHeating determines the ideal thermostat setting, based on price
func getThermostatTempForHeating(price float32) (uint64, error) {
	return uint64(0), errors.New("Heating mode is not yet implemented")
}

// setThermostat sends the setCoolingSetpoint command to a thermostat
func setThermostat(temp uint64, mode string) error {
	switch mode {
	case "cooling":
	case "heating":
	default:
		return errors.New("Invalid thermostat mode: " + mode)
	}

	// If the last request was for the same temp, do nothing.
	// This reduces the number of requests to the smartthings API.
	if lastTempUsed == temp {
		//log.Println("Temp is already " + strconv.FormatUint(temp, 10))
		return nil
	}

	// https://smartthings.developer.samsung.com/docs/devices/smartapp/working-with-devices.html
	httpReqBody := fmt.Sprintf(`
	{
		"commands": [
			{
				"component":  "main",
				"capability": "thermostat",
				"command":    "setCoolingSetpoint",
				"arguments":  [%d]
			}
		]
	}`, temp)

	// Create a new request using http
	httpRequest, httpRequestError := http.NewRequest("POST", config.ThermostatCommandURL, strings.NewReader(httpReqBody))
	if httpRequestError != nil {
		log.Println("Error on response.\n[ERRO] -", httpRequestError)
		return httpRequestError
	}

	// Add authorization header to the http request
	httpRequest.Header.Add("Authorization", "Bearer "+config.APIToken)

	// Send request using http Client
	httpClient := &http.Client{}
	clientDoResponse, clientDoErr := httpClient.Do(httpRequest)
	if clientDoErr != nil {
		log.Println("Error on response.\n[ERRO] -", clientDoErr)
		return clientDoErr
	}

	body, _ := ioutil.ReadAll(clientDoResponse.Body)
	log.Println("Set new temp:  " + strconv.FormatUint(temp, 10))
	log.Println("SmartThings API Response:  " + string([]byte(body)))

	sendEmailErr := notify.SendEmail("Set new temp:  "+strconv.FormatUint(temp, 10), "")
	if sendEmailErr != nil {
		log.Println("Error sending email notification.\n[ERRO] -", sendEmailErr)
		return sendEmailErr
	}

	lastTempUsed = temp
	return nil
}
