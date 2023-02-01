package changroup_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/maratori/changroup"
)

func ExampleGroup_subscribeBeforePublish() {
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

func ExampleGroup_subscribeDynamically() {
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

func ExampleAckableGroup() {
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
