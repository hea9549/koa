/*
 * Copyright 2018 De-labtory
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package vm

import (
	"encoding/binary"
	"errors"

	"github.com/DE-labtory/koa/opcode"
)

var ErrInvalidData = errors.New("Invalid data")
var ErrInvalidOpcode = errors.New("invalid opcode")

// The Execute function assemble the rawByteCode into an assembly code,
// which in turn executes the assembly logic.
func Execute(rawByteCode []byte, memory *Memory, callFunc *CallFunc) (*stack, error) {

	s := newStack()
	asm, err := disassemble(rawByteCode)
	if err != nil {
		return &stack{}, err
	}

	for h := asm.code[0]; h != nil; h = asm.next() {
		op, ok := h.(opCode)
		if !ok {
			return &stack{}, ErrInvalidOpcode
		}

		err := op.Do(s, asm, memory, callFunc)
		if err != nil {
			return s, err
		}
	}

	return s, nil
}

type CallFunc struct {
	Func []byte
	Args []byte
}

type opCode interface {
	Do(*stack, asmReader, *Memory, *CallFunc) error
	hexer
}

// Perform opcodes logic.
type add struct{}
type mul struct{}
type sub struct{}
type div struct{}
type mod struct{}
type lt struct{}
type gt struct{}
type eq struct{}
type not struct{}
type pop struct{}
type push struct{}
type mload struct{}
type mstore struct{}

func (add) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	y := stack.pop()
	x := stack.pop()

	stack.push(x + y)

	return nil
}

func (add) hex() []uint8 {
	return []uint8{uint8(opcode.Add)}
}

func (mul) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	y := stack.pop()
	x := stack.pop()

	stack.push(x * y)

	return nil
}

func (mul) hex() []uint8 {
	return []uint8{uint8(opcode.Mul)}
}

func (sub) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	y := stack.pop()
	x := stack.pop()

	stack.push(x - y)

	return nil
}

func (sub) hex() []uint8 {
	return []uint8{uint8(opcode.Sub)}
}

// Be careful! int.Div and int.Quo is different
func (div) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	y := stack.pop()
	x := stack.pop()

	item, _ := euclidean_div(x, y)

	stack.push(item)

	return nil
}

func (div) hex() []uint8 {
	return []uint8{uint8(opcode.Div)}
}

func (mod) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	y := stack.pop()
	x := stack.pop()

	_, item := euclidean_div(x, y)

	stack.push(item)

	return nil
}

func (mod) hex() []uint8 {
	return []uint8{uint8(opcode.Mod)}
}

func (lt) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	y, x := stack.pop(), stack.pop()

	if x < y { // x < y
		stack.push(item(1))
	} else {
		stack.push(item(0))
	}

	return nil
}

func (lt) hex() []uint8 {
	return []uint8{uint8(opcode.LT)}
}

func (gt) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	y, x := stack.pop(), stack.pop()

	if x > y { // x > y
		stack.push(item(1))
	} else {
		stack.push(item(0))
	}

	return nil
}

func (gt) hex() []uint8 {
	return []uint8{uint8(opcode.GT)}
}

func (eq) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	y, x := stack.pop(), stack.pop()

	if x == y { // x == y
		stack.push(item(1))
	} else {
		stack.push(item(0))
	}

	return nil
}

func (eq) hex() []uint8 {
	return []uint8{uint8(opcode.EQ)}
}

func (not) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	x := stack.pop()

	stack.push(^x)
	return nil
}

func (not) hex() []uint8 {
	return []uint8{uint8(opcode.NOT)}
}

func (pop) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	_ = stack.pop()
	return nil
}

func (pop) hex() []uint8 {
	return []uint8{uint8(opcode.Pop)}
}

func (push) Do(stack *stack, asm asmReader, _ *Memory, contract *CallFunc) error {
	code := asm.next()
	data, ok := code.(Data)
	if !ok {
		return ErrInvalidData
	}
	item := item(bytesToInt32(data.hex()))
	stack.push(item)

	return nil
}

func (push) hex() []uint8 {
	return []uint8{uint8(opcode.Push)}
}

// TODO: implement me w/ test cases :-)
func (mload) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	return nil
}

func (mload) hex() []uint8 {
	return []uint8{uint8(opcode.Mload)}
}

// TODO: implement me w/ test cases :-)
func (mstore) Do(stack *stack, _ asmReader, _ *Memory, _ *CallFunc) error {
	return nil
}

func (mstore) hex() []uint8 {
	return []uint8{uint8(opcode.Mstore)}
}

func int32ToBytes(int32 int32) []byte {
	byteSlice := make([]byte, 4)
	binary.BigEndian.PutUint32(byteSlice, uint32(int32))
	return byteSlice
}

func bytesToInt32(bytes []byte) int32 {
	int32 := int32(binary.BigEndian.Uint32(bytes))
	return int32
}

func euclidean_div(a item, b item) (item, item) {
	var q int32
	var r int32
	A := int32(a)
	B := int32(b)

	if A < 0 && B > 0 {
		q = int32(A/B) - 1
		r = A - (B * q)
	} else if A > 0 && B < 0 {
		q = int32(A / B)
		r = A - (B * q)
	} else if A > 0 && B > 0 {
		q = int32(A / B)
		r = A - (B * q)
	} else if A < 0 && B < 0 {
		q = int32((A + B) / B)
		r = A - (B * q)
	}

	return item(q), item(r)
}
