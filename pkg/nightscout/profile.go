package nightscout

import (
	"context"
	"errors"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/rest"
)

type ProfileGetter interface {
	Profiles() ProfileInterface
}

type ProfileInterface interface {
	Get(ctx context.Context) (*Profile, error)
}

type Profile struct {
	ID             string           `json:"_id"`
	DefaultProfile string           `json:"defaultProfile"`
	Store          map[string]Store `json:"store"`
	StartDate      time.Time        `json:"startDate"`
	Mills          int              `json:"mills"`
	Units          string           `json:"units"`
	CreatedAt      time.Time        `json:"created_at"`
}
type Carbratio struct {
	Time          string  `json:"time"`
	Value         float64 `json:"value"`
	TimeAsSeconds int     `json:"timeAsSeconds"`
}
type Sens struct {
	Time          string `json:"time"`
	Value         int    `json:"value"`
	TimeAsSeconds int    `json:"timeAsSeconds"`
}
type Basal struct {
	Time          string  `json:"time"`
	Value         float64 `json:"value"`
	TimeAsSeconds int     `json:"timeAsSeconds"`
}
type TargetLow struct {
	Time          string  `json:"time"`
	Value         float64 `json:"value"`
	TimeAsSeconds int     `json:"timeAsSeconds"`
}
type TargetHigh struct {
	Time          string  `json:"time"`
	Value         float64 `json:"value"`
	TimeAsSeconds int     `json:"timeAsSeconds"`
}
type Store struct {
	Dia        int          `json:"dia"`
	Carbratio  []Carbratio  `json:"carbratio"`
	CarbsHr    int          `json:"carbs_hr"`
	Delay      int          `json:"delay"`
	Sens       []Sens       `json:"sens"`
	Timezone   string       `json:"timezone"`
	Basal      []Basal      `json:"basal"`
	TargetLow  []TargetLow  `json:"target_low"`
	TargetHigh []TargetHigh `json:"target_high"`
	StartDate  time.Time    `json:"startDate"`
	Units      string       `json:"units"`
}

type Profiles []*Profile

type profile struct {
	client rest.Interface
}

// newProfile returns a profile
func newProfile(c Client) *profile {
	return &profile{
		client: c.RESTClient(),
	}
}

func (p profile) Get(ctx context.Context) (*Profile, error) {
	profiles := Profiles{}
	err := p.client.Get().Name("profile").Do(ctx).Into(&profiles)
	if err != nil {
		return nil, NewNightscoutError(err, "cant get profile")
	}

	if len(profiles) < 1 {
		return nil, NewNightscoutError(errors.New("empty profiles"), "no data")
	}
	return profiles[0], nil
}
