package checkin

//This file contains the structs use to represent the website structure built using the website builder
//With their JSON marshaller implementations which enforce requirements on them

//TODO Base64 string and validate in marshalling

import (
	"encoding/json"
	"reflect"
	"strconv"
)

//GuestSite represents the content and styling of the guest website
type GuestSite struct {
	Details DetailsSection `json:"details"`
	Main    ButtonSection  `json:"main"`
	Survey  SurveySection  `json:"survey"`
}

//DetailsSection is the section of the page with all the event details
type DetailsSection struct {
	Logo    string       `json:"logo"`
	Title   TextElement  `json:"title"`
	Tagline TextElement  `json:"tagline"`
	Items   []DetailItem `json:"items"`
}

//ButtonSection represents the bottom part of the page, with the buttons
type ButtonSection struct {
	Icon      bool            `json:"icon"`
	Schedule  ScheduleSection `json:"schedule"`
	Size      ButtonSize      `json:"sz"`
	ButtonRow ButtonMatrix    `json:"rows"`
}

//SurveySection represents a collection of questions that make up the event survey
type SurveySection []QuestionElement

//QuestionElement represents a single question and its input type
type QuestionElement struct {
	Type     QuestionType `json:"type"`
	Question string       `json:"question"`
}

//ScheduleSection represents the portion of the website with the event schedule information
type ScheduleSection struct {
	Display bool           `json:"check"`
	Pages   []SchedulePage `json:"menus"`
}

//SchedulePage refers to a page of the schedule
type SchedulePage struct {
	Subject string         `json:"subject"`
	Items   []ScheduleItem `json:"items"`
}

//ScheduleItem is a particular part of the schedule, which has a start and end time, and some topic, and possibly further information like
//abstract and speaker information
type ScheduleItem struct {
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
	Display bool   `json:"check"`
	Content string `json:"cont"`
}

//OptionalProfileItem represents a ProfileItem that may or may not exist
type OptionalProfileItem struct {
	Display bool        `json:"check"`
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
	Content string   `json:"cont"`
	Size    TextSize `json:"sz"`
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
type ButtonColumn struct {
	Buttons []ButtonElement `json:"cols"`
}

//ButtonElement represents a button, whih has a size, some text, a type (can link to either an external webpage
//or to a pop up modal), and a content, which is either the URL of the external page or the text in the modal,
//depending on ButtonType
type ButtonElement struct {
	Icon  string     `json:"icon"`
	Title string     `json:"title"`
	Type  ButtonType `json:"type"`
	Link  string     `json:"link"` //consider using markdown
	PopUp PopUp      `json:"popup"`
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
	if len(bytes) < 2 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(bytes) + " is not a valid  pop up component type"}
	}
	input := PopUpComponentType(bytes[1 : len(bytes)-1]) //to remove the quotes
	switch input {
	case Text:
		*pt = Text
		break
	case Image:
		*pt = Image
		break
	default:
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(bytes) + " is not a valid  pop up component type"}
	}

	return nil
}

//ButtonType a button in the website can either link to an external webpage, or bring a pop-up modal
type ButtonType string

const (
	//Link means that the button embeds a URL to another
	Link ButtonType = "link"
	//Modal means that the button triggers a pop up with some text
	Modal ButtonType = "popup"
)

//UnmarshalJSON validates that the button type is either link or modal, or returns and error
func (bt *ButtonType) UnmarshalJSON(bytes []byte) error {
	if len(bytes) < 2 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(bytes) + " is not a valid button type"}
	}
	input := ButtonType(bytes[1 : len(bytes)-1]) //to remove the quotes

	switch input {
	case Link:
		*bt = Link
		break
	case Modal:
		*bt = Modal
		break
	default:
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(bytes) + " is not a valid button type"}
	}

	return nil
}

//QuestionType tells the input of the question
type QuestionType string

const (
	//Scaled refers to a qualitative scaled input (ie Strongly Disagree to Agree etc)
	Scaled QuestionType = "scaled"
	//Rating refers to a quantitative slider input (i.e. 1 to 5)
	Rating QuestionType = "rating"
	//RadioButton refers to a radio button input
	RadioButton QuestionType = "radio"
	//OpenEnded refers to open ended input (like a text box)
	OpenEnded QuestionType = "open"
)

//UnmarshalJSON unmarshals JSON input into a QuestionType, making sure its one of Scaled, Rating, RadioButton or OpenEnded
func (qt *QuestionType) UnmarshalJSON(bytes []byte) error {
	if len(bytes) < 2 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(bytes) + " is not a valid question type"}
	}
	input := QuestionType(bytes[1 : len(bytes)-1]) //to remove the quotes

	switch input {
	case Scaled:
		*qt = Scaled
		break
	case Rating:
		*qt = Rating
		break
	case RadioButton:
		*qt = RadioButton
		break
	case OpenEnded:
		*qt = OpenEnded
		break
	default:
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: string(bytes) + " is not a valid question type"}
	}

	return nil
}

//TextSize represents an text elements sizing instruction
//and it should be between 1 and 7
type TextSize uint8

//ButtonSize represents a button's sizing instruction
//Should be between 1 and 4
type ButtonSize uint8

//Hour represents a 24 hour time
type Hour uint8

//Minute represents the minute component of time (between 0 and 59)
type Minute uint8

//UnmarshalJSON validates the bytes, making sure its first a valid number, and then between 0 and 23
func (h *Hour) UnmarshalJSON(bytes []byte) error {
	var number uint8
	err := json.Unmarshal(bytes, &number)
	if err != nil {
		return err
	}
	if number > 23 || number < 0 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: strconv.Itoa((int)(number)) + " is not a valid hour (must be between 0 and 23)"}
	}
	*h = Hour(number)
	return nil
}

//UnmarshalJSON validates the bytes, making sure its first a valid number, and then between 0 and 59
func (m *Minute) UnmarshalJSON(bytes []byte) error {
	var number uint8
	err := json.Unmarshal(bytes, &number)
	if err != nil {
		return err
	}
	if number > 59 || number < 0 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: strconv.Itoa((int)(number)) + " is not a valid minute (must be between 0 and 59)"}
	}
	*m = Minute(number)
	return nil
}

//UnmarshalJSON validates the bytes, making sure its first a valid number, and then between 1 and 7
func (s *TextSize) UnmarshalJSON(bytes []byte) error {
	var number uint8
	err := json.Unmarshal(bytes, &number)
	if err != nil {
		return err
	}
	if number > 7 || number < 1 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: strconv.Itoa((int)(number)) + " is not a valid text size (must be between 1 and 7)"}
	}
	*s = TextSize(number)
	return nil
}

//UnmarshalJSON validates the bytes, making sure its first a valid number, and then between 1 and 4
func (s *ButtonSize) UnmarshalJSON(bytes []byte) error {
	var number uint8
	err := json.Unmarshal(bytes, &number)
	if err != nil {
		return err
	}
	if number > 4 || number < 1 {
		return &json.UnsupportedValueError{Value: reflect.ValueOf(bytes),
			Str: strconv.Itoa((int)(number)) + " is not a valid button size (must be between 1 and 4)"}
	}
	*s = ButtonSize(number)
	return nil
}
