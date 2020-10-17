package service

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"implementsme.com/ifttt/smartthings/thermostatupdater/smartthings"
)

// ListenPort is the port this service listens on
const ListenPort = "65123"

// SecurityKey is required for all requests
const SecurityKey = "AH4NgXB8gvLEmMs7$b08vJsNMQO*UTWI%RPa@%1SuBk91E2B$"

// HandleHTTPRequests binds to a port and handles incoming http requests
func HandleHTTPRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/Test", test)
	myRouter.HandleFunc("/PriceChange/{newPrice}/{securityKey}", handlePriceChange)

	log.Fatal(http.ListenAndServe(":"+ListenPort, myRouter))
}

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Testing")
	fmt.Println("Endpoint Hit: /Test")
}

// handlePriceChange will notify all subscribers of the price change
func handlePriceChange(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	newPriceString := urlParams["newPrice"]
	inputSecurityKey := urlParams["securityKey"]

	if inputSecurityKey != SecurityKey {
		log.Println("Invalid security key:" + inputSecurityKey)
		return
	}

	newPriceFloat32, parsePriceErr := strconv.ParseFloat(newPriceString, 32)
	if parsePriceErr != nil {
		http.Error(w, "Unable to parse new price: "+newPriceString, http.StatusBadRequest)
		return
	}

	thermHandleErr := smartthings.ThermostatHandlePrice(float32(newPriceFloat32))
	if thermHandleErr != nil {
		http.Error(w, "Unable to set thermostat", http.StatusInternalServerError)
		return
	}
}
