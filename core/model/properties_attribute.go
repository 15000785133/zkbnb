package nftModels

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

type PropertiesAttribute struct {

	// displayType
	// Example: properties or levels or stats
	// Required: true
	DisplayType *string `json:"display_type"`

	// name
	// Example: ratio
	// Required: true
	Name *string `json:"trait_type"`

	// value
	// Example: 20
	// Required: true
	Value interface{} `json:"value"`
}

func (a *PropertiesAttribute) Validate() error {
	var res []error

	if err := a.validateName(); err != nil {
		res = append(res, err)
	}

	if err := a.validateValue(); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		err := fmt.Sprintln(res)
		return errors.New(err)
	}
	return nil
}

func (a *PropertiesAttribute) validateName() error {

	if err := validate.Required("trait_type", "attributes", a.Name); err != nil {
		return err
	}

	return nil
}

func (a *PropertiesAttribute) validateValue() error {

	if err := validate.Required("value", "attributes", a.Value); err != nil {
		return err
	}

	return nil
}
