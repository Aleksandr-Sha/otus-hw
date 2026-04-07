package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

const CurrentStageIndex = 0

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		return runWithoutStages(in, done)
	}

	return runStages(in, done, stages)
}

func runStages(in In, done In, stage []Stage) Out {
	inProxy := make(Bi)

	go readInputDataAndProxy(in, done, inProxy)

	out := stage[CurrentStageIndex](inProxy)

	if len(stage) > 1 {
		return runStages(out, done, stage[1:])
	}

	return out
}

func runWithoutStages(in In, done In) Out {
	out := make(Bi)
	go readInputDataAndProxy(in, done, out)
	return out
}

func readInputDataAndProxy(in In, done In, inProxy Bi) {
	func() {
		defer func() {
			close(inProxy)
			// Ждём пока закроется канал вывода информации с предыдущего stage (в рамках stage канал out) или первичный
			// канал получения данных если stage первый, обеспечив тем самым гарантированный выход из for range внутри
			// stage или отсутствие блокировок первичного писателя соответственно (если какие-то данные перед закрытием
			// успели встать на запись)
			waitCloseChan(in)
		}()
		for {
			select {
			case <-done:
				return
			default:
				select {
				case <-done:
					return
				case v, ok := <-in:

					if !ok {
						return
					}

					inProxy <- v
				}
			}
		}
	}()
}

func waitCloseChan(in In) {
	for range in {
	}
}
