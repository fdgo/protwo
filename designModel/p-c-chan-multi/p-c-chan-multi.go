package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

type (
	Producer interface {
		makeDuck()
	}
	Contosmer interface {
		buyDuck()
	}
	Basket struct {
		ducks chan Duck
	}

	ProducerImpl struct {
		Producer
	}
	ContosmerImpl struct {
		Contosmer
	}
	Duck struct {
		Number int
		Weight float64
	}
)

func (p *ProducerImpl)makeDuck(b * Basket)  {
	duck := Duck{Number:rand.Intn(1000),Weight:rand.Float64()}
	fmt.Println("produce num:" + strconv.Itoa(duck.Number))
	b.ducks <- duck
}
func (c *ContosmerImpl)buyDuck(b *Basket)  {
	duck, ok :=  <- b.ducks
	if ok {
		fmt.Println("consume num:" + strconv.Itoa(duck.Number))
	}
}
func main()  {
	wg := new(sync.WaitGroup)
	p := new(ProducerImpl)
	c := new(ContosmerImpl)
	b:= &Basket{
		ducks: make(chan Duck,100),
	}
	for i:=0;i<50000;i++{
		wg.Add(1)
		go func() {
			p.makeDuck(b)
			wg.Done()
		}()
	}
	for j:=0;j<50000;j++{
		wg.Add(1)
		go func() {
			c.buyDuck(b)
			wg.Done()
		}()
	}
	wg.Wait()
}