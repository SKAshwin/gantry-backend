package postgres_test

import (
	"checkin"
	"checkin/postgres"
	"checkin/test"
	"testing"
)

func TestGuestSite(t *testing.T) {
	gss := postgres.GuestSiteService{DB: db}

	//test expected behavior
	expectedSite := checkin.GuestSite{
		Details: checkin.DetailsSection{
			Logo: "",
			Title: checkin.TextElement{
				Content: "title",
				Size:    checkin.TextSize(2),
			},
			Tagline: checkin.TextElement{
				Content: "tagline",
				Size:    checkin.TextSize(5),
			},
			Items: []checkin.DetailItem{
				checkin.DetailItem{
					Title:   "Date",
					Content: "1 Jan 2020",
				},
				checkin.DetailItem{
					Title:   "Time",
					Content: "1300 - 1700",
				},
				checkin.DetailItem{
					Title:   "Venue",
					Content: "Kranji Camp 3 Blk 822, 90 Choa Chu Kang Way, Singapore 688264",
				},
				checkin.DetailItem{
					Title:   "Attire",
					Content: "Office Wear",
				},
			},
		},
		Main: checkin.ButtonSection{
			Icon: true,
			Schedule: checkin.ScheduleSection{
				Display: true,
				Pages: []checkin.SchedulePage{
					checkin.SchedulePage{
						Subject: "Menu A",
						Items: []checkin.ScheduleItem{
							checkin.ScheduleItem{
								Topic: "Item A",
								Start: checkin.TimeElement{},
								End:   checkin.TimeElement{},
								Abstract: checkin.OptionalContent{
									Display: true,
									Content: "hello",
								},
								Speaker: checkin.OptionalProfileItem{},
							},
							checkin.ScheduleItem{
								Topic: "Item B",
								Start: checkin.TimeElement{Hour: 12, Minute: 0},
								End:   checkin.TimeElement{Hour: 14, Minute: 0},
								Abstract: checkin.OptionalContent{
									Display: true,
									Content: "hello",
								},
								Speaker: checkin.OptionalProfileItem{
									Display: true,
									Content: checkin.ProfileItem{
										Image: "",
										Name:  "name",
										Title: "title",
										About: "about",
									},
								},
							},
						},
					},
					checkin.SchedulePage{
						Subject: "Menu B",
						Items: []checkin.ScheduleItem{
							checkin.ScheduleItem{
								Topic: "Item A",
								Start: checkin.TimeElement{},
								End:   checkin.TimeElement{},
								Abstract: checkin.OptionalContent{
									Display: false,
									Content: "hello",
								},
								Speaker: checkin.OptionalProfileItem{
									Display: true,
								},
							},
							checkin.ScheduleItem{
								Topic: "Item B",
								Start: checkin.TimeElement{Hour: 12, Minute: 0},
								End:   checkin.TimeElement{Hour: 14, Minute: 0},
								Abstract: checkin.OptionalContent{
									Display: false,
									Content: "hello",
								},
								Speaker: checkin.OptionalProfileItem{
									Display: false,
									Content: checkin.ProfileItem{
										Image: "",
										Name:  "name",
										Title: "title",
										About: "about",
									},
								},
							},
						},
					},
				},
			},
			Size: checkin.ButtonSize(4),
			ButtonRows: checkin.ButtonMatrix([]checkin.ButtonColumn{
				checkin.ButtonColumn{
					Buttons: []checkin.ButtonElement{
						checkin.ButtonElement{
							Title: "Link",
							Type:  checkin.ButtonType("link"),
							Icon:  "a1a.svg",
							Link:  "https://www.google.com",
							PopUp: checkin.PopUp([]checkin.PopUpComponent{
								checkin.PopUpComponent{
									Type:  checkin.PopUpComponentType("text"),
									Text:  "",
									Image: "",
								},
							}),
						},
						checkin.ButtonElement{
							Title: "Popup",
							Type:  checkin.ButtonType("popup"),
							Icon:  "a1b.svg",
							Link:  "",
							PopUp: checkin.PopUp([]checkin.PopUpComponent{
								checkin.PopUpComponent{
									Type:  checkin.PopUpComponentType("text"),
									Text:  "Hello",
									Image: "",
								},
							}),
						},
					},
				},
			}),
		},
		Survey: checkin.SurveySection([]checkin.QuestionElement{
			checkin.QuestionElement{
				Type:     checkin.QuestionType("scaled"),
				Question: "The length of the event was just nice.",
			},
			checkin.QuestionElement{
				Type:     checkin.QuestionType("rating"),
				Question: "How would you rate this event?",
			},
			checkin.QuestionElement{
				Type:     checkin.QuestionType("radio"),
				Question: "Is this your first time attending this event?",
			},
			checkin.QuestionElement{
				Type:     checkin.QuestionType("open"),
				Question: "Any other comments.",
			},
		}),
	}
	site, err := gss.GuestSite("2c59b54d-3422-4bdb-824c-4125775b44c8")
	test.Ok(t, err)
	test.Equals(t, site, expectedSite)

	//test event exists but no guest site
	site, err = gss.GuestSite("c14a592c-950d-44ba-b173-bbb9e4f5c8b4")
	test.Assert(t, err != nil, "No error thrown when trying to get non existent site")

	//test event does not exist
	site, err = gss.GuestSite("45d48803-a8b3-47e4-8670-b099fd0b7f29")
	test.Assert(t, err != nil, "No error thrown when trying to get site from non-existent event")

	//test event invalid UUID
	site, err = gss.GuestSite("lol")
	test.Assert(t, err != nil, "No error thrown when trying to get site from non-existent event (invalid UUID)")

	//test empty string
	//test event invalid UUID
	site, err = gss.GuestSite("")
	test.Assert(t, err != nil, "No error thrown when trying to get site from non-existent event (empty strings)")
}

