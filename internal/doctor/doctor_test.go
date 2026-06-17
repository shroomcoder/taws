package doctor

import (
	"testing"
)

func TestNew(t *testing.T) {
	d := New()
	if d == nil {
		t.Fatal("New() returned nil")
	}
}

func TestRunReturnsChecks(t *testing.T) {
	d := New()
	checks := d.Run()

	if len(checks) == 0 {
		t.Fatal("Run() returned no checks")
	}
}

func TestCheckGo(t *testing.T) {
	d := New()
	checks := d.Run()

	var goCheck *Check
	for i, c := range checks {
		if c.Name == "Go" {
			goCheck = &checks[i]
			break
		}
	}

	if goCheck == nil {
		t.Fatal("Go check not found")
	}

	if goCheck.Status != "OK" && goCheck.Status != "MISSING" {
		t.Errorf("unexpected Go status: %s", goCheck.Status)
	}
}

func TestCheckGit(t *testing.T) {
	d := New()
	checks := d.Run()

	var gitCheck *Check
	for i, c := range checks {
		if c.Name == "Git" {
			gitCheck = &checks[i]
			break
		}
	}

	if gitCheck == nil {
		t.Fatal("Git check not found")
	}

	if gitCheck.Status != "OK" && gitCheck.Status != "MISSING" {
		t.Errorf("unexpected Git status: %s", gitCheck.Status)
	}
}

func TestCheckTermux(t *testing.T) {
	d := New()
	checks := d.Run()

	var termuxCheck *Check
	for i, c := range checks {
		if c.Name == "Termux" {
			termuxCheck = &checks[i]
			break
		}
	}

	if termuxCheck == nil {
		t.Fatal("Termux check not found")
	}

	if termuxCheck.Status != "SKIP" && termuxCheck.Status != "OK" && termuxCheck.Status != "WARN" {
		t.Errorf("unexpected Termux status: %s", termuxCheck.Status)
	}
}

func TestFormatResults(t *testing.T) {
	checks := []Check{
		{Name: "Test1", Status: "OK", Message: "All good"},
		{Name: "Test2", Status: "MISSING", Message: "Not found", Critical: true},
		{Name: "Test3", Status: "WARN", Message: "Be careful"},
	}

	output := FormatResults(checks)

	if output == "" {
		t.Fatal("FormatResults returned empty string")
	}

	if len(output) < 50 {
		t.Error("FormatResults output too short")
	}
}

func TestCriticalCheckCounting(t *testing.T) {
	checks := []Check{
		{Name: "Critical1", Status: "MISSING", Critical: true},
		{Name: "Critical2", Status: "FAIL", Critical: true},
		{Name: "Warning1", Status: "WARN", Critical: false},
		{Name: "OK1", Status: "OK", Critical: false},
	}

	output := FormatResults(checks)

	if output == "" {
		t.Fatal("FormatResults returned empty string")
	}
}
