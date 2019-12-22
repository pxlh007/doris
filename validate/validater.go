package validate

type validater struct{}

// 定义验证接口
type ivalidate interface {
	validate(string, string) (bool, error)
}

// 定义对应关系
var _ ivalidate = validater{}

// 实现验证接口
func (v *validater) validate(data string, vRule string) (bool, error) {
	return true, nil
}
