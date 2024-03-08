package Utilities

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

const (
	URLPosition    = "https://api.kite.trade/portfolio/positions"
	GTTOrderURL    = "https://api.kite.trade/gtt/triggers"
	GetGTTOrderURL = "https://api.kite.trade/gtt/triggers/"
	DeleteGTTOrder = "https://api.kite.trade/gtt/triggers/"
)

func SetEnvironmentVariables(envMap map[string]string) {
	for key, value := range envMap {
		err := os.Setenv(key, value)
		if err != nil {
			return
		}
	}
}

func OpenFile(loc string) []byte {
	content, err := ioutil.ReadFile(loc)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	return content
}

func OpenJSON(loc string, out interface{}) {
	err := json.Unmarshal(OpenFile(loc), out)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// For checking position is available or Not
func IsExistingPosition(positions string, symbol string) bool {
	// for _, position := range positions {
	if positions == symbol {
		return true
	}
	// }
	return false
}

func SetConfigurations() {
	var config ConfigJson
	OpenJSON("./Configurations/Config.json", &config)

	SetEnvironmentVariables(map[string]string{
		"BASE_LOG_PATH": config.LogPath,
		"PORT":          config.Port,
		"ACCESS_TOKEN":  config.AccessToken,
		"API_KEY":       config.ApiKey,
		"CANCEL_MINS":   config.CancelMins,
	})
}

// Here we are fetching the positions using API-Key and Access-Token
func GetPositions() (kiteconnect.Positions, error) {
	var positions kiteconnect.Positions

	apiKey := os.Getenv("API_KEY")
	accessToken := os.Getenv("ACCESS_TOKEN")

	kc := kiteconnect.New(apiKey)
	kc.SetAccessToken(accessToken)

	positions, err := kc.GetPositions()
	if err != nil {
		return kiteconnect.Positions{}, err
	}

	return positions, nil
}

func PlaceGTTOrder(signelData kiteconnect.Order, stopLossPrice float64, product, transactionType string) (GTTOrerResponse, error) { //parameters of orderdetails

	apiKey := os.Getenv("API_KEY")
	accessToken := os.Getenv("ACCESS_TOKEN")

	oppositeTransactionType := transactionType
	if oppositeTransactionType == "buy" {
		oppositeTransactionType = "sell"
	} else if oppositeTransactionType == "sell" {
		oppositeTransactionType = "buy"
	}
	condition := `{"exchange":"` + signelData.Exchange + `", "tradingsymbol":"` + signelData.TradingSymbol + `", "trigger_values":[` + strconv.Itoa(int(stopLossPrice)) + `], "last_price":` + strconv.Itoa(int(signelData.AveragePrice)) + `}`
	orders := `[{"exchange":"` + signelData.Exchange + `", "tradingsymbol": "` + signelData.TradingSymbol + `", "transaction_type": "` + oppositeTransactionType + `", "quantity": ` + strconv.Itoa(int(signelData.Quantity)) + `, "order_type": "LIMIT","product": "` + product + `", "price": ` + strconv.Itoa(int(signelData.AveragePrice)) + `}]`

	payload := fmt.Sprintf("type=single&condition=%s&orders=%s", condition, orders)

	// payload := encodeValues(data)
	client := &http.Client{}
	req, err := http.NewRequest("POST", GTTOrderURL, strings.NewReader(payload))

	if err != nil {
		fmt.Println(err)
		return GTTOrerResponse{}, err
	}

	req.Header.Set("X-Kite-Version", "3")
	req.Header.Set("Authorization", "token "+apiKey+":"+accessToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", "_cfuvid=97BIlEcAGbGruFDFpLQ94u80VV5UtEPRuvehdZTxLyg-1702441985795-0-604800000")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return GTTOrerResponse{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return GTTOrerResponse{}, err
	}
	fmt.Println(string(body))

	var responseObject GTTOrerResponse
	err = json.Unmarshal([]byte(body), &responseObject)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return GTTOrerResponse{}, err
	}

	fmt.Println(responseObject.Data.TriggerID)
	fmt.Println(responseObject.Status)
	return responseObject, nil
}

func encodeValues(data map[string]string) string {
	values := url.Values{}
	for key, value := range data {
		values.Add(key, value)
	}
	return values.Encode()
}

func convertMapToStringMap(input map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for key, value := range input {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}

func DeleteGTTOrderById(gttOrderId string) {
	apiKey := os.Getenv("API_KEY")
	accessToken := os.Getenv("ACCESS_TOKEN")
	url := DeleteGTTOrder + gttOrderId
	method := "DELETE"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("X-Kite-Version", "3")
	req.Header.Add("Authorization", "token "+apiKey+":"+accessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

func GetAllGTTOrder() (GTTOrder, error) {
	apiKey := os.Getenv("API_KEY")
	accessToken := os.Getenv("ACCESS_TOKEN")

	url := "https://api.kite.trade/gtt/triggers"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return GTTOrder{}, err
	}
	req.Header.Add("X-Kite-Version", "3")
	req.Header.Add("Authorization", "token "+apiKey+":"+accessToken)
	req.Header.Add("Cookie", "_cfuvid=_XS4.FIRP8ctFbBIUJFSDCqm3ZusAKg_tINeMln5uw8-1702625462969-0-604800000")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return GTTOrder{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return GTTOrder{}, err
	}
	fmt.Println(string(body))

	var gttOrder GTTOrder

	if err := json.Unmarshal(body, &gttOrder); err != nil {
		fmt.Println("Error mapping JSON to struct:", err)
		return GTTOrder{}, err

	}
	return gttOrder, nil
}
