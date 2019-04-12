package checkin

import (
	"encoding/json"
	"reflect"
)

//Size represents an elements sizing instruction
//and it should be between 0 and 7
type Size uint8

//TextElement some words which have a size and some content
type TextElement struct {
	Content string `json:"cont"`
	Size    Size   `json:"sz"`
}

//DetailsElement refers to structure providing info about some event attribute
//For example, Venue: KC3, will be a DetailsElement with title Venue and Content KC3
type DetailsElement struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

//ButtonType a button in the website can either link to an external webpage, or bring a pop-up modal
type ButtonType string

const (
	//Unknown is the fall back empty ButtonType, implies something went wrong
	Unknown ButtonType = ""
	//Link means that the button embeds a URL to another
	Link ButtonType = "link"
	//Modal means that the button triggers a pop up with some text
	Modal ButtonType = "modal"
)

//UnmarshalJSON validates that the button type is either link or modal, or returns and error
func (bt *ButtonType) UnmarshalJSON(bytes []byte) error {
	input := ButtonType(bytes)

	switch input {
	case Link:
		*bt = Link
		break
	case Modal:
		*bt = Modal
		break
	default:
		*bt = Unknown
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(input) + " is not a valid button type"}
	}

	return nil
}

//ButtonElement represents a button, whih has a size, some text, a type (can link to either an external webpage
//or to a pop up modal), and a content, which is either the URL of the external page or the text in the modal,
//depending on ButtonType
type ButtonElement struct {
	Size    Size       `json:"sz"`
	Title   string     `json:"title"`
	Type    ButtonType `json:"type"`
	Content string     `json:"cont"` //consider using markdown
}

//ButtonRow represents a row of buttons
type ButtonRow []ButtonElement

//GuestSite represents the content and styling of the guest website
type GuestSite struct {
	Title   TextElement      `json:"title"`
	Tagline TextElement      `json:"tagline"`
	LogoURL string           `json:"logo"`
	Details []DetailsElement `json:"details"`
	Buttons []ButtonRow      `json:"buttons"`
}
