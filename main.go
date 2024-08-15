package goSystemd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	goPermissions "github.com/alirezasn3/go-permissions"
)

type Service struct {
	Name       string
	ExecStart  string
	Restart    string
	RestartSec string
}

func CreateService(service *Service) error {
	// check if systemctl
	if _, e := exec.Command("systemctl", "--version").Output(); e != nil {
		return errors.New("systemctl not found")
	}

	// check if service exists
	b, e := exec.Command("systemctl", "status", service.Name).CombinedOutput()
	if e != nil && e.Error() != "exit status 3" {
		// create service file
		if strings.Contains(string(b), "could not be found") {
			// check if user has permission to write to /etc/systemd/system/
			permissions, e := goPermissions.GetPermissions("/etc/systemd/system/")
			if e != nil {
				return e
			}
			if !permissions.Write {
				return errors.New("you don't have write access to /etc/systemd/system/")
			}

			// parse service
			text := fmt.Sprintf("[Unit]\nDescription=%s\nAfter=network-online.target\nWants=network-online.target\n\n[Service]\nType=simple\nPIDFile=/run/%s.pid\nExecStart=%s\nRestart=%s\nRestartSec=%s\n\n[Install]\nWantedBy=multi-user.target", service.Name, service.Name, service.ExecStart, service.Restart, service.RestartSec)

			// write service file
			e = os.WriteFile("/etc/systemd/system/"+service.Name+".service", []byte(text), 0666)
			if e != nil {
				return e
			}
		} else {
			return errors.New(string(b))
		}
	} else {
		// check if service is enabled
		if !strings.Contains(string(b), "disabled;") && !strings.Contains(string(b), "enabled;") {
			return nil
		}
	}

	// enable service
	b, e = exec.Command("systemctl", "enable", service.Name).CombinedOutput()
	if e != nil {
		if strings.Contains(string(b), "Interactive authentication required") {
			return errors.New("permission denied")
		}
		return errors.New(e.Error() + ": " + string(b))
	}

	// start service
	b, e = exec.Command("systemctl", "start", service.Name).CombinedOutput()
	if e != nil {
		if strings.Contains(string(b), "Interactive authentication required") {
			return errors.New("permission denied")
		}
		return errors.New(e.Error() + ": " + string(b))
	}

	return nil
}

func DeleteService(serviceName string) error {
	// check if systemctl
	if _, e := exec.Command("systemctl", "--version").Output(); e != nil {
		return errors.New("systemctl not found")
	}

	// check if service exists
	b, e := exec.Command("systemctl", "status", serviceName).CombinedOutput()
	if e != nil && e.Error() != "exit status 3" {
		if strings.Contains(string(b), "could not be found") {
			return errors.New("service not found")
		}
		return errors.New("failed to check service status: " + string(b))
	}

	// check if user has permission to delete file from /etc/systemd/system/
	permissions, e := goPermissions.GetPermissions("/etc/systemd/system/")
	if e != nil {
		return e
	}
	if !permissions.Write {
		return errors.New("you don't have write access to /etc/systemd/system/")
	}

	// stop service
	b, e = exec.Command("systemctl", "stop", serviceName).CombinedOutput()
	if e != nil {
		if strings.Contains(string(b), "Interactive authentication required") {
			return errors.New("permission denied")
		}
		return errors.New(e.Error() + ": " + string(b))
	}

	// disable service
	b, e = exec.Command("systemctl", "disable", serviceName).CombinedOutput()
	if e != nil {
		if strings.Contains(string(b), "Interactive authentication required") {
			return errors.New("permission denied")
		}
		return errors.New(e.Error() + ": " + string(b))
	}

	// write service file
	e = os.Remove("/etc/systemd/system/" + serviceName + ".service")
	if e != nil {
		return e
	}

	return nil
}
