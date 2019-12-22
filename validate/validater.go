package validate

type validater struct{}

// 定义验证接口
type ivalidate interface {
	Validate(string, string) (bool, error)
}

// 定义对应关系
var _ ivalidate = validater{}

// 实现验证接口
func (v *validater) Validate(data string, vRule string) (bool, error) {
	return true, nil
}
