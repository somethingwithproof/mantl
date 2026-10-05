package framework

import (
	"testing"
	"time"
)

func TestLoadPackagedFramework(t *testing.T) {
	fw, err := Load("../../compliance/frameworks", "soc2", "2017")
	if err != nil || len(fw.Controls) == 0 {
		t.Fatalf("load: %v", err)
	}
	for _, name := range []string{"../soc2", "/soc2", "soc2/../../etc"} {
		if _, err := Load("../../compliance/frameworks", name, ""); err == nil {
			t.Fatal("unsafe framework accepted")
		}
	}
	if _, err := Load("../../compliance/frameworks", "soc2", "wrong"); err == nil {
		t.Fatal("version mismatch accepted")
	}
}
func TestUTCCollectorSchedules(t *testing.T) {
	last := time.Date(2026, 1, 1, 20, 10, 0, 0, time.FixedZone("offset", 3600))
	next, err := Next("0 * * * *", last)
	if err != nil || next.Location() != time.UTC || next.Hour() != 20 {
		t.Fatalf("next=%v err=%v", next, err)
	}
	if _, err := Next("invalid", last); err == nil {
		t.Fatal("invalid schedule accepted")
	}
}
