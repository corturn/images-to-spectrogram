package spectro2

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-audio/audio"
)

type Work struct {
	dqs []*DeQueue
}

func WorkSteal(filePath string, numThreads int, timing bool) {
	var startTime time.Time
	var elapsed time.Duration
	if timing {
		startTime = time.Now()
	}
	data := ReadJson(filePath)

	for i := range data {
		data[i].orderNum = i
	}

	numFiles := len(data)
	filesPerThread := int(math.Ceil(float64(numFiles) / float64(numThreads)))
	var wg sync.WaitGroup
	var c context
	var mu sync.Mutex
	var work Work
	work.dqs = make([]*DeQueue, numThreads)
	for i := 0; i < numThreads; i++ {
		work.dqs[i] = NewDeQueue()
	}
	cond := sync.NewCond(&mu)
	c.allData = make([][]int, numFiles)
	c.prefix = make([]int, numFiles)
	c.runningThreads = numThreads
	for i := 0; i < numThreads; i++ {
		start := i * filesPerThread
		if start >= numFiles {
			c.runningThreads -= numThreads - i
			// In case a thread had finished before the update above
			cond.Signal()
			break
		}
		end := start + filesPerThread
		if end > numFiles {
			end = numFiles
		}
		wg.Add(1)
		go func(i int, start int, end int, cnd *sync.Cond, wrk *Work) {
			defer wg.Done()
			RunWorkSteal(i, data[start:end], &c, cnd, &mu, wrk)
		}(i, start, end, cond, &work)
	}
	wg.Wait()

	if timing {
		elapsed = time.Since(startTime)
		fmt.Printf("%.16f\n", elapsed.Seconds())
	}
}

func RunWorkSteal(threadNum int, data []JsonData, c *context, cond *sync.Cond, mu *sync.Mutex, work *Work) {
	for i := range data {
		work.dqs[threadNum].PushBottom(&data[i])
	}

	for {
		d := work.dqs[threadNum].PopBottom()

		if d == nil {
			// Steal, continue, else break
			d = Steal(threadNum, work.dqs)
			if d == nil {
				break
			}
		}

		s := NewSpectro(d.ImgPath, d.Duration, d.SampleRate, d.MinFreq, d.MaxFreq, d.Height, d.NumTones, d.Contrast)
		s.processImage()
		c.allData[d.orderNum] = s.buf.Data

	}

	mu.Lock()
	c.runningThreads--
	if c.runningThreads == 0 {
		finalDataLen := 0
		for j, s := range c.allData {
			c.prefix[j] = finalDataLen
			finalDataLen += len(s)
		}
		c.finalData = make([]int, finalDataLen)

		cond.Broadcast()
	} else {
		cond.Wait()
	}
	c.runningThreads++
	mu.Unlock()

	for {
		j := int(c.processed.Add(1))
		j -= 1
		if j >= len(c.allData) {
			break
		}
		start := c.prefix[j]
		for k, sample := range c.allData[j] {
			c.finalData[k+start] = sample
		}
	}

	mu.Lock()
	c.runningThreads--
	if c.runningThreads == 0 {
		ExportWav(
			&audio.IntBuffer{
				Format: &audio.Format{
					NumChannels: 1,
					SampleRate:  44100,
				},
				Data:           c.finalData,
				SourceBitDepth: 16,
			}, "work_steal")
	}
	mu.Unlock()
}

func Steal(threadNum int, dqs []*DeQueue) *JsonData {

	n := len(dqs)
	for i := 0; i < n; i++ {
		if i == threadNum {
			continue
		}
		if stolen := dqs[i].PopTop(); stolen != nil {
			return stolen
		}
	}
	return nil
}
