package spectro2

import (
	"math"
	"os"
	"proj3-redesigned/utils"
	"proj3-redesigned/wavIO"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type Spectro struct {
	Img     *wavIO.Image
	width   int
	WaveGen *WaveGen
	buf     *audio.IntBuffer
}

func NewSpectro(imgPath string, duration int, sampleRate int,
	minFreq, maxFreq float64,
	height, numTones int, contrast float64) *Spectro {

	s := &Spectro{}

	s.LoadImage(imgPath, height)

	s.buf = &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  sampleRate,
		},
		SourceBitDepth: 16,
	}

	s.width = s.Img.Width
	s.WaveGen = NewWaveGen(duration, sampleRate, s.Img.Width, s.Img.Height, minFreq, maxFreq, numTones, contrast)
	return s
}

func (s *Spectro) LoadImage(imgPath string, height int) {
	s.Img = &wavIO.Image{}
	s.Img.Load(imgPath, height)
}

func (s *Spectro) processAndScale(wave []float64) {
	// Converts output of wave functions to proper format for .wav file, puts
	// in buffer for export
	scalingFactor := 0.0
	for _, sample := range wave {
		absSample := math.Abs(sample)
		if absSample > scalingFactor {
			scalingFactor = absSample
		}
	}

	// Avoid division by zero if the wave is all 0.
	if scalingFactor == 0 {
		scalingFactor = 1
	}

	// Calculate scaling factor to normalize to a maximum amplitude of 32767.
	scale := 32767.0 / scalingFactor

	s.buf.Data = make([]int, len(wave))
	// Multiply each sample by the scale, convert to int, and add to .wav buffer data
	for i, samples := range wave {
		s.buf.Data[i] = int(samples * scale)
	}
}

func ExportWav(buf *audio.IntBuffer, name string) error {
	outFile, err := os.Create("output_" + name + ".wav")
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create a new WAV encoder.
	encoder := wav.NewEncoder(outFile, buf.Format.SampleRate, 16, 1, 1)

	// Write the audio buffer to the file.
	if err := encoder.Write(buf); err != nil {
		return err
	}

	// Close the encoder to finalize the WAV file.
	if err := encoder.Close(); err != nil {
		return err
	}
	return nil

}

type WaveGen struct {
	duration         float64 // length in seconds (may slightly different from )
	NumSamples       int     // number of samples in wave
	samplesPerPx     int     // number of samples per pixel
	width, height    int     // width and height of image
	minFreq, maxFreq float64
	linspace         []float64 // time value slice for each individual sample,
	// evenly spaced 0 -> (sampleRate * duration)
	numTones  int     // each vertical pixel filled with numTones number of tones to fill out sound/reduce unnatural freq gaps
	tonesDist float64 // distance in frequency between each tone that makes up a single vertical pixel
	contrast  float64
}

func NewWaveGen(duration, sampleRate, width, height int, minFreq, maxFreq float64, numTones int, contrast float64) *WaveGen {
	WaveGen := &WaveGen{width: width, height: height, minFreq: minFreq, maxFreq: maxFreq, numTones: numTones, contrast: contrast}

	// A bit of a roundabout way to get samples and duration, but this helps ensure
	// that when rounding occurs, the total duration of the audio file is rounded
	// (which matters less to the image produced) rather than the number of samples
	WaveGen.samplesPerPx = int(float64(duration*sampleRate) / float64(width))
	WaveGen.NumSamples = WaveGen.samplesPerPx * width
	WaveGen.duration = float64(WaveGen.NumSamples) / float64(sampleRate)

	WaveGen.linspace = utils.LinspaceInt(0, WaveGen.duration, WaveGen.NumSamples)

	WaveGen.tonesDist = (maxFreq - minFreq) / float64(height) / float64(numTones)

	return WaveGen
}

func (row *WaveGen) RowToWave(y int, r []float64) []float64 {
	// Converts a pixel row of intensity values to wave vals
	// y is the height of the row in px in the image

	// Each pixel row is assigned a base frequency value to map it vertically onto spectrogram.
	heightFlt := float64(row.height)
	pxHeightFlt := float64(row.height - y)
	baseFreq := (row.maxFreq-row.minFreq)/heightFlt*pxHeightFlt + row.minFreq

	start := 0
	end := start + row.samplesPerPx

	wave := row.pxToWave(baseFreq, r[0], row.linspace[start:end])
	for x := 1; x < row.width; x++ {
		start = x * row.samplesPerPx
		end = start + row.samplesPerPx
		wave = append(wave, row.pxToWave(baseFreq, r[x], row.linspace[start:end])...)
	}
	return wave
}

func (row *WaveGen) pxToWave(baseFreq, amplitude float64, times []float64) []float64 {
	wave := row.wave(baseFreq, amplitude, times)
	for i := 1; i < row.numTones; i++ {
		baseFreq += row.tonesDist
		utils.AddSliceElems(wave, row.wave(baseFreq, amplitude, times))
	}
	return wave
}

func (row *WaveGen) wave(freq, amplitude float64, times []float64) []float64 {
	wave := make([]float64, len(times))
	// phase := math.Pi / 2.0
	for i, time := range times {
		wave[i] = math.Pow(amplitude, row.contrast) * math.Sin(freq*time*2*math.Pi)
	}

	return wave
}
