package site24x7

import (
	"errors"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/models"
	site24x7 "github.com/Bonial-International-GmbH/site24x7-go"
	site24x7api "github.com/Bonial-International-GmbH/site24x7-go/api"
	"k8s.io/klog"
)

type builder struct {
	client     site24x7.Client
	defaults   config.Site24x7MonitorDefaults
	monitor    *site24x7api.Monitor
	defaulters []defaulter
	err        error
}

func newBuilder(client site24x7.Client, defaults config.Site24x7MonitorDefaults) *builder {
	return &builder{
		client:   client,
		defaults: defaults,
	}
}

func (b *builder) withDefaulters(defaulters ...defaulter) *builder {
	if b.err != nil {
		return b
	}

	b.defaulters = defaulters

	return b
}

func (b *builder) withModel(model *models.Monitor) *builder {
	if b.err != nil {
		return b
	}

	a := model.Annotations
	d := b.defaults

	m := &site24x7api.Monitor{
		Type:                  "URL",
		MonitorID:             model.ID,
		DisplayName:           model.Name,
		Website:               model.URL,
		CheckFrequency:        a.String(config.AnnotationSite24x7CheckFrequency, d.CheckFrequency),
		HTTPMethod:            a.String(config.AnnotationSite24x7HTTPMethod, d.HTTPMethod),
		AuthUser:              a.String(config.AnnotationSite24x7AuthUser, d.AuthUser),
		AuthPass:              a.String(config.AnnotationSite24x7AuthPass, d.AuthPass),
		MatchCase:             a.Bool(config.AnnotationSite24x7MatchCase, d.MatchCase),
		UserAgent:             a.String(config.AnnotationSite24x7UserAgent, d.UserAgent),
		Timeout:               a.Int(config.AnnotationSite24x7Timeout, d.Timeout),
		UseNameServer:         a.Bool(config.AnnotationSite24x7UseNameServer, d.UseNameServer),
		UserGroupIDs:          a.StringSlice(config.AnnotationSite24x7UserGroupIDs, d.UserGroupIDs),
		MonitorGroups:         a.StringSlice(config.AnnotationSite24x7MonitorGroupIDs, d.MonitorGroupIDs),
		LocationProfileID:     a.String(config.AnnotationSite24x7LocationProfileID, d.LocationProfileID),
		NotificationProfileID: a.String(config.AnnotationSite24x7NotificationProfileID, d.NotificationProfileID),
		ThresholdProfileID:    a.String(config.AnnotationSite24x7ThresholdProfileID, d.ThresholdProfileID),
	}

	b.err = a.JSON(config.AnnotationSite24x7CustomHeaders, &m.CustomHeaders)
	if b.err != nil {
		return b
	} else if m.CustomHeaders == nil {
		m.CustomHeaders = d.CustomHeaders
	}

	b.err = a.JSON(config.AnnotationSite24x7Actions, &m.ActionIDs)
	if b.err != nil {
		return b
	} else if m.ActionIDs == nil {
		m.ActionIDs = d.Actions
	}

	b.monitor = m

	return b
}

func (b *builder) build() (*site24x7api.Monitor, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.monitor == nil {
		return nil, errors.New("cannot build monitor without model")
	}

	for _, defaulter := range b.defaulters {
		err := defaulter(b.client, b.monitor)
		if err != nil {
			return nil, err
		}
	}

	klog.V(2).Infof("built site24x7 monitor: %#v", b.monitor)

	return b.monitor, nil
}
