package promise

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var promise = New(func(resolve func(Any), reject func(error)) {
		resolve(nil)
	})

	if promise == nil {
		t.Error("Promise is nil")
	}
}

func TestPromise_Then(t *testing.T) {
	var promise = New(func(resolve func(Any), reject func(error)) {
		resolve(1 + 1)
	})

	promise.
		Then(func(data Any) (Any, error) {
			return data.(int) + 1, nil
		}).
		Then(func(data Any) (Any, error) {
			if data.(int) != 3 {
				t.Error("Result doesn't propagate")
			}
			return data, nil
		}).
		Catch(func(err error) (Any, error) {
			t.Error("Catch triggered in .Then test")
			return nil, err
		})

	promise.Await()
}

func TestPromise_ThenError(t *testing.T) {
	var promise = New(func(resolve func(Any), reject func(error)) {
		resolve(1 + 1)
	})

	promise.
		Then(func(data Any) (Any, error) {
			return nil, errors.New("error in .Then handler")
		}).
		Catch(func(err error) (Any, error) {
			if err.Error() != "error in .Then handler" {
				t.Error("Error from .Then handler didn't propagate")
			}
			return nil, err
		})

	promise.Await()
}

func TestPromise_ThenNested(t *testing.T) {
	var promise = New(func(resolve func(Any), reject func(error)) {
		resolve(New(func(res func(Any), rej func(error)) {
			res("Hello, World")
		}))
	})

	promise.
		Then(func(data Any) (Any, error) {
			if data.(string) != "Hello, World" {
				t.Error("Resolved promise doesn't flatten")
			}
			return data, nil
		}).
		Catch(func(err error) (Any, error) {
			t.Error("Catch triggered in .Then test")
			return nil, err
		})

	promise.Await()
}

func TestPromise_Catch(t *testing.T) {
	var promise = New(func(resolve func(Any), reject func(error)) {
		reject(errors.New("very serious err"))
	})

	promise.
		Then(func(data Any) (Any, error) {
			t.Error("Then 1 triggered in .Catch test")
			return data, nil
		}).
		Catch(func(err error) (Any, error) {
			if err.Error() == "very serious err" {
				return nil, errors.New("dealing with err at this stage")
			}
			return nil, err
		}).
		Catch(func(err error) (Any, error) {
			if err.Error() != "dealing with err at this stage" {
				t.Error("Error doesn't propagate")
			}
			return nil, err
		}).
		Then(func(data Any) (Any, error) {
			t.Error("Then 2 triggered in .Catch test")
			return data, nil
		})

	promise.Await()
}

func TestPromise_CatchNested(t *testing.T) {
	var promise = New(func(resolve func(Any), reject func(error)) {
		resolve(New(func(res func(Any), rej func(error)) {
			rej(errors.New("nested fail"))
		}))
	})

	promise.
		Then(func(data Any) (Any, error) {
			t.Error("Then triggered in .Catch test")
			return data, nil
		}).
		Catch(func(err error) (Any, error) {
			if err.Error() != "nested fail" {
				t.Error("Rejected promise doesn't flatten")
			}
			return nil, err
		})

	promise.Await()
}

func TestPromise_CatchRecovers(t *testing.T) {
	var promise = New(func(resolve func(Any), reject func(error)) {
		reject(errors.New("error in promise"))
	})

	promise.
		Catch(func(err error) (Any, error) {
			if err.Error() != "error in promise" {
				t.Error("Promise didn't raise")
			}
			return 3, nil
		}).
		Then(func(data Any) (Any, error) {
			if data.(int) != 3 {
				t.Error("Catch didn't recover")
			}
			return nil, nil
		})

	promise.Await()
}

func TestPromise_Panic(t *testing.T) {
	var promise = New(func(resolve func(Any), reject func(error)) {
		panic("much panic")
	})

	promise.
		Then(func(data Any) (Any, error) {
			t.Error("Then triggered in .Catch test")
			return data, nil
		})

	promise.Await()
}

