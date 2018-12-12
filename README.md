# reMP3

## Building
```
docker build -t larrabee/remp3 .
```

## Usage
### Resize MP3 with POST request:
```
>> curl -X POST 'http://127.0.0.1:7090?bitrate=48000&samplerate=44100' --data-binary "@/tmp/file.mp3" -o /tmp/resized.mp3
```
Accepted args:
 * **bitrate**: output file bitrate, int, required.
 * **samplerate**: output file sample rate, int, required.
 * **duration**:  crop output file to first N seconds, int (seconds), optional, 0 by default.


### Resize MP3 with GET request:
```
>> curl 'http://127.0.0.1:7090/backingtracks/tracks/6/62d077bdd8e354c07aaf7a46c79123b205660eac.mp3?bitrate=48000&samplerate=44100' -H 'x-resize-base: ustatik.com'
```

Accepted args:
 * **bitrate**: output file bitrate, int, required.
 * **samplerate**: output file sample rate, int, required.
 * **duration**:  crop output file to first N seconds, int (seconds), optional, 0 by default.

Accepted headers:
 * **x-resize-base**: host with source file.
 * **x-resize-scheme**: source server schema, possible values: "http", "https". Default: "https".