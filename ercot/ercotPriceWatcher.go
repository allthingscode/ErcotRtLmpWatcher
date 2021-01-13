package ercot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
)

// TODO:  Update this to use channels for a pub/sub model of communication.
// https://flaviocopes.com/golang-event-listeners/

// RtLmpPriceWebpageURL is the public webpage that has current pricing
const RtLmpPriceWebpageURL = "http://www.ercot.com/content/cdr/html/hb_lz"

// RtLmpLastUpdatedTimestampLayOut is the format of the Last Updated timestamp on the website
const RtLmpLastUpdatedTimestampLayOut = "Jan 02, 2006 15:04:05"

// GetRtLmpPrice returns the current ERCOT price
func GetRtLmpPrice(settlementZone string) (float32, string, error) {
	// Get the HTML
	html, err := getWebpageHTML()
	if err != nil {
		return float32(0), "", err
	}
	//fmt.Println(html)

	price, asOfTimestampRaw, err := extractPrice(html, settlementZone)
	if err != nil {
		return float32(0), "", err
	}
	log.Printf("%f As Of %s\n", price, asOfTimestampRaw)

	return price, asOfTimestampRaw, nil
}

// getWebpageHTML returns the raw HTML from the ERCOT live pricing web page
func getWebpageHTML() (string, error) {

	response, httpGetError := http.Get(RtLmpPriceWebpageURL)
	if httpGetError != nil {
		return "", httpGetError
	}

	htmlBytes, bodyReadError := ioutil.ReadAll(response.Body)
	if bodyReadError != nil {
		return "", bodyReadError
	}

	return string(htmlBytes), nil
}

// extractPrice returns the current price for the requested settlement zone on the ERCOT web site
func extractPrice(
	html string,
	settlementZone string,
) (float32, string, error) {

	// Extract the Last Updated timestamp before we remove spaces
	asOfRegExp := regexp.MustCompile(`<div[^>]+>Last Updated:&nbsp;[ ]*([^<]+)</div>`)
	asOfHTMLMatches := asOfRegExp.FindAllStringSubmatch(html, -1)
	if asOfHTMLMatches == nil {
		return float32(0), "", errors.New("Unable to extract Last Updated value")
	}
	if len(asOfHTMLMatches) != 1 {
		return float32(0), "", errors.New("Got an unexpected number of extracted Last Updated values" + fmt.Sprint(len(asOfHTMLMatches)))
	}
	if len(asOfHTMLMatches[0]) != 2 {
		return float32(0), "", errors.New("Got an unexpected number of extracted Last Updated sub values" + fmt.Sprint(len(asOfHTMLMatches[0])))
	}
	asOfString := string(asOfHTMLMatches[0][1])

	whitespaceRegExp := regexp.MustCompile(`\s`)
	html = whitespaceRegExp.ReplaceAllLiteralString(html, "")
	//fmt.Println(html)

	// Parse the HTML
	// TODO:  See if we can just look for submatches in the row html
	// <tr><tdclass="tdLeft">LZ_NORTH</td><tdclass="tdLeft">18.30</td><tdclass="tdLeftred_text">-0.25</td><tdclass="tdLeft">18.30</td><tdclass="tdLeftred_text">-0.25</td></tr>
	settlementZoneRowRegExp := regexp.MustCompile(`<tr><tdclass="[^"]*">` + settlementZone + `</td><tdclass="[^"]*">[^<]*</td><tdclass="[^"]*">[^<]*</td><tdclass="[^"]*">([^<]+)</td><tdclass="[^"]*">[^<]*</td></tr>`)
	settlementZoneRowHTML := settlementZoneRowRegExp.FindString(html)
	//fmt.Println(settlementZoneRowHTML)
	//fmt.Println("\n")

	pricePointCellsRegExp := regexp.MustCompile(`<td[^>]+">([^<]+)</td>`)
	pricePointCellsHTMLMatches := pricePointCellsRegExp.FindAllStringSubmatch(settlementZoneRowHTML, -1)
	if pricePointCellsHTMLMatches == nil {
		return float32(0), asOfString, errors.New("Unable to extract price point table cells from " + settlementZone + " html row")
	}
	if len(pricePointCellsHTMLMatches) != 5 {
		return float32(0), asOfString, errors.New("Got an unexpected number of price point table cells from the " + settlementZone + " html row:  " + fmt.Sprint(len(pricePointCellsHTMLMatches)))
	}

	// The 4th table cell is "RTORPA + RTORDPA + LMP"
	priceMatch := pricePointCellsHTMLMatches[3]
	//fmt.Println(priceMatch)
	if len(priceMatch) != 2 {
		return float32(0), asOfString, errors.New("Got an unexpected number of price point matches:  " + fmt.Sprint(len(priceMatch)))
	}

	priceCleanupRegExp := regexp.MustCompile(`[^0-9.]+`)
	cleanedErcotPrice := priceCleanupRegExp.ReplaceAllString(priceMatch[1], "")

	price, parsePriceErr := strconv.ParseFloat(cleanedErcotPrice, 32)
	if parsePriceErr != nil {
		return float32(0), asOfString, errors.New("Unable to parse new price: " + priceMatch[1])
	}

	// Convert 19.20 to 1.920
	price = (math.Round(price*100) / 1000)

	return float32(price), asOfString, nil
}
