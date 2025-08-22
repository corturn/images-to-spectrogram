package spectro2

import (
	"fmt"
	"time"
)

func ImgToWavSeq(filePath string, timing bool) {

	var startTime time.Time
	var elapsed time.Duration
	if timing {
		startTime = time.Now()
	}

	data := ReadJson(filePath)

	if len(data) == 0 {
		return
	}

	var initS *Spectro

	for i := range data {
		d := data[i]
		s := NewSpectro(d.ImgPath, d.Duration, d.SampleRate, d.MinFreq, d.MaxFreq, d.Height, d.NumTones, d.Contrast)
		s.processImage()
		if i == 0 {
			initS = s
		} else {
			initS.buf.Data = append(initS.buf.Data, s.buf.Data...)
		}

	}

	if err := ExportWav(initS.buf, "seq"); err != nil {
		panic(err)
	}
	if timing {
		elapsed = time.Since(startTime)
		fmt.Printf("%.16f\n", elapsed.Seconds())
	}
}
