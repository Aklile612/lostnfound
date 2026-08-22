package domain

import "time"

type ReportType string

const (
	TypeLost  ReportType = "lost"
	TypeFound ReportType = "found"
)

type Status string

const (
	StatusOpen      Status = "open"
	StatusUnclaimed Status = "unclaimed"
	StatusClaimed   Status = "claimed"
	StatusResolved  Status = "resolved"
)

type Report struct {
	ID              string
	Type            ReportType
	Category        string
	Title           string
	Description     string
	UniqueFeatures  string
	Location        string
	LocationDetails string
	IncidentDate    time.Time
	Photos          []string
	Phone           string
	Telegram        string
	Status          Status
	CreatedAt       time.Time
}

func (r Report) HasPhotos() bool {
	return len(r.Photos) > 0
}

var Categories = map[string]string{
	"electronics": "Electronics",
	"bags":        "Bags",
	"ids":         "IDs & Documents",
	"cards":       "ATM & Bank Cards",
	"keys":        "Keys",
	"clothing":    "Clothing",
	"jewelry":     "Jewelry",
	"books":       "Books",
	"other":       "Other",
}

var Locations = map[string]string{
	"cafeteria":      "Cafeteria",
	"library":        "Library",
	"coffee_shop":    "Coffee shop",
	"lecture_hall":   "Lecture hall",
	"dormitory":      "Dormitory",
	"gym":            "Gym",
	"football_field": "Football field",
	"parking":        "Parking",
	"student_center": "Student center",
	"other":          "Other",
}

func ValidCategory(v string) bool {
	_, ok := Categories[v]
	return ok
}

func ValidLocation(v string) bool {
	_, ok := Locations[v]
	return ok
}

func NearbyLocations(location string) []string {
	related := map[string][]string{
		"cafeteria":      {"coffee_shop", "student_center"},
		"coffee_shop":    {"cafeteria", "student_center"},
		"library":        {"student_center", "lecture_hall"},
		"lecture_hall":   {"library", "student_center"},
		"student_center": {"cafeteria", "coffee_shop", "library"},
		"dormitory":      {"parking"},
		"parking":        {"dormitory", "football_field"},
		"football_field": {"parking", "gym"},
		"gym":            {"football_field"},
	}
	return related[location]
}
