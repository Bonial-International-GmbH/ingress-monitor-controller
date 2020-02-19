package fake

import (
	"github.com/stretchr/testify/mock"
	"k8s.io/api/extensions/v1beta1"
)

type Service struct {
	mock.Mock
}

func (s *Service) EnsureMonitor(ingress *v1beta1.Ingress) error {
	args := s.Called(ingress)

	return args.Error(0)
}

func (s *Service) DeleteMonitor(ingress *v1beta1.Ingress) error {
	args := s.Called(ingress)

	return args.Error(0)
}

func (s *Service) GetProviderIPSourceRanges(ingress *v1beta1.Ingress) ([]string, error) {
	args := s.Called(ingress)

	var ips []string
	if arg, ok := args.Get(0).([]string); ok {
		ips = arg
	}

	return ips, args.Error(1)
}

func (s *Service) AnnotateIngress(ingress *v1beta1.Ingress) (updated bool, err error) {
	args := s.Called(ingress)

	return args.Bool(0), args.Error(1)
}
