package main

import (
	"context"
	"sync"
)

type controller struct {
	chans ProcChannels
}

type ProcChannels struct {
	newPros  chan *Process
	oldPros  chan *Process
	Killall  chan interface{}
	DoneChan chan *Process
}

func (p *ProcChannels) init() {
	p.newPros = make(chan *Process)
	p.oldPros = make(chan *Process)
	p.Killall = make(chan interface{})
	p.DoneChan = make(chan *Process)
}

func (c *controller) run(waitchan chan interface{}) {
	var wg sync.WaitGroup
	maplock := make(chan interface{}, 1)
	maplock <- 1
	envlock := make(chan interface{}, 1)
	envlock <- 1
	ctx := context.Background()
	cancelMap := map[string]context.CancelFunc{}

	go c.checkDoneChan(&cancelMap, &maplock)

	for {
		select {
		case newPros := <-c.chans.newPros:
			c.createProcess(newPros, &ctx, &cancelMap, &maplock, &envlock, &wg)
		case oldPros := <-c.chans.oldPros:
			c.stopProcess(oldPros, &cancelMap, &maplock)
		case <-c.chans.Killall:
			c.killAll(&cancelMap, &maplock, &wg, waitchan)
			return
		}
	}
}

func (c *controller) checkDoneChan(cancelMap *map[string]context.CancelFunc, maplock *chan interface{}) {
	for {
		done := <-c.chans.DoneChan
		<-*maplock
		delete(*cancelMap, done.Name)
		*maplock <- 1
	}
}

func (c *controller) createProcess(newPros *Process, ctx *context.Context, cancelMap *map[string]context.CancelFunc, maplock *chan interface{}, envlock *chan interface{}, wg *sync.WaitGroup) {
	<-*maplock
	if _, ok := (*cancelMap)[newPros.Name]; ok {
		logger.Println("Process already running.  Not restarting:", newPros.Name)
	} else {
		logger.Println("Running process:", newPros.Name)
		*ctx, (*cancelMap)[newPros.Name] = context.WithCancel(*ctx)
		wg.Add(1)
		go ProcessContainer(*ctx, newPros, wg, *envlock, c.chans.DoneChan)
	}
	*maplock <- 1
}

func (c *controller) stopProcess(oldPros *Process, cancelMap *map[string]context.CancelFunc, maplock *chan interface{}) {
	logger.Println("Canceling process:", oldPros.Name)
	<-*maplock
	cancel := (*cancelMap)[oldPros.Name]
	if cancel != nil {
		cancel()
	} else {
		logger.Println("Unable to cancel process:", oldPros.Name)
	}
	*maplock <- 1
}

func (c *controller) killAll(cancelMap *map[string]context.CancelFunc, maplock *chan interface{}, wg *sync.WaitGroup, waitchan chan interface{}) {
	logger.Println("Killing all processes")
	<-*maplock
	for name, f := range *cancelMap {
		f()
		delete(*cancelMap, name)
	}
	*maplock <- 1
	wg.Wait()
	waitchan <- 1
}
