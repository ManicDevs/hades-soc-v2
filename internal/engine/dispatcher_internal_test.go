package engine_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hades-v2/internal/engine"
	"hades-v2/pkg/sdk"
)

func TestDispatcher_Start(t *testing.T) {
	dispatcher := engine.NewDispatcher(&engine.DispatcherConfig{
		MaxWorkers: 5,
		QueueSize:  100,
	})

	err := dispatcher.Start()
	assert.NoError(t, err)

	dispatcher.Stop()
}

func TestDispatcher_SubmitTask(t *testing.T) {
	dispatcher := engine.NewDispatcher(&engine.DispatcherConfig{
		MaxWorkers: 2,
		QueueSize:  10,
	})
	defer dispatcher.Stop()

	err := dispatcher.Start()
	require.NoError(t, err)

	testModule := sdk.NewBaseModule("test-module", "Test module for dispatcher", sdk.CategoryAuxiliary)

	err = dispatcher.RegisterModule(testModule)
	require.NoError(t, err)

	ctx := context.Background()
	task, err := dispatcher.SubmitTask("test-module", ctx)
	require.NoError(t, err)
	assert.NotNil(t, task)
}

func TestDispatcher_ConcurrentTasks(t *testing.T) {
	dispatcher := engine.NewDispatcher(&engine.DispatcherConfig{
		MaxWorkers: 3,
		QueueSize:  50,
	})
	defer dispatcher.Stop()

	err := dispatcher.Start()
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		module := sdk.NewBaseModule(fmt.Sprintf("test-module-%d", i), "Test module", sdk.CategoryAuxiliary)
		err := dispatcher.RegisterModule(module)
		require.NoError(t, err)
	}

	ctx := context.Background()
	var tasks []*engine.Task
	for i := 0; i < 10; i++ {
		task, err := dispatcher.SubmitTask(fmt.Sprintf("test-module-%d", i%5), ctx)
		require.NoError(t, err)
		tasks = append(tasks, task)
	}

	time.Sleep(100 * time.Millisecond)

	for _, task := range tasks {
		assert.NotNil(t, task)
	}
}

func TestDispatcher_RedundantTaskPrevention(t *testing.T) {
	t.Skip("Redundant task detection requires database integration")

	dispatcher := engine.NewDispatcher(&engine.DispatcherConfig{
		MaxWorkers: 2,
		QueueSize:  10,
	})
	defer dispatcher.Stop()

	err := dispatcher.Start()
	require.NoError(t, err)

	testModule := sdk.NewBaseModule("test-module", "Test module", sdk.CategoryAuxiliary)

	err = dispatcher.RegisterModule(testModule)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = dispatcher.SubmitTaskWithTarget("test-module", ctx, "same-target", "test-type")
	require.NoError(t, err)

	_, err = dispatcher.SubmitTaskWithTarget("test-module", ctx, "same-target", "test-type")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redundant task")
}

func TestDispatcher_ModuleRegistration(t *testing.T) {
	dispatcher := engine.NewDispatcher(&engine.DispatcherConfig{
		MaxWorkers: 2,
		QueueSize:  10,
	})
	defer dispatcher.Stop()

	err := dispatcher.Start()
	require.NoError(t, err)

	validModule := sdk.NewBaseModule("valid-module", "Valid test module", sdk.CategoryAuxiliary)

	err = dispatcher.RegisterModule(validModule)
	assert.NoError(t, err)

	err = dispatcher.RegisterModule(validModule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	invalidModule := sdk.NewBaseModule("", "Invalid module", sdk.CategoryAuxiliary)

	err = dispatcher.RegisterModule(invalidModule)
	assert.Error(t, err)
}

func BenchmarkDispatcher_SubmitTask(b *testing.B) {
	dispatcher := engine.NewDispatcher(&engine.DispatcherConfig{
		MaxWorkers: 5,
		QueueSize:  1000,
	})
	defer dispatcher.Stop()

	err := dispatcher.Start()
	if err != nil {
		b.Fatal(err)
	}

	testModule := sdk.NewBaseModule("benchmark-module", "Benchmark module", sdk.CategoryAuxiliary)

	err = dispatcher.RegisterModule(testModule)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dispatcher.SubmitTask("benchmark-module", ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
