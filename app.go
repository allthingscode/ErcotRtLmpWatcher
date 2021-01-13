package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/allthingscode/ErcotRtLmpWatcher/ercot"
	"github.com/allthingscode/ErcotRtLmpWatcher/smartthings"
)

var exitRequestReceived = false

func main() {

	//service.HandleHTTPRequests()

	err := watchPrices()
	if err != nil {
		fmt.Println(err)
	}
}

func watchPrices() error {

	LoadSettings()
	//fmt.Printf("%v\n", config)
	smartthings.Configure(
		config.SmartThings.APIToken,
		config.SmartThings.ThermostatCommandURL,
		config.Gmail.Oauth2Config.ClientID,
		config.Gmail.Oauth2Config.ClientSecret,
		config.Gmail.Oauth2Token.AccessToken,
		config.Gmail.Oauth2Token.RefreshToken,
		config.Gmail.To,
		config.Gmail.SubjectPrefix,
	)

	setupCloseHandler()

	// TODO:  Move this loop into the ercotPriceWatcher.go script
	for exitRequestReceived == false {

		price, asOfTimestampRaw, err := ercot.GetRtLmpPrice("LZ_NORTH")
		if err != nil {
			return err
		}

		thermostatHandlerErr := smartthings.HandlePrice(price)
		if thermostatHandlerErr != nil {
			return thermostatHandlerErr
		}

		sleepDuration, err := calculateDurationUntilNextReload(asOfTimestampRaw)
		if err != nil {
			return err
		}
		//log.Printf("Sleeping for %v\n", sleepDuration)

		// BAIL if requested
		if exitRequestReceived == true {
			break
		}

		time.Sleep(sleepDuration)
		//fmt.Printf("Done sleeping at:  %v\n", time.Now())
		//fmt.Println("")
	}

	return nil
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS.
func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		exitRequestReceived = true
	}()
}

// calculateDurationUntilNextReload calculates the duration to wait before the next ERCOT reload
// TODO:  Move this into the ercotPriceWatcher.go script
func calculateDurationUntilNextReload(lastErcotReload string) (time.Duration, error) {

	myLocation, loadLocationError := time.LoadLocation("America/Chicago")
	if loadLocationError != nil {
		return time.Duration(0), loadLocationError
	}

	asOfTimestamp, _ := time.ParseInLocation(ercot.RtLmpLastUpdatedTimestampLayOut, lastErcotReload, myLocation)
	//fmt.Printf("As-Of time:  %v\n", asOfTimestamp)
	targetReload := asOfTimestamp.Add(time.Minute * 5).Add(time.Second * 30)
	//fmt.Printf("Target reload time:  %v\n", targetReload)

	now := time.Now()
	//fmt.Println(now)
	difference := targetReload.Sub(now)
	//fmt.Printf("Wait time:  %v\n", difference)

	if difference < 0 {
		// This means that the ERCOT webpage should have already reloaded by now
		// We'll give it 10 seconds and check again
		difference = time.Duration(10) * time.Second
	}

	return difference, nil
}
