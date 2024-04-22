package site24x7

import (
	"fmt"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/models"
	site24x7 "github.com/Bonial-International-GmbH/site24x7-go"
	site24x7api "github.com/Bonial-International-GmbH/site24x7-go/api"
	site24x7apierrors "github.com/Bonial-International-GmbH/site24x7-go/api/errors"
)

type builder struct {
	client     site24x7.Client
	defaults   config.Site24x7MonitorDefaults
	finalizers []finalizer
}

func newBuilder(client site24x7.Client, defaults config.Site24x7MonitorDefaults) *builder {
	b := &builder{
		client:   client,
		defaults: defaults,
	}

	b.finalizers = []finalizer{
		b.finalizeLocationProfile,
		b.finalizeNotificationProfile,
		b.finalizeThresholdProfile,
		b.finalizeMonitorGroup,
		b.finalizeUserGroup,
	}

	return b
}

func (b *builder) FromModel(model *models.Monitor) (*site24x7api.Monitor, error) {
	anno := model.Annotations
	defaults := b.defaults

	monitor := &site24x7api.Monitor{
		Type:        "URL",
		MonitorID:   model.ID,
		DisplayName: model.Name,
		Website:     model.URL,
	}

	monitor.CheckFrequency = anno.StringValue(config.AnnotationSite24x7CheckFrequency, defaults.CheckFrequency)
	monitor.HTTPMethod = anno.StringValue(config.AnnotationSite24x7HTTPMethod, defaults.HTTPMethod)
	monitor.AuthUser = anno.StringValue(config.AnnotationSite24x7AuthUser, defaults.AuthUser)
	monitor.AuthPass = anno.StringValue(config.AnnotationSite24x7AuthPass, defaults.AuthPass)
	monitor.MatchCase = anno.BoolValue(config.AnnotationSite24x7MatchCase, defaults.MatchCase)
	monitor.UserAgent = anno.StringValue(config.AnnotationSite24x7UserAgent, defaults.UserAgent)
	monitor.Timeout = anno.IntValue(config.AnnotationSite24x7Timeout, defaults.Timeout)
	monitor.UseNameServer = anno.BoolValue(config.AnnotationSite24x7UseNameServer, defaults.UseNameServer)
	monitor.UserGroupIDs = anno.StringSliceValue(config.AnnotationSite24x7UserGroupIDs, defaults.UserGroupIDs)
	monitor.MonitorGroups = anno.StringSliceValue(config.AnnotationSite24x7MonitorGroupIDs, defaults.MonitorGroupIDs)
	monitor.LocationProfileID = anno.StringValue(config.AnnotationSite24x7LocationProfileID, defaults.LocationProfileID)
	monitor.NotificationProfileID = anno.StringValue(config.AnnotationSite24x7NotificationProfileID, defaults.NotificationProfileID)
	monitor.ThresholdProfileID = anno.StringValue(config.AnnotationSite24x7ThresholdProfileID, defaults.ThresholdProfileID)

	err := anno.ParseJSON(config.AnnotationSite24x7CustomHeaders, &monitor.CustomHeaders)
	if err != nil {
		return nil, err
	}

	if monitor.CustomHeaders == nil {
		monitor.CustomHeaders = defaults.CustomHeaders
	}

	if err := b.attachActions(anno, monitor); err != nil {
		return nil, err
	}

	return b.finalizeMonitor(monitor)
}

func (b *builder) finalizeMonitor(monitor *site24x7api.Monitor) (*site24x7api.Monitor, error) {
	for _, f := range b.finalizers {
		if err := f(monitor); err != nil {
			return nil, err
		}
	}

	return monitor, nil
}

func (b *builder) findITAutomation(actionNameOrID string) (*site24x7api.ITAutomation, error) {
	automation, err := b.client.ITAutomations().Get(actionNameOrID)
	if site24x7apierrors.IsNotFound(err) {
		// Try to look up the IT automation by name.
		automations, err := b.client.ITAutomations().List()
		if err != nil {
			return nil, err
		}

		for _, automation := range automations {
			if automation.ActionName == actionNameOrID {
				return automation, nil
			}
		}

		return nil, fmt.Errorf("IT automation %q not found", actionNameOrID)
	} else if err != nil {
		return nil, err
	}

	return automation, nil
}

func (b *builder) attachActions(anno config.Annotations, monitor *site24x7api.Monitor) error {
	err := anno.ParseJSON(config.AnnotationSite24x7Actions, &monitor.ActionIDs)
	if err != nil {
		return err
	}

	err = b.attachAction(anno.StringValue(config.AnnotationSite24x7ActionDown), site24x7api.Down, monitor)
	if err != nil {
		return err
	}

	err = b.attachAction(anno.StringValue(config.AnnotationSite24x7ActionUp), site24x7api.Up, monitor)
	if err != nil {
		return err
	}

	if monitor.ActionIDs == nil {
		monitor.ActionIDs = b.defaults.Actions
	}

	return nil
}

func (b *builder) attachAction(actionNameOrID string, status site24x7api.Status, monitor *site24x7api.Monitor) error {
	if actionNameOrID == "" {
		return nil
	}

	automation, err := b.findITAutomation(actionNameOrID)
	if err != nil {
		return err
	}

	actionRef := site24x7api.ActionRef{
		ActionID:  automation.ActionID,
		AlertType: status,
	}

	monitor.ActionIDs = append(monitor.ActionIDs, actionRef)

	return nil
}
