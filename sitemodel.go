package checkin

//This file contains the structs use to represent the website structure built using the website builder
//With their JSON marshaller implementations which enforce requirements on them

//TODO Base64 string and validate in marshalling

import (
	"encoding/json"
	"reflect"
)

//Size represents an elements sizing instruction
//and it should be between 0 and 7
//TODO write special marshaller to enforce this constraint
type Size uint8

//GuestSite represents the content and styling of the guest website
type GuestSite struct {
	Details DetailsSection `json:"details"`
	Main    MainSection    `json:"main"`
	Survey  SurveySection  `json:"survey"`
}

//DetailsSection is the section of the page with all the event details
type DetailsSection struct {
	Logo    ImageElement `json:"logo"`
	Title   TextElement  `json:"title"`
	Tagline TextElement  `json:"tagline"`
	Items   []DetailItem `json:"items"`
}

//MainSection represents the main part of the page
//TODO ask what this means
type MainSection struct {
	Icon      bool            `json:"icon"`
	Schedule  ScheduleSection `json:"schedule"`
	Size      Size            `json:"sz"` //TODO confirm this is limited between 0 and 7
	ButtonRow ButtonMatrix    `json:"rows"`
}

//SurveySection represents a collection of questions that make up the event survey
type SurveySection []QuestionElement

//QuestionElement represents a single question and its input type
type QuestionElement struct {
	Type     string `json:"type"` //TODO assign special type when you work out the possibilities
	Question string `json:"question"`
}

//ScheduleSection represents the portion of the website with the event schedule information
type ScheduleSection struct {
	Check bool          `json:"check"` //TODO ask what this means
	Menus []MenuSection `json:"menus"`
}

//MenuSection represents a menu from which the schedule can be selected
//TODO ask what this means
type MenuSection struct {
	Subject string     `json:"subject"`
	Items   []MenuItem `json:"items"`
}

type MenuItem struct {
	Topic    string              `json:"topic"`
	Start    TimeElement         `json:"start"`
	End      TimeElement         `json:"end"`
	Abstract OptionalContent     `json:"abstract"`
	Speaker  OptionalProfileItem `json:"speaker"`
}

//TimeElement refers to a HH:MM time displayed
type TimeElement struct {
	Hour   Hour   `json:"hour"`
	Minute Minute `json:"mins"`
}

//OptionalContent refers to text component that may or may not exist
type OptionalContent struct {
	Check   bool   `json:"check"`
	Content string `json:"cont"`
}

//OptionalProfileItem represents a ProfileItem that may or may not exist
type OptionalProfileItem struct {
	Check   bool        `json:"check"`
	Content ProfileItem `json:"cont"`
}

//ProfileItem refers to a component that profiles a particular individual
type ProfileItem struct {
	Image string `json:"img"`
	Name  string `json:"name"`
	Title string `json:"title"`
	About string `json:"about"`
}

//TextElement some words which have a size and some content
type TextElement struct {
	Content string `json:"cont"`
	Size    Size   `json:"sz"`
}

//ImageElement refers to an image
type ImageElement struct {
	Type    string `json:"type"` //TODO ask Henry what the heck type means
	Content string `json:"cont"` //the Base64 encoding of the ImageElement
}

//DetailItem refers to structure providing info about some event attribute
//For example, Venue: KC3, will be a DetailItem with title Venue and Content KC3
type DetailItem struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

//ButtonMatrix represents a row of button columns
type ButtonMatrix []ButtonColumn

//ButtonColumn represents a row of buttons
type ButtonColumn []ButtonElement

//ButtonElement represents a button, whih has a size, some text, a type (can link to either an external webpage
//or to a pop up modal), and a content, which is either the URL of the external page or the text in the modal,
//depending on ButtonType
type ButtonElement struct {
	Icon  string     `json:"icon"`
	Title string     `json:"title"`
	Type  ButtonType `json:"type"`
	Link  string     `json:"link"` //consider using markdown
	PopUp PopUp      `json:"popup`
}

//PopUp a pop up consists of many components
type PopUp []PopUpComponent

//PopUpComponent represents part of the content of a pop-up
type PopUpComponent struct {
	Type  PopUpComponentType `json:"type"`
	Text  string             `json:"text"`
	Image string             `json:"img"`
}

//PopUpComponentType represents the type of a pop up component
type PopUpComponentType string

const (
	//Text means that the content of a pop up type is to be interpreted as text
	Text PopUpComponentType = "text"

	//Image means it is to be interpreted as a Base64 encoding of an Image
	Image PopUpComponentType = "img"
)

//UnmarshalJSON validates that the pop up type is either text or image, or returns an error
func (pt *PopUpComponentType) UnmarshalJSON(bytes []byte) error {
	input := PopUpComponentType(bytes)

	switch input {
	case Text:
		*pt = Text
		break
	case Image:
		*pt = Image
		break
	default:
		//leave current PopUpComponentType unchanged
		if *pt != Text && *pt != Image { //but if the pop-up component type is currently a invalid value, throw an error
			return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
				Str: string(input) + " is not a valid button type"}
		}

		return nil
	}

	return nil
}

//ButtonType a button in the website can either link to an external webpage, or bring a pop-up modal
type ButtonType string

const (
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
		//leave current ButtonType unchanged
		if *bt != Link && *bt != Modal { //but if the button type is currently the invalid value, throw an error
			return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
				Str: string(input) + " is not a valid button type"}
		}

		return nil
	}

	return nil
}

//Hour represents a 24 hour time
type Hour uint8 //TODO custom marshaller
//Minute represents the minute component of time (between 0 and 59)
type Minute uint8 //TODO custom marshaller

//UnmarshalJSON validates the bytes, making sure its first a valid number, and then between 0 and 23
func (h *Hour) UnmarshalJSON(bytes []byte) error {
	var number uint8
	err := json.Unmarshal(bytes, &number)
	if err != nil {
		return err
	}
	if number > 23 || number < 0 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(number) + " is not a valid hour (must be between 0 and 23)"}
	}
	return nil
}

//UnmarshalJSON validates the bytes, making sure its first a valid number, and then between 0 and 59
func (h *Minute) UnmarshalJSON(bytes []byte) error {
	var number uint8
	err := json.Unmarshal(bytes, &number)
	if err != nil {
		return err
	}
	if number > 59 || number < 0 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(number) + " is not a valid hour (must be between 0 and 59)"}
	}
	return nil
}
