# changroup <br> [![go minimal version][go-img]][go-url] [![go tested version][go-latest-img]][go-latest-url] [![CI][ci-img]][ci-url] [![Codecov][codecov-img]][codecov-url] [![Maintainability][codeclimate-img]][codeclimate-url] [![Go Report Card][goreportcard-img]][goreportcard-url] [![License][license-img]][license-url] [![Go Reference][godoc-img]][godoc-url]


`changroup` is a Go library to create a group of channels (publish/subscribe pattern). A value is sent to each channel in the group. Channels can be acquired/released dynamically.

`changroup.Group` allows to acquire/release channel and to send a value to all acquired channels.

`changroup.AckableGroup` does the same, but sends `changroup.Ackable` value. It calls original ack function only after all subscribers acked their copy of value. It's useful if you need to know when the message is processed.


## Generics

The minimal supported go version is 1.19 because the library uses generics.

## Installation

```shell
go get github.com/maratori/changroup
```

## Usage

### Scenario 1

Create all channels before sending values (before publisher starts). In this case all channels are guaranteed to receive all values.

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/maratori/changroup"
)

func main() {
	group := changroup.NewGroup[int]()
	ch1, _ := group.Acquire()
	ch2, _ := group.Acquire()

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		defer group.ReleaseAll() // close all channels and remove from group

		for i := 0; i < 100; i++ {
			group.Send(i)
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		defer wg.Done()
		for i := range ch1 {
			fmt.Println("subscriber 1 received", i)
		}
		fmt.Println("ch1 is closed because group.ReleaseAll() is called")
	}()

	go func() {
		defer wg.Done()
		for i := range ch2 {
			fmt.Println("subscriber 2 received", i)
		}
		fmt.Println("ch2 is closed because group.ReleaseAll() is called")
	}()

	wg.Wait()
}
```

### Scenario 2

Create channels dynamically (after publisher started). In this case some values may be dropped because there are no subscribers at the moment.

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/maratori/changroup"
)

func main() {
	group := changroup.NewGroup[int]()

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		defer group.ReleaseAll() // close all channels and remove from group

		for i := 0; i < 100; i++ {
			group.Send(i)
			time.Sleep(1 * time.Second)
		}
	}()

	// subscribe, wait for a specific value, then unsubscribe
	go func() {
		defer wg.Done()
		ch, release := group.Acquire()
		defer release() // close channel and remove from group

		for i := range ch {
			fmt.Println("subscriber 1 received", i)
			if i == 20 {
				// do something
				return
			}
		}
	}()

	// read all values
	go func() {
		defer wg.Done()
		ch, _ := group.Acquire()
		for i := range ch {
			fmt.Println("subscriber 2 received", i)
		}
		fmt.Println("ch is closed because group.ReleaseAll() is called")
	}()

	wg.Wait()
}
```

### Scenario 3

Do something in publisher after ack from all subscribers.

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/maratori/changroup"
)

func main() {
	group := changroup.NewAckableGroup[int]()
	ch1, _ := group.Acquire()
	ch2, _ := group.Acquire()

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		defer group.ReleaseAll() // close all channels and remove from group

		for i := 0; i < 100; i++ {
			group.Send(changroup.NewAckable(i, func() {
				fmt.Println("publisher received acks from all subscribers for i =", i)
			}))
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		defer wg.Done()
		for a := range ch1 {
			fmt.Println("subscriber 1 received", a.Value)
			a.Ack() // value is processed
		}
		fmt.Println("ch1 is closed because group.ReleaseAll() is called")
	}()

	go func() {
		defer wg.Done()
		for a := range ch2 {
			fmt.Println("subscriber 2 received", a.Value)
			a.Ack() // value is processed
		}
		fmt.Println("ch2 is closed because group.ReleaseAll() is called")
	}()

	wg.Wait()
}
```

## Contribution

You are welcome to create an issue or pull request with improvements and fixes. See [guide](/.github/CONTRIBUTING.md).

## License

[MIT License][license-url]


[go-img]: https://img.shields.io/github/go-mod/go-version/maratori/changroup
[go-url]: /go.mod
[go-latest-img]: https://img.shields.io/github/go-mod/go-version/maratori/changroup?filename=.github%2Flatest-deps%2Fgo.mod&label=tested
[go-latest-url]: /.github/latest-deps/go.mod
[ci-img]: https://github.com/maratori/changroup/actions/workflows/ci.yml/badge.svg
[ci-url]: https://github.com/maratori/changroup/actions/workflows/ci.yml
[codecov-img]: https://codecov.io/gh/maratori/changroup/branch/main/graph/badge.svg?token=vbDpr5rl0h
[codecov-url]: https://codecov.io/gh/maratori/changroup
[codeclimate-img]: https://api.codeclimate.com/v1/badges/ff2cd8265ab506c847d4/maintainability
[codeclimate-url]: https://codeclimate.com/github/maratori/changroup/maintainability
[goreportcard-img]: https://goreportcard.com/badge/github.com/maratori/changroup
[goreportcard-url]: https://goreportcard.com/report/github.com/maratori/changroup
[license-img]: https://img.shields.io/github/license/maratori/changroup.svg
[license-url]: /LICENSE
[godoc-img]: https://pkg.go.dev/badge/github.com/maratori/changroup.svg
[godoc-url]: https://pkg.go.dev/github.com/maratori/changroup
