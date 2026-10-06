#!/usr/bin/env python3
"""Local Riva SDK adapter for the Go server. Never serves HTTP or calls hosted AI."""
from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import wave


def requested_language(value: str) -> str:
    if value == "ru":
        return "ru-RU"
    if value == "kk":
        return "kk-KZ"
    if value == "mixed":
        raise ValueError("Mixed-language recognition has not been verified")
    return os.environ.get("ASR_LANGUAGE", "ru-RU")


def run(source: Path, endpoint: str, language: str) -> dict:
    import grpc
    import riva.client
    from riva.client.proto import riva_asr_pb2

    language = requested_language(language)
    demuxer = {".mp3": "mp3", ".wav": "wav", ".m4a": "mov", ".mp4": "mov",
               ".ogg": "ogg", ".opus": "ogg", ".aac": "aac",
               ".webm": "matroska", ".flac": "flac"}.get(source.suffix.lower())
    if not demuxer:
        raise ValueError("Unsupported audio format")
    ffmpeg = shutil.which("ffmpeg")
    if not ffmpeg:
        import imageio_ffmpeg
        ffmpeg = imageio_ffmpeg.get_ffmpeg_exe()
    os.umask(0o077)
    with tempfile.TemporaryDirectory(prefix="butterfly-riva-") as temp:
        decoded = Path(temp) / "audio.wav"
        # A fixed demuxer prevents uploaded playlists from loading other files.
        subprocess.run([ffmpeg, "-hide_banner", "-loglevel", "error", "-nostdin",
                        "-protocol_whitelist", "file,pipe", "-f", demuxer, "-i", str(source),
                        "-t", "3601", "-vn", "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le",
                        str(decoded)], check=True, capture_output=True, timeout=120)
        with wave.open(str(decoded), "rb") as wav:
            if wav.getnchannels() != 1 or wav.getsampwidth() != 2 or wav.getframerate() != 16000:
                raise ValueError("Invalid PCM format")
            duration = wav.getnframes() / 16000
            if not 0 < duration <= 3600:
                raise ValueError("Audio must be nonempty and at most 60 minutes")
            auth = riva.client.Auth(uri=endpoint, use_ssl=False, options=[("grpc.enable_http_proxy", 0)])
            try:
                grpc.channel_ready_future(auth.channel).result(timeout=20)
                service = riva.client.ASRService(auth)
                catalog = service.stub.GetRivaSpeechRecognitionConfig(
                    riva_asr_pb2.RivaSpeechRecognitionConfigRequest(), timeout=20)
                supported = [model for model in catalog.model_config if language in
                             {item.strip() for item in model.parameters.get("language_code", "").split(",")}]
                if not supported:
                    raise ValueError("Requested language is not advertised by the loaded ASR")
                config = riva.client.RecognitionConfig(
                    encoding=riva.client.AudioEncoding.LINEAR_PCM, sample_rate_hertz=16000,
                    audio_channel_count=1, language_code=language, model=supported[0].model_name,
                    max_alternatives=1, enable_automatic_punctuation=True, enable_word_time_offsets=True)
                segments = []
                offset = 0.0
                while True:
                    pcm = wav.readframes(60 * 16000)
                    if not pcm:
                        break
                    seconds = len(pcm) / 32000
                    response = service.stub.Recognize(
                        riva_asr_pb2.RecognizeRequest(config=config, audio=pcm), timeout=180)
                    for result in response.results:
                        if not result.alternatives:
                            continue
                        alternative = result.alternatives[0]
                        text = alternative.transcript.strip()
                        if not text:
                            continue
                        # Riva word offsets are milliseconds relative to this request.
                        words = alternative.words
                        start = max(0.0, min(seconds, words[0].start_time / 1000)) if words else 0.0
                        end = max(start, min(seconds, words[-1].end_time / 1000)) if words else seconds
                        segments.append({"text": text, "start": offset + start, "end": offset + end})
                    offset += seconds
                if not segments:
                    raise ValueError("No speech recognized")
                return {"segments": segments, "language": language, "diarization": False,
                        "duration_seconds": duration, "chunk_seconds": 60}
            finally:
                auth.channel.close()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--endpoint", required=True)
    parser.add_argument("--language", default="auto")
    args = parser.parse_args()
    try:
        print(json.dumps(run(args.input.resolve(), args.endpoint, args.language), ensure_ascii=False))
        return 0
    except Exception as exc:
        # No traceback, transcripts, paths, endpoints or credentials in process output.
        print(json.dumps({"status": "failed", "error_type": type(exc).__name__}), file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
