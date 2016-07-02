package smpp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/LeKovr/go-base/logger"
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
func IsBalanceOk(cfg *Flags, log *logger.Log) (ok bool, err error) {

	if cfg.BalFormat == "" {
		return true, nil
	}
	url := fmt.Sprintf(cfg.BalFormat, cfg.SmppID, cfg.BalKey)

	response, err := http.Get(url)
	if err != nil {
		log.Errorf("Resp error: %+v", err)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Read resp error: %+v", err)
		return
	}

	var d bhResp
	err = json.Unmarshal(body, &d)
	if err != nil {
		log.Errorf("Parse resp error: %+v", err)
		return
	}
	// log.Printf("BAL JSON: %+v\n", d)
	if d.Status != 0 {
		log.Errorf("Response error: %+v", d)
		return
	}
	num, err := strconv.ParseFloat(d.Data, 64)
	log.Infof("Balance: %.2f\n", num)
	ok = num >= float64(cfg.MinBalance)
	return
}
