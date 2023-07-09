package client

type Service struct{}

func MustNewClientService() Service {
	service := Service{}

	return service
}

func (r Service) Run() {
}
