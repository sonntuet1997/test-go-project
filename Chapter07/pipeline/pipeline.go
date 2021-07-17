package pipeline

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"
	"sync"
)

func sourceWorker(ctx context.Context, source Source, outCh chan<- Payload, errCh chan<- error) {
	for source.Next(ctx) {
		payload := source.Payload()
		select {
		case outCh <- payload:
		case <-ctx.Done():
			return
		}
		if err := source.Error(); err != nil {
			wrappedErr := xerrors.Errorf("pipeline source: %w", err)
			maybeEmitError(wrappedErr, errCh)
		}
	}
}

func sinkWorker(ctx context.Context, sink Sink, inCh <-chan Payload, errCh chan<- error) {
	for {
		select {
		case payload, ok := <-inCh:
			if !ok {
				return
			}
			if err := sink.Consume(ctx, payload); err != nil {
				wrappedErr := xerrors.Errorf("pipeline sink: %w", err)
				maybeEmitError(wrappedErr, errCh)
				return
			}
			payload.MarkAsProcessed()
		case <-ctx.Done():
			return
		}
	}
}

type Pipeline struct {
	stages []StageRunner
}

func New(stages ...StageRunner) *Pipeline {
	return &Pipeline{
		stages: stages,
	}
}

func (p *Pipeline) Process(ctx context.Context, source Source, sink Sink) error {
	var wg sync.WaitGroup
	pCtx, ctxCancelFn := context.WithCancel(ctx)
	stageCh := make([]chan Payload, len(p.stages)+1)
	errCh := make(chan error, len(p.stages)+2)
	for i := 0; i < len(stageCh); i++ {
		stageCh[i] = make(chan Payload)
	}
	for i := 0; i < len(p.stages); i++ {
		wg.Add(1)
		go func(stageIndex int) {
			p.stages[stageIndex].Run(pCtx, &workerParams{
				stage: stageIndex,
				inCh:  stageCh[stageIndex],
				outCh: stageCh[stageIndex+1],
				errCh: errCh,
			})
			close(stageCh[stageIndex+1])
			wg.Done()
		}(i)
	}
	wg.Add(2)
	go func() {
		sourceWorker(pCtx, source, stageCh[0], errCh)
		close(stageCh[0])
		wg.Done()
	}()
	go func() {
		sinkWorker(pCtx, sink, stageCh[len(stageCh)-1], errCh)
		wg.Done()
	}()
	go func() {
		wg.Wait()
		close(errCh)
		ctxCancelFn()
	}()
	var err error
	for pErr := range errCh {
		err = multierror.Append(err, pErr)
		ctxCancelFn()
	}
	return err
}
