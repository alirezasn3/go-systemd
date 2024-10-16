package gosystemd

type Service struct {
	Name       string
	ExecStart  string
	Restart    string
	RestartSec string
}

// this function only work on linux
func CreateService(service *Service) error {
	return nil
}

// this function only work on linux
func DeleteService(serviceName string) error {
	return nil
}
