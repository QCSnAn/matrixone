// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indexbuild

import (
	"github.com/matrixorigin/matrixone/pkg/common/reuse"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

var _ vm.Operator = new(Argument)

const (
	ReceiveBatch = iota
	HandleRuntimeFilter
	End
)

type container struct {
	colexec.ReceiverOperator
	state   int
	isMerge bool
	batch   *batch.Batch
}

type Argument struct {
	ctr                  *container
	RuntimeFilterSenders []*colexec.RuntimeFilterChan
	vm.OperatorBase
}

func (arg *Argument) GetOperatorBase() *vm.OperatorBase {
	return &arg.OperatorBase
}

func init() {
	reuse.CreatePool[Argument](
		func() *Argument {
			return &Argument{}
		},
		func(a *Argument) {
			*a = Argument{}
		},
		reuse.DefaultOptions[Argument]().
			WithEnableChecker(),
	)
}

func (arg Argument) TypeName() string {
	return argName
}

func NewArgument() *Argument {
	return reuse.Alloc[Argument](nil)
}

func (arg *Argument) Release() {
	if arg != nil {
		reuse.Free[Argument](arg, nil)
	}
}

func (arg *Argument) SetRuntimeFilterSenders(rfs []*colexec.RuntimeFilterChan) {
	arg.RuntimeFilterSenders = rfs
}

func (arg *Argument) Free(proc *process.Process, pipelineFailed bool, err error) {
	ctr := arg.ctr
	if ctr != nil {
		if ctr.batch != nil {
			proc.PutBatch(ctr.batch)
		}
		ctr.FreeMergeTypeOperator(pipelineFailed)
		if ctr.isMerge {
			ctr.FreeMergeTypeOperator(pipelineFailed)
		} else {
			ctr.FreeAllReg()
		}
	}
}