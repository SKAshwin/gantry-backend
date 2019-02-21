package qrcode

import (
	"github.com/skip2/go-qrcode"
)

//Generator implements checkin.QRGenerator using go-qrcode,
//given a RecoveryLevel, gives a method to generate QRCodes
type Generator struct {
	Level RecoveryLevel
}

//RecoveryLevel is the error detection/recovery capacity
//of the QR code
//Higher levels of error recovery are able to correct more errors,
//with the trade-off of increased symbol size.
type RecoveryLevel int

const (
	//Low 7% error recovery.
	Low RecoveryLevel = iota

	//Medium 15% error recovery. Good default choice.
	Medium

	//High 25% error recovery.
	High

	//Highest 30% error recovery.
	Highest
)

//Encode generates a PNG qr code given a message and size.
func (qrg Generator) Encode(msg string, size int) ([]byte, error) {
	return qrcode.Encode(msg, qrcode.RecoveryLevel(qrg.Level), 256)
}
