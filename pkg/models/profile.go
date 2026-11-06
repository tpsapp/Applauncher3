package models

import "errors"

var ErrEmptyProfileName = errors.New("profile name cannot be empty")

type Profile struct {
	Name  string       `json:"name"`
	Items []LaunchItem `json:"items"`
}

func (p *Profile) Validate() error {
	if p.Name == "" {
		return ErrEmptyProfileName
	}
	if p.Items == nil {
		p.Items = []LaunchItem{}
	}
	for i := range p.Items {
		if err := p.Items[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}
