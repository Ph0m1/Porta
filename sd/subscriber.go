package sd

type Subscriber interface {
	Hosts() ([]string, error)
}

type FixedSubscriber []string

// Hosts 实现订阅者接口
func (s FixedSubscriber) Hosts() ([]string, error) { return s, nil }
