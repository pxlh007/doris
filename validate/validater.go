package validate

type Validater struct{}

// 定义验证接口
type IValidate interface {
	Validate(string, string) (bool, error)
}

// 定义对应关系
var _ IValidate = &Validater{}

// 实现验证接口
func (v *Validater) Validate(data string, vRule string) (bool, error) {
	return true, nil
}
