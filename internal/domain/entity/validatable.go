package entity

type Validatable interface {
	validate() error
}

type Validated[T Validatable] struct {
	Value   T
	isValid bool
}

func (v *Validated[T]) IsValid() bool {
	return v.isValid
}

func NewValidated[T Validatable](value T) (*Validated[T], error) {
	if err := value.validate(); err != nil {
		return nil, err
	}

	return &Validated[T]{
		Value:   value,
		isValid: true,
	}, nil
}
