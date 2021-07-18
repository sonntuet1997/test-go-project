package pipeline

import (
	"context"
	"golang.org/x/xerrors"
	"sync"
)

type fifo struct {
	proc Processor
}

func maybeEmitError(err error, errCh chan<- error) {
	select {
	case errCh <- err:
	default:
	}
}

func (f fifo) Run(ctx context.Context, s StageParams) {
	for {
		select {
		case <-ctx.Done():
			return
		case payloadIn, ok := <-s.Input():
			if !ok {
				return
			}
			payloadOut, err := f.proc.Process(ctx, payloadIn)
			if err != nil {
				wrapperErr := xerrors.Errorf("pipeline stage %d: %w", s.StageIndex(), err)
				maybeEmitError(wrapperErr, s.Error())
				return
			}
			if payloadOut == nil {
				payloadIn.MarkAsProcessed()
				continue
			}
			select {
			case s.Output() <- payloadOut:
			case <-ctx.Done():
				return
			}
		}
	}
}

func FIFO(proc Processor) StageRunner {
	return fifo{proc: proc}
}

type fixedWorkerPool struct {
	fifos []StageRunner
}

func (f fixedWorkerPool) Run(ctx context.Context, s StageParams) {
	var wg sync.WaitGroup
	for i := 0; i < len(f.fifos); i++ {
		wg.Add(1)
		go func(fifoIndex int) {
			f.fifos[fifoIndex].Run(ctx, s)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func FixedWorkerPool(proc Processor, numWorkers int) StageRunner {
	if numWorkers <= 0 {
		panic("Must > 0")
	}
	fifos := make([]StageRunner, numWorkers)
	for i := 0; i < numWorkers; i++ {
		fifos[i] = FIFO(proc)
	}
	return &fixedWorkerPool{fifos: fifos}
}

type dynamicWorkerPool struct {
	proc      Processor
	tokenPool chan struct{}
}

func (d dynamicWorkerPool) Run(ctx context.Context, s StageParams) {
stop:
	for {
		select {
		case <-ctx.Done():
			break stop
		case payloadIn, ok := <-s.Input():
			if !ok {
				break stop
			}
			var token struct{}
			select {
			case token = <-d.tokenPool:
				go func(payloadIn Payload, token struct{}) {
					defer func() { d.tokenPool <- token }()
					payloadOut, err := d.proc.Process(ctx, payloadIn)
					if err != nil {
						wrappedErr := xerrors.Errorf("pp stage %d: %w", s.StageIndex(), err)
						maybeEmitError(wrappedErr, s.Error())
						return
					}
					if payloadOut == nil {
						payloadIn.MarkAsProcessed()
						return
					}
					select {
					case s.Output() <- payloadOut:
					case <-ctx.Done():
					}
				}(payloadIn, token)
			case <-ctx.Done():
				break stop
			}
		}
	}
}

func DynamicWorkerPool(proc Processor, maxWorkers int) StageRunner {
	if maxWorkers <= 0 {
		panic("must > 0")
	}
	tokenPool := make(chan struct{}, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		tokenPool <- struct{}{}
	}
	return &dynamicWorkerPool{proc: proc, tokenPool: tokenPool}
}

type broadcast struct {
	fifos []StageRunner
}

func (b broadcast) Run(ctx context.Context, s StageParams) {
	var wg sync.WaitGroup
	inCh := make([]chan Payload, len(b.fifos))
	for i := 0; i < len(b.fifos); i++ {
		wg.Add(1)
		inCh[i] = make(chan Payload)
		go func(fifoIndex int) {
			fifoParams := &workerParams{
				stage: s.StageIndex(),
				inCh:  inCh[fifoIndex],
				outCh: s.Output(),
				errCh: s.Error(),
			}
			b.fifos[fifoIndex].Run(ctx, fifoParams)
			wg.Done()
		}(i)
	}
done:
	for {
		select {
		case <-ctx.Done():
			break done
		case payload, ok := <-s.Input():
			if !ok {
				break
			}
			for i := len(b.fifos) - 1; i >= 0; i-- {
				fifoPayload := payload
				if i != 0 {
					fifoPayload = payload.Clone()
				}
				select {
				case <-ctx.Done():
					break done
				case inCh[i] <- fifoPayload:
				}
			}
		}
	}
	for _, ch := range inCh {
		close(ch)
	}
	wg.Wait()
}

func Broadcast(procs ...Processor) StageRunner {
	if len(procs) == 0 {
		panic("proces > 0")
	}
	fifos := make([]StageRunner, len(procs))
	for i, p := range procs {
		fifos[i] = FIFO(p)
	}
	return &broadcast{fifos: fifos}
}
