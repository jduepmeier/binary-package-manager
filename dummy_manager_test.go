package bpm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummyManager(t *testing.T) {
	manager := DummyManager{}
	assert.NoError(t, manager.Init(), "init should not return an error")
	assert.Equal(t, 1, manager.GetCounter("Init"), "manager.GetCounter should return 1 for Init")
	manager.ResetCounters()
	assert.Equal(t, 0, manager.GetCounter("Init"), "manager.GetCounter should return 0 after reset")
	manager = DummyManager{}
	assert.Equal(t, 0, manager.GetCounter("Init"), "manager.GetCounter return 0")
}
