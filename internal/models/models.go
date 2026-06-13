package models

import "time"

type Dog struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Born   string `json:"born"`
	Gender string `json:"gender"`
}

type Buyer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

type Purchase struct {
	Date  string `json:"date"`
	Price int    `json:"price"`
}

type Puppy struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Farbe        Fellfarbe `json:"farbe"`
	Geburtsdatum string    `json:"geburtsdatum"`
	Geschlecht   string    `json:"geschlecht"`
	Gewicht      float64   `json:"gewicht"`
	Charakter    string    `json:"charakter"`
	Geimpft      bool      `json:"geimpft"`
	Gechippt     bool      `json:"gechippt"`
	Entwurmt     bool      `json:"entwurmt"`
	Eltern       []string  `json:"eltern"`
	Notizen      string    `json:"notizen"`
	Bilder       []string  `json:"bilder"`
}

type Contact struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
	MailSent  bool      `json:"mailSent"`
	MailError string    `json:"mailError,omitempty"`
}

// Fellfarbe für Labrador Retriever.
type Fellfarbe string

const (
	FellfarbeUnbekannt Fellfarbe = ""
	FarbeSchwarz       Fellfarbe = "schwarz"
	FarbeGelb          Fellfarbe = "gelb"
	FarbeBraun         Fellfarbe = "braun"
	FarbeFoxRed        Fellfarbe = "fox red"
	FarbeSilber        Fellfarbe = "silber"
	FarbeChampagner    Fellfarbe = "champagner"
	FarbeCharcoal      Fellfarbe = "charcoal"
)

func IstGueltigeFarbe(f Fellfarbe) bool {
	switch f {
	case FarbeSchwarz, FarbeGelb, FarbeBraun, FarbeFoxRed, FarbeSilber, FarbeChampagner, FarbeCharcoal:
		return true
	default:
		return false
	}
}
