package site24x7

import (
	"errors"

	site24x7 "github.com/Bonial-International-GmbH/site24x7-go"
	site24x7api "github.com/Bonial-International-GmbH/site24x7-go/api"
)

type defaulter func(site24x7.Client, *site24x7api.Monitor) error

func withDefaultLocationProfile(c site24x7.Client, monitor *site24x7api.Monitor) error {
	if monitor.LocationProfileID != "" {
		return nil
	}

	profiles, err := c.LocationProfiles().List()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		return errors.New("no location profiles configured")
	}

	monitor.LocationProfileID = profiles[0].ProfileID

	return nil
}

func withDefaultNotificationProfile(c site24x7.Client, monitor *site24x7api.Monitor) error {
	if monitor.NotificationProfileID != "" {
		return nil
	}

	profiles, err := c.NotificationProfiles().List()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		return errors.New("no notification profiles configured")
	}

	monitor.NotificationProfileID = profiles[0].ProfileID

	return nil
}

func withDefaultThresholdProfile(c site24x7.Client, monitor *site24x7api.Monitor) error {
	if monitor.ThresholdProfileID != "" {
		return nil
	}

	profiles, err := c.ThresholdProfiles().List()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		return errors.New("no threshold profiles configured")
	}

	monitor.ThresholdProfileID = profiles[0].ProfileID

	return nil
}

func withDefaultMonitorGroup(c site24x7.Client, monitor *site24x7api.Monitor) error {
	if len(monitor.MonitorGroups) > 0 {
		return nil
	}

	groups, err := c.MonitorGroups().List()
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return errors.New("no monitor groups configured")
	}

	monitor.MonitorGroups = []string{groups[0].GroupID}

	return nil
}

func withDefaultUserGroup(c site24x7.Client, monitor *site24x7api.Monitor) error {
	if len(monitor.UserGroupIDs) > 0 {
		return nil
	}

	groups, err := c.UserGroups().List()
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return errors.New("no user groups configured")
	}

	monitor.UserGroupIDs = []string{groups[0].UserGroupID}

	return nil
}
