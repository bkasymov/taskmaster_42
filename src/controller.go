package main

import (
	"context"
	"sync"
)

type controller struct {
	chans ProcChannels
}

/***
 * Контроллер используется для остановки / запуска процессов
 * chans - каналы для передачи информации о том, какие каналы остановить или запустить
 * newPros - канал для запуска новых процессов
 * oldPros - канал для остановки старых процессов
 * Killall - канал для остановки всех процессов
 * DoneChan - канал для завершения процессов
 * ProcChannels - используется для передачи информации о том, какие каналы остановить или запустить
 */

type ProcChannels struct {
	newPros  chan *Process
	oldPros  chan *Process
	Killall  chan interface{}
	DoneChan chan *Process
}

/***
 * Инициализация каналов для передачи информации о том, какие каналы остановить или запустить
 */

func (p *ProcChannels) init() {
	p.newPros = make(chan *Process)
	p.oldPros = make(chan *Process)
	p.Killall = make(chan interface{})
	p.DoneChan = make(chan *Process)
}

/***
 * Запуск контроллера процессов
 * waitchan - канал для ожидания завершения всех процессов
 * maplock - блокировка для cancelMap (map[string]context.CancelFunc)
 * envlock - блокировка для env (map[string]string)
 * ctx - контекст для отмены процессов
 * cancelMap - процессы и функции отмены
 * wg - ожидание завершения всех процессов
 * doneChan - канал для отслеживания завершения процессов
 * newPros - канал для запуска новых процессов
 * oldPros - канал для остановки процессов
 * Killall - канал для остановки всех процессов
 */

func (c *controller) run(waitchan chan interface{}) {
	var wg sync.WaitGroup
	maplock := make(chan interface{}, 1)
	maplock <- 1
	envlock := make(chan interface{}, 1)
	envlock <- 1
	ctx := context.Background()
	cancelMap := map[string]context.CancelFunc{}

	// для отслеживания завершения процессов
	go c.checkDoneChan(&cancelMap, &maplock)

	/***
	 * Ожидание новых процессов, остановки старых процессов или остановки всех процессов и запуск их в отдельной горутине (если процесса с таким именем нет)
	 */
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

/***
 * Проверка завершения процессов
 */
func (c *controller) checkDoneChan(cancelMap *map[string]context.CancelFunc, maplock *chan interface{}) {
	for {
		done := <-c.chans.DoneChan
		<-*maplock
		delete(*cancelMap, done.Name)
		*maplock <- 1
	}
}

/***
 * Создание процесса и запуск его в отдельной горутине (если процесса с таким именем нет)
 */

func (c *controller) createProcess(newPros *Process, ctx *context.Context, cancelMap *map[string]context.CancelFunc, maplock *chan interface{}, envlock *chan interface{}, wg *sync.WaitGroup) {
	<-*maplock
	if _, ok := (*cancelMap)[newPros.Name]; ok {
		logger.Println("Process already running.  Not restarting:", newPros.Name)
	} else {
		logger.Println("Running process:", newPros.Name)
		*ctx, (*cancelMap)[newPros.Name] = context.WithCancel(*ctx)
		wg.Add(1)
		go ProcessContainer(*ctx, newPros, wg, envlock, c.chans.DoneChan)
	}
	*maplock <- 1
}

/***
 * Остановка процесса (если процесс с таким именем есть)
 */

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

/***
 * Остановка всех процессов (если процесс с таким именем есть) и ожидание завершения всех процессов (wg.Wait()) и завершение работы контроллера (waitchan <- 1)
 * cancelMap - процессы и функции отмены
 * maplock - блокировка для cancelMap (map[string]context.CancelFunc)
 * wg - ожидание завершения всех процессов
 * waitchan - канал для ожидания завершения всех процессов
 */

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
