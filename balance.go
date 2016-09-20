package smpp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// smpp_bal="http://bytehand.com:3800/balance?id=%s&key=%s"

// -----------------------------------------------------------------------------

type bhResp struct {
	Status int    `json:"status"`
	Data   string `json:"description"`
}

// -----------------------------------------------------------------------------

// IsBalanceOk checks smsc balance via http API if BalFormat config given
// code tested only with bytehand.com provider
func IsBalanceOk(cfg *Flags, logger *log.Logger) (ok bool, err error) {

	if cfg.BalFormat == "" {
		return true, nil
	}
	url := fmt.Sprintf(cfg.BalFormat, cfg.SmppID, cfg.BalKey)

	response, err := http.Get(url)
	if err != nil {
		logger.Printf("error: Resp error: %+v", err)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Printf("error: Read resp error: %+v", err)
		return
	}

	var d bhResp
	err = json.Unmarshal(body, &d)
	if err != nil {
		logger.Printf("error: Parse resp error: %+v", err)
		return
	}
	// log.Printf("BAL JSON: %+v\n", d)
	if d.Status != 0 {
		logger.Printf("error: Response error: %+v", d)
		return
	}
	num, err := strconv.ParseFloat(d.Data, 64)
	logger.Printf("info: Balance: %.2f\n", num)
	ok = num >= float64(cfg.MinBalance)
	return
}
