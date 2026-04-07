package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	g := createStageGeneratorWithoutWg()

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go writeInputData(data, in)

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		abortDur := sleepPerStage * 2
		go closeDoneChanAfterTimeout(abortDur, done)
		go writeInputData(data, in)

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStopWhenDone(t *testing.T) {
	wg := sync.WaitGroup{}
	g := createStageGeneratorWithWg(&wg)

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go closeDoneChanAfterTimeout(sleepPerStage*2, done)
		go writeInputData(data, in)

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func TestAllStageStopWhenProcessAllData(t *testing.T) {
	wg := sync.WaitGroup{}
	g := createStageGeneratorWithWg(&wg)

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	in := make(Bi)
	data := []int{1, 2, 3, 4, 5}

	go writeInputData(data, in)

	result := make([]string, 0, 10)
	for s := range ExecutePipeline(in, nil, stages...) {
		result = append(result, s.(string))
	}

	wg.Wait()

	require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
}

func TestOneStageWhenProcessAllData(t *testing.T) {
	wg := sync.WaitGroup{}
	g := createStageGeneratorWithWg(&wg)

	stages := []Stage{g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 })}

	in := make(Bi)
	data := []int{1, 2, 3, 4, 5}

	go writeInputData(data, in)

	result := make([]int, 0, 10)
	for s := range ExecutePipeline(in, nil, stages...) {
		result = append(result, s.(int))
	}

	wg.Wait()

	require.Equal(t, []int{2, 4, 6, 8, 10}, result)
}

func TestZeroStageWhenProcessAllData(t *testing.T) {
	in := make(Bi)
	data := []int{1, 2, 3, 4, 5}

	go writeInputData(data, in)

	result := make([]int, 0, 10)
	for s := range ExecutePipeline(in, nil, []Stage{}...) {
		result = append(result, s.(int))
	}

	require.Equal(t, []int{1, 2, 3, 4, 5}, result)
}

func TestZeroStageWhenDone(t *testing.T) {
	in := make(Bi)
	done := make(Bi)
	data := []int{1, 2, 3, 4, 5}

	go close(done)
	go writeInputData(data, in)

	result := make([]int, 0, 10)
	for s := range ExecutePipeline(in, done, []Stage{}...) {
		result = append(result, s.(int))
	}

	require.Len(t, result, 0)
}

func createStageGeneratorWithoutWg() func(_ string, f func(v interface{}) interface{}) Stage {
	return func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}
}

func createStageGeneratorWithWg(wg *sync.WaitGroup) func(_ string, f func(v interface{}) interface{}) Stage {
	return func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}
}

func writeInputData(data []int, in Bi) {
	func() {
		for _, v := range data {
			in <- v
		}
		close(in)
	}()
}

func closeDoneChanAfterTimeout(abortDur time.Duration, done Bi) {
	func() {
		<-time.After(abortDur)
		close(done)
	}()
}
