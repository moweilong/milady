package stat

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestInitBase(t *testing.T) {
	l, _ := zap.NewDevelopment()
	Init(
		// test empty
		WithLog(nil),
		WithPrintInterval(0),

		WithLog(l),
		WithPrintInterval(time.Second),
		WithPrintField(zap.String("host", "127.0.0.1")),

		WithAlarm(WithCPUThreshold(0.85), WithMemoryThreshold(0.85)),
	)

	time.Sleep(time.Second * 2)
}

func TestInit(t *testing.T) {
	l, _ := zap.NewDevelopment()
	Init(
		WithLog(l),
		WithPrintInterval(time.Second),
		WithPrintField(zap.String("host", "127.0.0.1")),

		WithAlarm(WithCPUThreshold(0.85), WithMemoryThreshold(0.85)),
		WithCustomHandler(func(ctx context.Context, sd *StatData) error {
			t.Logf("stat data: %+v\n", sd)
			return nil
		}),
	)

	time.Sleep(time.Second * 3)
}
