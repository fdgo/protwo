package p_c

import (
	"container/list"
	"fmt"
	"math/rand"
	"runtime"
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
		ducks *list.List
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
		for b.ducks.Len() > 0{
			runtime.Gosched()
		}
		duck :=&Duck{
			Number:rand.Intn(10000),
			Weight:rand.Float64(),
		}

		b.ducks.PushBack(duck)
		fmt.Println("produce:"+strconv.Itoa(duck.Number))
	}
}
func (c *ContosmerImpl)buyDuck(b *Basket)  {
	for{
		for b.ducks.Len() <= 0{
			runtime.Gosched()
		}
		element := b.ducks.Back()
		duck := element.Value.(*Duck)
		fmt.Println("contosmer:"+strconv.Itoa(duck.Number))
		b.ducks.Remove(element)
	}
}
func main()  {
	wg := new(sync.WaitGroup)
	p := new(ProducerImpl)
	c := new(ContosmerImpl)
	b:= &Basket{
		ducks:list.New(),
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