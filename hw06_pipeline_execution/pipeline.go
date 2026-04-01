package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	return stageRunWithSubSlice(stages, in, done)
}

func stageRunWithSubSlice(stage []Stage, in In, done In) Out {
	inWrap := make(Bi)

	go func() {
		defer func() {
			close(inWrap)
			<-in // Читаем значение на случай, если у нас горутина исполняющая stage успела зайти в range
			// с записью в канал out, но читателей уже не осталось
		}()
		for {
			select {
			// А что тут будет в тот момент, когда будет и значение в in и done закроется?
			case <-done:
				return
			case v, ok := <-in:

				if !ok {
					return
				}

				inWrap <- v
			}
		}
	}()

	if len(stage) > 1 {
		return stageRunWithSubSlice(stage[1:], stage[0](inWrap), done)
	}

	return stage[0](inWrap)
}

// Пробный вариант с подслайсом
