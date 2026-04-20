import atexit
import os
import subprocess

from urllib.parse import urlparse

from flask import Flask, request, Response

app = Flask(__name__)

SCRIPT_DIR = os.path.dirname(os.path.realpath(__file__))

VIDEO_W = 640
VIDEO_H = 480
STREAM_VIDEO_PORT = 8080
STREAM_AUDIO_PORT = 8081
VIDEO_DEVICE = "/dev/video0"
AUDIO_DEVICE = "plughw:CARD=C930e,DEV=0"

streamer_video_process = None
streamer_audio_process = None


def start_video_streamer():
    global streamer_video_process

    streamer_video_process = subprocess.Popen([
        "ustreamer",
        "-d", VIDEO_DEVICE,
        "-r", f"{VIDEO_W}x{VIDEO_H}",
        "-m", "MJPEG",
        "-f", "5",
        "-p", f"{STREAM_VIDEO_PORT}",
        "-s", "0.0.0.0",
    ])


def stop_video_streamer():
    global streamer_video_process
    if streamer_video_process:
        streamer_video_process.terminate()


def start_audio_streamer():
    global streamer_audio_process
    
    streamer_audio_process = subprocess.Popen([
        os.path.join(SCRIPT_DIR, "audioserver"),
        "-audioDevice", AUDIO_DEVICE,
        "-port", f"{STREAM_AUDIO_PORT}",
    ])


def stop_audio_streamer():
    global streamer_audio_process
    if streamer_audio_process:
        streamer_audio_process.terminate()


atexit.register(stop_video_streamer)
atexit.register(stop_audio_streamer)

@app.route("/audio")
def audio():
    return Response(audio_stream(), mimetype="audio/mpeg")



@app.route("/")
def index():
    hostname = urlparse(request.base_url).hostname
    with open("index.html") as f:
        return f.read().format(HOSTNAME=hostname, STREAM_VIDEO_PORT=STREAM_VIDEO_PORT, STREAM_AUDIO_PORT=STREAM_AUDIO_PORT)


if __name__ == "__main__":
    start_video_streamer()
    start_audio_streamer()
    app.run(host="0.0.0.0", port=5000, threaded=True)
