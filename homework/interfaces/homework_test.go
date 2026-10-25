package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type UserService struct {
	// not need to implement
	NotEmptyStruct bool
}
type MessageService struct {
	// not need to implement
	NotEmptyStruct bool
}

type Container struct {
	constructors map[string]interface{}
}

func NewContainer() *Container {
	return &Container{
		constructors: make(map[string]interface{}),
	}
}

func (c *Container) RegisterType(name string, constructor interface{}) {
	c.constructors[name] = constructor
}

func (c *Container) Resolve(name string) (interface{}, error) {
	constructor, ok := c.constructors[name]
	if !ok {
		return nil, fmt.Errorf("type %q is not registered in container", name)
	}

	fn := reflect.ValueOf(constructor)
	if fn.Kind() != reflect.Func || fn.Type().NumIn() != 0 {
		return nil, fmt.Errorf("constructor for %q must be a func without arguments", name)
	}

	numOut := fn.Type().NumOut()
	if numOut < 1 || numOut > 2 {
		return nil, fmt.Errorf("constructor for %q must return (interface{}) or (interface{}, error)", name)
	}
	if numOut == 2 && !fn.Type().Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return nil, fmt.Errorf("second return value of constructor for %q must be error", name)
	}

	results := fn.Call(nil)
	if err, ok := results[numOut-1].Interface().(error); numOut == 2 && ok && err != nil {
		return nil, fmt.Errorf("failed to construct %q: %w", name, err)
	}

	obj := results[0].Interface()
	if obj == nil {
		return nil, fmt.Errorf("constructor for %q returned nil", name)
	}
	return obj, nil
}

func TestDIContainer(t *testing.T) {
	container := NewContainer()
	container.RegisterType("UserService", func() interface{} {
		return &UserService{}
	})
	container.RegisterType("MessageService", func() interface{} {
		return &MessageService{}
	})

	userService1, err := container.Resolve("UserService")
	assert.NoError(t, err)
	userService2, err := container.Resolve("UserService")
	assert.NoError(t, err)

	u1 := userService1.(*UserService)
	u2 := userService2.(*UserService)
	assert.False(t, u1 == u2)

	messageService, err := container.Resolve("MessageService")
	assert.NoError(t, err)
	assert.NotNil(t, messageService)

	paymentService, err := container.Resolve("PaymentService")
	assert.Error(t, err)
	assert.Nil(t, paymentService)
}

func TestDIContainer_ConstructorWithError(t *testing.T) {
	container := NewContainer()
	container.RegisterType("FailingService", func() (interface{}, error) {
		return nil, errors.New("construction failed")
	})

	svc, err := container.Resolve("FailingService")
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestDIContainer_ReRegisterOverwrites(t *testing.T) {
	container := NewContainer()
	container.RegisterType("UserService", func() interface{} {
		return &UserService{NotEmptyStruct: true}
	})

	svc, err := container.Resolve("UserService")
	assert.NoError(t, err)
	assert.True(t, svc.(*UserService).NotEmptyStruct)
}

func TestDIContainer_InvalidConstructors(t *testing.T) {
	container := NewContainer()

	// not a function
	container.RegisterType("NotAFunc", 42)
	_, err := container.Resolve("NotAFunc")
	assert.Error(t, err)

	// function with arguments
	container.RegisterType("WithArgs", func(x int) interface{} { return x })
	_, err = container.Resolve("WithArgs")
	assert.Error(t, err)

	// function without return values
	container.RegisterType("NoReturn", func() {})
	_, err = container.Resolve("NoReturn")
	assert.Error(t, err)
}
