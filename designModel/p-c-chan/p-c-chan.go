package p_c_chan

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
	for{
		duck := Duck{Number:rand.Intn(1000),Weight:rand.Float64()}
		fmt.Println("produce num:" + strconv.Itoa(duck.Number))
		b.ducks <- duck
	}
}
func (c *ContosmerImpl)buyDuck(b *Basket)  {
	for{
		duck, ok :=  <- b.ducks
		if ok {
			fmt.Println("consume num:" + strconv.Itoa(duck.Number))
		}

	}
}
func main()  {
	wg := new(sync.WaitGroup)
	p := new(ProducerImpl)
	c := new(ContosmerImpl)
	b:= &Basket{
		ducks: make(chan Duck),
	}
	wg.Add(2)
	go func() {
		p.makeDuck(b)
		wg.Done()
	}()
	go func() {
		c.buyDuck(b)
		wg.Done()
	}()
	wg.Wait()
}