func TestPromise_Await(t *testing.T) {
	var promises = make([]*Promise, 10)

	for x := 0; x < 10; x++ {
		var promise = New(func(resolve func(Any), reject func(error)) {
			resolve(time.Now())
		})

		promise.Then(func(data Any) (Any, error) {
			return data.(time.Time).Add(time.Second).Nanosecond(), nil
		})

		promises[x] = promise
	}

	var promise1 = Resolve("WinRAR")
	var promise2 = Reject(errors.New("fail"))

	for _, p := range promises {
		_, err := p.Await()

		if err != nil {
			t.Error(err)
		}
	}

	result, err := promise1.Await()
	if err != nil && result != "WinRAR" {
		t.Error(err)
	}

	result, err = promise2.Await()
	if err == nil {
		t.Error(err)
	}
}

func TestPromise_Resolve(t *testing.T) {
	var promise = Resolve(123).
		Then(func(data Any) (Any, error) {
			return data.(int) + 1, nil
		}).
		Then(func(data Any) (Any, error) {
			t.Helper()
			if data.(int) != 124 {
				t.Errorf("Then resolved with unexpected value: %v", data.(int))
			}
			return nil, nil
		})

	promise.Await()
}

func TestPromise_Reject(t *testing.T) {
	var promise = Reject(errors.New("rejected")).
		Then(func(data Any) (Any, error) {
			return data.(int) + 1, nil
		}).
		Catch(func(err error) (Any, error) {
			if err.Error() != "rejected" {
				t.Errorf("Catch rejected with unexpected value: %v", err)
			}
			return nil, err
		})

	promise.Await()
}

func TestPromise_All(t *testing.T) {
	var promises = make([]*Promise, 10)

	for x := 0; x < 10; x++ {
		if x == 8 {
			promises[x] = Reject(errors.New("bad promise"))
			continue
		}

		promises[x] = Resolve("All Good")
	}

	_, err := All(promises...).Await()
	if err == nil {
		t.Error("Combined promise failed to return single err")
	}
}

func TestPromise_All2(t *testing.T) {
	var promises = make([]*Promise, 10)

	for index := 0; index < 10; index++ {
		promises[index] = Resolve(fmt.Sprintf("All Good %d", index))
	}

	result, err := All(promises...).Await()
	if err != nil {
		t.Error(err)
	} else {
		for index, res := range result.([]Any) {
			s := fmt.Sprintf("All Good %d", index)
			if res == nil {
				t.Error("Result is nil!")
				return
			}
			if res.(string) != s {
				t.Error("Wrong index!")
				return
			}
		}
	}
}

func TestPromise_All3(t *testing.T) {
	var promises []*Promise

	result, err := All(promises...).Await()
	if err != nil {
		t.Error(err)
		return
	}

	res := result.([]Any)
	if len(res) != 0 {
		t.Error("Wrong result on nil slice")
		return
	}
}

func TestPromise_AllSettled(t *testing.T) {
	var promises = make([]*Promise, 10)

	for x := 0; x < 10; x++ {
		if x == 8 {
			promises[x] = Reject(errors.New("bad promise"))
			continue
		}

		promises[x] = Resolve("All Good")
	}

	_, err := AllSettled(promises...).Await()
	if err != nil {
		t.Error("Combined promise failed to reject on singular error")
	}
}

func TestPromise_Race1(t *testing.T) {
	var p1 = Resolve("Promise 1")
	var p2 = Resolve("Promise 2")

	_, err := Race(p1, p2).Await()
	if err != nil {
		t.Error("Combined promise failed for some reason")
	}
}

func TestPromise_Race2(t *testing.T) {
	var p1 = Reject(errors.New("Promise 1"))
	var p2 = Reject(errors.New("Promise 2"))

	_, err := Race(p1, p2).Await()
	if err == nil {
		t.Error("Combined promise failed to account for a rejection in a race")
	}
}
