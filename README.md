# Petcam

A simple petcam using a Raspberry Pi 1 + Logitech C930e / C920.

The hardware sets the constraints:

- MJPEG streaming instead of h264 or other codecs since this doesn't
  require transcoding (the camera supplies a MJPEG stream) and we
  can use existing software (ustreamer)

- for audio streaming littel transcoding (LAME) and a small golang
  server for piping since python would be too slow on the hardware

## Architecture

```
[ustreamer /dev/video0]
[    :8080/stream     ]
         ^
         | starts
         |
   [flask server]- starts ->[audioserver C930e]
   [   :5000/   ]           [  :8081/stream   ]
         |
         | serves
         v
   [ index.html ]
   [ MJPEG streaming using browser / chunked reads ]
   [ audio streaming using WebAudio + JS           ]

```

The audioserver simply invokes ffmpeg and streams from ffmpeg to the
`/stream` HTTP endpoint so that we don't have to use the flaky ffmpeg
HTTP serving which breaks when seeking, for example.

# Installation

External packages:

    pip install flask
    sudo apt install ustreamer ffmpeg

Internal packages:

    cd audioserver
    make
    cp audioserver ../

petcam expects the audioserver executable to be in the same folder as `app.py`.
You will need a working go setup but there are no external dependencies.

# Configuration

If you use a different camera than the Logitech C930e or the
ALSA audio identifiers differ on your system you need to replace
the "plughw:CARD=C930e,DEV=0" string with the corresponding
output of `arecord -L` in `app.py`.
