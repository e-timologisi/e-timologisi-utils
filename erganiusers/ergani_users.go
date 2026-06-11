package erganiusers

import (
	"strings"
	"time"

	"github.com/e-timologisi/e-timologisi-utils/resterrors"
)

type User struct {
	ID                     string                   `json:"id"`
	AAErgani               int                      `json:"aa_ergani"`
	Username               string                   `json:"username"`
	Password               string                   `json:"password"`
	SubscriptionBuy        float64                  `json:"subscription_buy"`
	SubscriptionSell       float64                  `json:"subscription_sell"`
	PassPhrase             string                   `json:"pass_phrase"`
	Forename               string                   `json:"forename"`
	Surname                string                   `json:"surname"`
	Email                  string                   `json:"email"`
	AccountantEmail        string                   `json:"accountant_email"`
	BusinessEmail          string                   `json:"business_email"`
	Hostname               string                   `json:"hostname"`
	MobileNumber           string                   `json:"mobile_number"`
	PhoneNumber            string                   `json:"phone_number"`
	Vat                    string                   `json:"vat"`
	DiakritikosTitlos      string                   `json:"diakritikos_titlos"`
	Drastiriotita          string                   `json:"drastiriotita"`
	Doy                    string                   `json:"doy"`
	Eponimia               string                   `json:"eponimia"`
	Address                string                   `json:"address"`
	AddressNumber          string                   `json:"address_number"`
	City                   string                   `json:"city"`
	Area                   string                   `json:"area"`
	Postcode               string                   `json:"postcode"`
	Country                string                   `json:"country"`
	LastLogin              time.Time                `json:"last_login"`
	ValidUntil             time.Time                `json:"valid_until"`
	AllDrastiriotites      []map[string]interface{} `json:"all_drastiriotites"`
	Role                   string                   `json:"role"`
	Token                  string                   `json:"token"`
	ActivationToken        string                   `json:"activation_token"`
	IsActive               bool                     `json:"is_active"`
	ErganiAuth             ErganiAuth               `json:"ergani_auth"`
	BranchStores           []BranchStore            `json:"branch_stores"`
	CanCreateUsers         CanCreateUsers           `json:"can_create_users"`
	DateCreated            time.Time                `json:"date_created"`
	DateUpdated            time.Time                `json:"date_updated"`
	CreatedFrom            CreatedFrom              `json:"created_from"`
	PinNumber              string                   `json:"pin_number"`
	TotalNumberOfEmployees int                      `json:"total_number_of_employees"`
	SubscriptionType       string                   `json:"subscription_type"`
}
type CanCreateUsers struct {
	CanCreate bool     `json:"can_create"`
	Users     []string `json:"users"`
}

func (user *User) Validate() resterrors.RestErr {
	user.Forename = strings.TrimSpace(user.Forename)
	user.Surname = strings.TrimSpace(user.Surname)
	user.Surname = strings.TrimSpace(user.Surname)
	user.Role = strings.TrimSpace(user.Role)
	user.MobileNumber = strings.TrimSpace(user.MobileNumber)

	user.Email = strings.TrimSpace(strings.ToLower(user.Email))
	if user.Email == "" {
		return resterrors.NewBadRequestError("invalid email address")
	}

	user.Password = strings.TrimSpace(user.Password)
	if user.Password == "" {
		return resterrors.NewBadRequestError("invalid password")
	}
	return nil
}

type UserRenewalRequest struct {
	Vat        string    `json:"vat"`
	ValidUntil time.Time `json:"valid_until"`
	UserID     string    `json:"user_id"`
}

type ErganiAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserType string `json:"user_type"`
}
type BranchStore struct {
	ID             string `json:"id"`
	AAErgani       int    `json:"aa_ergani"`
	Eponimia       string `json:"eponimia"`
	Address        string `json:"address"`
	AddressNumber  string `json:"address_number"`
	City           string `json:"city"`
	Postcode       string `json:"postcode"`
	Drastiriotita  string `json:"drastiriotita"`
	Active         bool   `json:"active"`
	ManualCreation bool   `json:"manual_creation"`
}
type CreatedFrom struct {
	Id          string    `json:"id"`
	UserId      string    `json:"user_id"`
	Type        string    `json:"type"`
	DateCreated time.Time `json:"date_created"`
}
type UserResponse struct {
	ID                string    `json:"id"`
	AAErgani          int       `json:"aa_ergani"`
	Password          string    `json:"password"`
	Email             string    `json:"email"`
	MobileNumber      string    `json:"mobile_number"`
	PhoneNumber       string    `json:"phone_number"`
	Vat               string    `json:"vat"`
	DiakritikosTitlos string    `json:"diakritikos_titlos"`
	Drastiriotita     string    `json:"drastiriotita"`
	Eponimia          string    `json:"eponimia"`
	LastLogin         time.Time `json:"last_login"`
	ValidUntil        time.Time `json:"valid_until"`
	IsActive          bool      `json:"is_active"`
	DateCreated       time.Time `json:"date_created"`
	DateUpdated       time.Time `json:"date_updated"`
}

func (user *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:                user.ID,
		Email:             user.Email,
		MobileNumber:      user.MobileNumber,
		PhoneNumber:       user.MobileNumber,
		Vat:               user.Vat,
		DiakritikosTitlos: user.DiakritikosTitlos,
		Drastiriotita:     user.Drastiriotita,
		Eponimia:          user.Eponimia,
		LastLogin:         user.LastLogin,
		ValidUntil:        user.ValidUntil,
		IsActive:          user.IsActive,
		DateCreated:       user.DateCreated,
		DateUpdated:       user.DateUpdated,
	}
}
