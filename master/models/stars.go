package models

type Star struct {
	// gorm.Model // NOTE: Is this necessary?
	Galaxy   int  `json:"galaxy"`
	Solar    int  `json:"solar"`
	Location int  `json:"location"`
	IsMoon   bool `json:"is_moon"`
}

func (p *Star) NewStar(galaxy int, solar int, location int, isMoon bool) *Star {
	return &Star{
		Galaxy:   galaxy,
		Solar:    solar,
		Location: location,
		IsMoon:   isMoon,
	}
}
