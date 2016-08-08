// Package smpp allows sending SMS via SMPP server
package smpp

import (
	"fmt"
	"strings"

	"github.com/CodeMonkeyKevin/smpp34"

	"github.com/LeKovr/go-base/logger"
)

// Params - attributes for SubmitSm
var Params = &smpp34.Params{
	"registered_delivery": 1,
}

// -----------------------------------------------------------------------------

// Flags is a package flags sample
// in form ready for use with github.com/jessevdk/go-flags
type Flags struct {
	SmppHost string `long:"smpp_host"                description:"SMPP server ip (default: do not send SMS)"`
	SmppPort int    `long:"smpp_port" default:"3200" description:"SMPP server port"`
	SmppID   string `long:"smpp_id"                  description:"SMPP user id"`
	SmppPass string `long:"smpp_pass"                description:"SMPP user key"`

	SmppFrom          string `long:"smpp_from"   default:"a.elfire.ru"     description:"SMPP message signature"`
	SmppMessageFormat string `long:"smpp_msg"    default:"Access code: %s" description:"SMPP message format"`
	SmppPhonePrefix   string `long:"smpp_prefix" default:"+7"              description:"SMPP phone prefix"`

	BalFormat  string `long:"smpp_bal"               description:"SMPP account balance request format"`
	BalKey     string `long:"smpp_key"               description:"SMPP account balance request key"`
	MinBalance int    `long:"smpp_minbalance" default:"1" description:"Do not send SMS if balance is lower than"`
}

// -----------------------------------------------------------------------------

// Send sends given code to phone number via SMS
func Send(cfg *Flags, log *logger.Log, phone, code string) error {

	if cfg.SmppHost == "" || cfg.SmppHost == "none" {
		// log only mode
		log.Infof("Log SMS to %s with code %s", phone, code)
		return nil
	}
	log.Infof("Send SMS to %s with code %s", phone, code)
	// connect and bind

	//log.Printf("Connect %s /%s", cfgSmppId, pass)

	trx, err := smpp34.NewTransceiver(
		cfg.SmppHost, // ip
		cfg.SmppPort, // port
		60,           // EnquireLink interval, in seconds.
		smpp34.Params{
			"system_type": "WWW",
			"system_id":   cfg.SmppID,
			"password":    cfg.SmppPass,
		},
	)
	if err != nil {
		log.Error("Connection Err:", err)
		return err
	}

	// Send SubmitSm (source_addr, destination_addr, short_message string..
	from := cfg.SmppFrom
	msg := fmt.Sprintf(cfg.SmppMessageFormat, code)
	//	seq, err := trx.SubmitSm(from, "+79184304451", "Access code is E45Rew4", &smpp34.Params{"registered_delivery": 1})
	seq, err := trx.SubmitSm(from, cfg.SmppPhonePrefix+phone, msg, Params)
	// Should save seq to match with message_id

	// Pdu gen errors
	if err != nil {
		log.Errorf("SubmitSm to %s err: %v", phone, err)
	}

	// start reading PDUs
	for i := 0; i < 4; i++ { // not more than 5 times
		pdu, err := trx.Read() // This is blocking
		if err != nil {
			break
		}

		// Transceiver auto handles EnquireLinks
		switch pdu.GetHeader().Id {
		case smpp34.SUBMIT_SM_RESP:
			// message_id should match this with seq message
			// fmt.Println("MSG ID:", pdu.GetField("message_id").Value())
		case smpp34.DELIVER_SM:
			// received Deliver Sm

			// Respond back to Deliver SM with Deliver SM Resp
			err := trx.DeliverSmResp(pdu.GetHeader().Sequence, smpp34.ESME_ROK)

			// destination_addr : +79184304451
			// short_message : id:119609566430770966 sub:001 dlvrd:001 submit date:1510160852 done date:1510160853 stat:DELIVRD err:000 text:a
			f := pdu.GetField("short_message")
			if strings.Contains(f.String(), " stat:ACCEPTD ") {
				// just accept
			} else if strings.Contains(f.String(), " stat:DELIVRD ") {
				log.Infof("Delivered SMS to %s", phone)
				return nil
			} else {
				// someting wrong
				log.Warningf("Unknown SMPP (seq %v) resp: %s", seq, f)
				return nil
			}
			// Look at DeliverSmResp err
			if err != nil {
				log.Errorf("DeliverSmResp err: %+v", err)
			}

			log.Debugf("Got DeliverSm: %s", f)
		default:
			log.Debugf("PDU ID:", pdu.GetHeader().Id)
		}
	}

	//fmt.Println("ending...")
	return nil
}
