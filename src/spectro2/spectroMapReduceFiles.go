package spectro2

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"proj3-redesigned/utils"
	"sync"
	"sync/atomic"

	"time"

	"github.com/go-audio/audio"
)

type JsonData struct {
	ImgPath          string
	Duration         int
	SampleRate       int
	MinFreq, MaxFreq float64
	Height, NumTones int
	Contrast         float64
	orderNum         int // Important only for work stealing implementation where thread order != image order
}

type context struct {
	runningThreads int
	allData        [][]int
	prefix         []int
	processed      atomic.Int32
	finalData      []int
}

func MapReduce(filePath string, numThreads int, timing bool) {
	var startTime time.Time
	var elapsed time.Duration
	if timing {
		startTime = time.Now()
	}
	data := ReadJson(filePath)
	numFiles := len(data)
	filesPerThread := int(math.Ceil(float64(numFiles) / float64(numThreads)))
	var wg sync.WaitGroup
	var c context
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	c.allData = make([][]int, numThreads)
	for i := 0; i < numThreads; i++ {
		start := i * filesPerThread
		if start >= numFiles {
			break
		}
		end := start + filesPerThread
		// endSample :=
		if end > numFiles {
			end = numFiles
		}
		wg.Add(1)
		c.runningThreads++
		go func(i, start, end int) {
			defer wg.Done()
			RunMapReduce(i, data[start:end], &c, cond, &mu)
		}(i, start, end)
	}
	wg.Wait()
	if timing {
		elapsed = time.Since(startTime)
		fmt.Printf("%.16f\n", elapsed.Seconds())
	}
}

func ReadJson(filePath string) []JsonData {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var configs []JsonData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configs); err != nil {
		panic(err)
	}
	return configs
}

func RunMapReduce(threadNum int, data []JsonData, c *context, cond *sync.Cond, mu *sync.Mutex) {
	spectra := make([]*Spectro, len(data))
	for i, d := range data {
		spectra[i] = NewSpectro(d.ImgPath, d.Duration, d.SampleRate, d.MinFreq, d.MaxFreq, d.Height, d.NumTones, d.Contrast)
		spectra[i].processImage()
		if i > 0 {
			spectra[0].buf.Data = append(spectra[0].buf.Data, spectra[i].buf.Data...)
		}
	}
	c.allData[threadNum] = spectra[0].buf.Data
	mu.Lock()
	c.runningThreads--
	if c.runningThreads == 0 {
		finalDataLen := 0
		for _, s := range c.allData {
			finalDataLen += len(s)
		}
		c.finalData = make([]int, finalDataLen)

		cond.Broadcast()
	} else {
		cond.Wait()
	}
	c.runningThreads++
	mu.Unlock()

	start := 0
	for j := 0; j < threadNum; j++ {
		start += len(c.allData[j])
	}
	for k, sample := range spectra[0].buf.Data {
		c.finalData[k+start] = sample
	}

	mu.Lock()
	c.runningThreads--
	if c.runningThreads == 0 {
		ExportWav(
			&audio.IntBuffer{
				Format: &audio.Format{
					NumChannels: 1,
					SampleRate:  spectra[0].buf.Format.SampleRate,
				},
				Data:           c.finalData,
				SourceBitDepth: 16,
			}, "map_reduce")
	}
	mu.Unlock()
}

func (s *Spectro) processImage() {
	wave := make([]float64, s.WaveGen.NumSamples)
	for y := 0; y < s.Img.Height; y++ {
		row := s.Img.IntensityRow(y)
		utils.AddSliceElems(wave, s.WaveGen.RowToWave(y, row))
	}
	s.processAndScale(wave)
}
