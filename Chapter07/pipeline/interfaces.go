package pipeline

import "context"

type Payload interface {
	Clone() Payload
	MarkAsProcessed()
}

type Processor interface {
	Process(context.Context, Payload) (Payload, error)
}

type ProcessorFunc func(ctx context.Context, payload Payload) (Payload, error)

type PPP int

func (f ProcessorFunc) Process(ctx context.Context, payload Payload) (Payload, error) {
	return f(ctx, payload)
}

func retryingProcessor(proc Processor, isTransient func(err error) bool, maxRetries int) Processor {
	return ProcessorFunc(func(ctx context.Context, payload Payload) (Payload, error) {
		var out Payload
		var err error
		for i := 0; i < maxRetries; i++ {
			if out, err = proc.Process(ctx, payload); err != nil && !isTransient(err) {
				continue
			}
			break
		}
		return out, err
	})
}

type Source interface {
	Next(context.Context) bool
	Payload() Payload
	Error() error
}

type Sink interface {
	Consume(context.Context, Payload) error
}

type StageRunner interface {
	Run(ctx context.Context, s StageParams)
}

type StageParams interface {
	StageIndex() int
	Input() <-chan Payload
	Output() chan<- Payload
	Error() chan<- error
}

type workerParams struct {
	stage int

	// Channels for the worker's input, output and errors.
	inCh  <-chan Payload
	outCh chan<- Payload
	errCh chan<- error
}

func (p *workerParams) StageIndex() int        { return p.stage }
func (p *workerParams) Input() <-chan Payload  { return p.inCh }
func (p *workerParams) Output() chan<- Payload { return p.outCh }
func (p *workerParams) Error() chan<- error    { return p.errCh }
