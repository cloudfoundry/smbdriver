// Code generated by counterfeiter. DO NOT EDIT.
package smbdriverfakes

import (
	"sync"

	"code.cloudfoundry.org/voldriver"
	"github.com/AbelHu/smbdriver/driveradmin"
)

type FakeDriverAdmin struct {
	EvacuateStub        func(env voldriver.Env) driveradmin.ErrorResponse
	evacuateMutex       sync.RWMutex
	evacuateArgsForCall []struct {
		env voldriver.Env
	}
	evacuateReturns struct {
		result1 driveradmin.ErrorResponse
	}
	evacuateReturnsOnCall map[int]struct {
		result1 driveradmin.ErrorResponse
	}
	PingStub        func(env voldriver.Env) driveradmin.ErrorResponse
	pingMutex       sync.RWMutex
	pingArgsForCall []struct {
		env voldriver.Env
	}
	pingReturns struct {
		result1 driveradmin.ErrorResponse
	}
	pingReturnsOnCall map[int]struct {
		result1 driveradmin.ErrorResponse
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDriverAdmin) Evacuate(env voldriver.Env) driveradmin.ErrorResponse {
	fake.evacuateMutex.Lock()
	ret, specificReturn := fake.evacuateReturnsOnCall[len(fake.evacuateArgsForCall)]
	fake.evacuateArgsForCall = append(fake.evacuateArgsForCall, struct {
		env voldriver.Env
	}{env})
	fake.recordInvocation("Evacuate", []interface{}{env})
	fake.evacuateMutex.Unlock()
	if fake.EvacuateStub != nil {
		return fake.EvacuateStub(env)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.evacuateReturns.result1
}

func (fake *FakeDriverAdmin) EvacuateCallCount() int {
	fake.evacuateMutex.RLock()
	defer fake.evacuateMutex.RUnlock()
	return len(fake.evacuateArgsForCall)
}

func (fake *FakeDriverAdmin) EvacuateArgsForCall(i int) voldriver.Env {
	fake.evacuateMutex.RLock()
	defer fake.evacuateMutex.RUnlock()
	return fake.evacuateArgsForCall[i].env
}

func (fake *FakeDriverAdmin) EvacuateReturns(result1 driveradmin.ErrorResponse) {
	fake.EvacuateStub = nil
	fake.evacuateReturns = struct {
		result1 driveradmin.ErrorResponse
	}{result1}
}

func (fake *FakeDriverAdmin) EvacuateReturnsOnCall(i int, result1 driveradmin.ErrorResponse) {
	fake.EvacuateStub = nil
	if fake.evacuateReturnsOnCall == nil {
		fake.evacuateReturnsOnCall = make(map[int]struct {
			result1 driveradmin.ErrorResponse
		})
	}
	fake.evacuateReturnsOnCall[i] = struct {
		result1 driveradmin.ErrorResponse
	}{result1}
}

func (fake *FakeDriverAdmin) Ping(env voldriver.Env) driveradmin.ErrorResponse {
	fake.pingMutex.Lock()
	ret, specificReturn := fake.pingReturnsOnCall[len(fake.pingArgsForCall)]
	fake.pingArgsForCall = append(fake.pingArgsForCall, struct {
		env voldriver.Env
	}{env})
	fake.recordInvocation("Ping", []interface{}{env})
	fake.pingMutex.Unlock()
	if fake.PingStub != nil {
		return fake.PingStub(env)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.pingReturns.result1
}

func (fake *FakeDriverAdmin) PingCallCount() int {
	fake.pingMutex.RLock()
	defer fake.pingMutex.RUnlock()
	return len(fake.pingArgsForCall)
}

func (fake *FakeDriverAdmin) PingArgsForCall(i int) voldriver.Env {
	fake.pingMutex.RLock()
	defer fake.pingMutex.RUnlock()
	return fake.pingArgsForCall[i].env
}

func (fake *FakeDriverAdmin) PingReturns(result1 driveradmin.ErrorResponse) {
	fake.PingStub = nil
	fake.pingReturns = struct {
		result1 driveradmin.ErrorResponse
	}{result1}
}

func (fake *FakeDriverAdmin) PingReturnsOnCall(i int, result1 driveradmin.ErrorResponse) {
	fake.PingStub = nil
	if fake.pingReturnsOnCall == nil {
		fake.pingReturnsOnCall = make(map[int]struct {
			result1 driveradmin.ErrorResponse
		})
	}
	fake.pingReturnsOnCall[i] = struct {
		result1 driveradmin.ErrorResponse
	}{result1}
}

func (fake *FakeDriverAdmin) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.evacuateMutex.RLock()
	defer fake.evacuateMutex.RUnlock()
	fake.pingMutex.RLock()
	defer fake.pingMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDriverAdmin) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ driveradmin.DriverAdmin = new(FakeDriverAdmin)
