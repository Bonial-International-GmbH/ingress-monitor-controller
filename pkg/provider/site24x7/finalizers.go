package site24x7

import (
	"errors"

	site24x7api "github.com/Bonial-International-GmbH/site24x7-go/api"
)

// finalizer finalizes the configuration of a Site24x7 website monitor.
type finalizer func(*site24x7api.Monitor) error

func (b *builder) finalizeLocationProfile(monitor *site24x7api.Monitor) error {
	if monitor.LocationProfileID != "" || !b.defaults.AutoLocationProfile {
		return nil
	}

	profiles, err := b.client.LocationProfiles().List()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		return errors.New("no location profiles configured")
	}

	monitor.LocationProfileID = profiles[0].ProfileID

	return nil
}

func (b *builder) finalizeNotificationProfile(monitor *site24x7api.Monitor) error {
	if monitor.NotificationProfileID != "" || !b.defaults.AutoNotificationProfile {
		return nil
	}

	profiles, err := b.client.NotificationProfiles().List()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		return errors.New("no notification profiles configured")
	}

	monitor.NotificationProfileID = profiles[0].ProfileID

	return nil
}

func (b *builder) finalizeThresholdProfile(monitor *site24x7api.Monitor) error {
	if monitor.ThresholdProfileID != "" || !b.defaults.AutoThresholdProfile {
		return nil
	}

	profiles, err := b.client.ThresholdProfiles().List()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		return errors.New("no threshold profiles configured")
	}

	monitor.ThresholdProfileID = profiles[0].ProfileID

	return nil
}

func (b *builder) finalizeMonitorGroup(monitor *site24x7api.Monitor) error {
	if len(monitor.MonitorGroups) > 0 || !b.defaults.AutoMonitorGroup {
		return nil
	}

	groups, err := b.client.MonitorGroups().List()
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return errors.New("no monitor groups configured")
	}

	monitor.MonitorGroups = []string{groups[0].GroupID}

	return nil
}

func (b *builder) finalizeUserGroup(monitor *site24x7api.Monitor) error {
	if len(monitor.UserGroupIDs) > 0 || !b.defaults.AutoUserGroup {
		return nil
	}

	groups, err := b.client.UserGroups().List()
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return errors.New("no user groups configured")
	}

	monitor.UserGroupIDs = []string{groups[0].UserGroupID}

	return nil
}