func TestCreateGuestSite(t *testing.T) {
	gss := postgres.GuestSiteService{DB: db}

	//test normal functionality
	site := checkin.GuestSite{
		Details: checkin.DetailsSection{
			Logo: "something.jpg",
			Title: checkin.TextElement{
				Content: "CSSCOM iWPS",
				Size:    checkin.TextSize(7),
			},
			Tagline: checkin.TextElement{
				Content: "2019",
				Size:    checkin.TextSize(5),
			},
		},
		Main: checkin.ButtonSection{
			Size: checkin.ButtonSize(2),
		},
	}
	err := gss.CreateGuestSite("03293b3b-df83-407e-b836-fb7d4a3c4966", site)
	test.Ok(t, err)
	fetched, err := gss.GuestSite("03293b3b-df83-407e-b836-fb7d4a3c4966")
	test.Ok(t, err)
	test.Equals(t, site, fetched)
	err = gss.DeleteGuestSite("03293b3b-df83-407e-b836-fb7d4a3c4966") //clear out website
	test.Ok(t, err)

	//test that empty sites work fine
	err = gss.CreateGuestSite("03293b3b-df83-407e-b836-fb7d4a3c4966", checkin.GuestSite{})
	test.Ok(t, err)
	fetched, err = gss.GuestSite("03293b3b-df83-407e-b836-fb7d4a3c4966")
	test.Ok(t, err)
	test.Equals(t, checkin.GuestSite{}, fetched)

	//testing attempting to give one event multiple guest sites
	err = gss.CreateGuestSite("2c59b54d-3422-4bdb-824c-4125775b44c8", checkin.GuestSite{})
	test.Assert(t, err != nil, "No error thrown when trying to create multiple sites for an event")

	//test event does not exist
	err = gss.CreateGuestSite("f031996b-c049-45d6-ba18-3bd379fc8a7c", site)
	test.Assert(t, err != nil, "No error thrown when trying to create site for event that does not exist")

	//test event does not exist (invalid UUID)
	err = gss.CreateGuestSite("hello", site)
	test.Assert(t, err != nil, "No error thrown when trying to create site for event that does not exist (invalid UUID)")
}

func TestUpdateGuestSite(t *testing.T) {
	gss := postgres.GuestSiteService{DB: db}

	//test normal functionality
	original, err := gss.GuestSite("2c59b54d-3422-4bdb-824c-4125775b44c8")
	test.Ok(t, err)
	new := original
	new.Main.Schedule.Pages[0].Subject = "New Subject"
	err = gss.UpdateGuestSite("2c59b54d-3422-4bdb-824c-4125775b44c8", new)
	test.Ok(t, err)
	fetched, err := gss.GuestSite("2c59b54d-3422-4bdb-824c-4125775b44c8")
	test.Ok(t, err)
	test.Equals(t, new, fetched)

	//set back to original
	err = gss.UpdateGuestSite("2c59b54d-3422-4bdb-824c-4125775b44c8", original)
	test.Ok(t, err)

	//Test event does not have an existing website
	err = gss.UpdateGuestSite("03293b3b-df83-407e-b836-fb7d4a3c4966", new)
	test.Assert(t, err != nil, "No error when trying to update non-existent site of event")

	//Test event does not exist
	err = gss.UpdateGuestSite("8062e111-a861-470a-8d86-922cb58db1f8", new)
	test.Assert(t, err != nil, "No error when trying to update non-existent event's site")

	//Test event does not exist (invalid UUID)
	err = gss.UpdateGuestSite("literally anything", new)
	test.Assert(t, err != nil, "No error when trying to update non-existent event's site (invalid UUID)")

}

func TestDeleteGuestSite(t *testing.T) {
	gss := postgres.GuestSiteService{DB: db}

	//test normal functionality
	_, err := gss.GuestSite("3820a980-a207-4738-b82b-45808fe7aba8")
	test.Ok(t, err)
	err = gss.DeleteGuestSite("3820a980-a207-4738-b82b-45808fe7aba8")
	test.Ok(t, err)
	_, err = gss.GuestSite("3820a980-a207-4738-b82b-45808fe7aba8")
	test.Assert(t, err != nil, "No error trying to fetch guest site after its deletion")

	//try deleting non-existent guest site (of event that exists), should return no error
	err = gss.DeleteGuestSite("03293b3b-df83-407e-b836-fb7d4a3c4966")
	test.Ok(t, err)

	//try deleting guest site of non-existent event, should not throw an error
	err = gss.DeleteGuestSite("7d9ccc97-ae6c-472d-9696-c7bb61eff8bd")
	test.Ok(t, err)
}

func TestGuestSiteExists(t *testing.T) {
	gss := postgres.GuestSiteService{DB: db}

	//test normal functionality, guest site exists and doesn't
	exists, err := gss.GuestSiteExists("2c59b54d-3422-4bdb-824c-4125775b44c8")
	test.Ok(t, err)
	test.Equals(t, true, exists)

	exists, err = gss.GuestSiteExists("c14a592c-950d-44ba-b173-bbb9e4f5c8b4")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//try event doesn't exist (should return false)
	exists, err = gss.GuestSiteExists("0484d402-08b6-4c95-aa4d-99e85092a326")
	test.Ok(t, err)
	test.Equals(t, false, exists)

	//try event does not exist (invalid UUID), should return false
	exists, err = gss.GuestSiteExists("helloMyDude")
	test.Ok(t, err)
	test.Equals(t, false, exists)
}
