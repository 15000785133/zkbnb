package nftModels

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

type Attribute struct {

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
	Value *int64 `json:"value"`

	// maximum value
	// Example: 40
	// Required: false
	MaxValue *int64 `json:"max_value"`
}

func (a *Attribute) Validate() error {
	var res []error

	if err := a.validateDisplayType(); err != nil {
		res = append(res, err)
	}

	if err := a.validateMaxValue(); err != nil {
		res = append(res, err)
	}

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

func (a *Attribute) validateMaxValue() error {

	if err := validate.Required("max_value", "attributes", a.MaxValue); err != nil {
		return err
	}

	return nil
}

func (a *Attribute) validateName() error {

	if err := validate.Required("trait_type", "attributes", a.Name); err != nil {
		return err
	}

	return nil
}

func (a *Attribute) validateValue() error {

	if err := validate.Required("value", "attributes", a.Value); err != nil {
		return err
	}

	return nil
}

func (a *Attribute) validateDisplayType() error {

	if err := validate.Required("display_type", "attributes", a.DisplayType); err != nil {
		return err
	}

	return nil
}
