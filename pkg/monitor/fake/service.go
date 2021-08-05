package fake

import (
	"github.com/stretchr/testify/mock"
	networkingv1 "k8s.io/api/networking/v1"
)

type Service struct {
	mock.Mock
}

func (s *Service) EnsureMonitor(ingress *networkingv1.Ingress) error {
	args := s.Called(ingress)

	return args.Error(0)
}

func (s *Service) DeleteMonitor(ingress *networkingv1.Ingress) error {
	args := s.Called(ingress)

	return args.Error(0)
}

func (s *Service) GetProviderIPSourceRanges(ingress *networkingv1.Ingress) ([]string, error) {
	args := s.Called(ingress)

	var ips []string
	if arg, ok := args.Get(0).([]string); ok {
		ips = arg
	}

	return ips, args.Error(1)
}

func (s *Service) AnnotateIngress(ingress *networkingv1.Ingress) (updated bool, err error) {
	args := s.Called(ingress)

	return args.Bool(0), args.Error(1)
}
