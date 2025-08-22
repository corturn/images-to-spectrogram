This project was originally written for a Parallel Computing course.

My project takes in png images as input and exports an audio file where the spectrogram of that
file (a kind of 3d graph representation of audio with time on x axis, frequency on y axis,
amplitude as color intensity) looks like the input images. Originally, my project was going to be
parallelizing the process of converting a single image to a spectrogram, but I realized this did not
mesh well with the project criteria for my Parallel Computing class. Without going into too much detail,
the largest problem was there wasn't any performance improvement with the dequeue as each task (a single row of
pixels) was the same amount of work. I therefore adjusted my original problem to a slightly more
contrived one: drawing multiple images to the same audio file to produce a panoramic view.
Rather than parallelizing each row of a single image, I parallelized the processing of multiple
different images.


Project Description:

A JSON file input provides the program with information on what images to use and parameters
for the output:

[
{
"ImgPath": "images/img1.png",
"Duration": 20,
"SampleRate": 44100,
"MinFreq": 100,
"MaxFreq": 20000,
"Height": 900,
"NumTones": 2,
"Contrast": 1
},
...
]

Duration specifies the length in seconds that image should cover in the audio file, the sample rate
is the sample rate of the output audio file (my program only really supports 44100 hz at the
moment), min and max freq specifies the frequency range the image should cover in the output,
height scales the image to a certain number of pixels, which is helpful particularly for large
images, numTones specifies how many different frequencies should be used for a single pixel
row (more often than not more tones introduces strange phase issues that reduces the image
clarity, but this is very good for testing as for a value _t_ = numTones, the runtime is multiplied by
_t_ ), and contrast increases the difference between the darkest and brightest pixels which is
sometimes helpful for image clarity but does not impact runtime.

The order of these images in the JSON file is important as it determines the order in which the
images will appear in the output audio file.

Duration, Frequency, Height, and NumTones are the parameters that I manipulate to greatly
influence how long processing a given image takes a thread.

Each image is scaled and converted to grayscale. For every row of the image, the y value of that
row is mapped to a frequency based on the input parameters and the intensity of each pixel is
used to get the amplitude of the wave at that point. Each pixel is mapped to a certain number of
samples determined by the desired duration, the sample rate, and how many pixels are in the
resulting scaled image. Each sample of each row is added together and normalized, then the
program repeats for all images, concatenating their audio buffer results then exporting the final
file.

I use the free software Audacity to view the audio files as a spectrogram. There are a few tools
online that allow for viewing spectrograms such as this one (https://www.dcode.fr/spectral-analysis).
If you use this site or one like it I would suggest unchecking "Logarithmic Scale" and also
turning down your volume.