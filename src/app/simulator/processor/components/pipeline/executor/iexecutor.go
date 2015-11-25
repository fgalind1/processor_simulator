package executor

import "app/simulator/processor/models/operation"

type IExecutor interface {
	Process(operation *operation.Operation) error
}
