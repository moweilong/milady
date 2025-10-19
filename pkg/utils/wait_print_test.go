package utils

import (
	"testing"
	"time"
)

func TestLoopPrint(t *testing.T) {
	runningTip := "Upgrading milady binary"
	finishTip := "Upgrade milady binary done"
	failedTip := "Upgrade milady binary failed"

	p := NewWaitPrinter(time.Millisecond * 100)
	p.LoopPrint(runningTip)
	time.Sleep(time.Millisecond * 1000)
	p.StopPrint(finishTip)

	p = NewWaitPrinter(0)
	p.LoopPrint(runningTip)
	time.Sleep(time.Millisecond * 1100)
	p.StopPrint(failedTip)
	time.Sleep(time.Millisecond * 100)
}
