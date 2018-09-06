package util

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

var (
	howManyOperations int = 100
	operations        []OperationRec
)

func fillRequest(N int) (req http.Request) {
	req.Method = "GET"
	req.Proto = fmt.Sprintf("req %d", N)
	req.Header = http.Header{"one": []string{"two"}}
	return
}

func fillResponse(N int) (resp http.Response) {
	resp.Status = "OK 200"
	resp.Proto = fmt.Sprintf("resp %d", N)
	resp.Header = http.Header{fmt.Sprintf("key %d", N): []string{fmt.Sprintf("value %d", N)}}
	return
}

func fillOperations() {
	for N := 0; N < howManyOperations; N++ {
		var op OperationRec
		op.Caller = fmt.Sprintf("caller %d", N)
		op.Data = "zzzzzzzzzzzzzzzzzzzzzzzzzzz"
		if N%2 == 0 {
			op.OpType = "request"
			op.Operation = "GET"
			op.Url = "xxxx/yyyy"
			req := fillRequest(N)
			op.Req = &req
		} else {
			op.OpType = "response"
			resp := fillResponse(N)
			op.Resp = &resp
		}
		operations = append(operations, op)
	}
}

func okEqualInt(label string, a, b int, t *testing.T) {
	if a == b {
		t.Logf("ok - %s %d == %d", label, a, b)
	} else {
		t.Logf("not ok - %s %d != %d", label, a, b)
		t.Fail()
	}
}

func okEqualString(label, a, b string, t *testing.T) {
	if a == b {
		t.Logf("ok - %s %s == %s", label, a, b)
	} else {
		t.Logf("not ok - %s %s != %s", label, a, b)
		t.Fail()
	}
}

func TestInt(t *testing.T) {
	var maxElements int = 2
	var maxInput int = 100
	fss := NewFixedSizeStack(maxElements)
	for N := 0; N < maxInput; N++ {
		fss.Push(N)
	}
	for J := 1; J < maxElements; J++ {
		foundInt := fss.Pop().(int)
		okEqualInt(fmt.Sprintf("last element %d", J), maxInput-J, foundInt, t)
	}
}

func TestOps(t *testing.T) {
	var maxElements int = 4
	fss := NewFixedSizeStack(maxElements)
	for _, op := range operations {
		fss.Push(op)
	}
	okEqualInt("FixedSizeStack size", maxElements, fss.Len(), t)
	for N := 0; N < maxElements; N++ {
		op1 := fss.Pop().(OperationRec)
		if op1.OpType == "request" {
			okEqualString("operation request", fmt.Sprintf("req %d", howManyOperations-1-N), op1.Req.Proto, t)
		} else {
			okEqualString("operation response", fmt.Sprintf("resp %d", howManyOperations-1-N), op1.Resp.Proto, t)
		}
	}
}

func TestMain(m *testing.M) {
	// Build up function
	fillOperations()
	// Runs all the tests
	exitCode := m.Run()
	// Add teardown function if needed
	os.Exit(exitCode)
}
