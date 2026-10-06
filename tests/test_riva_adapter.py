"""The formats accepted by uploads must reach Riva's fixed-demuxer pipeline."""
from types import ModuleType, SimpleNamespace
import sys
import wave

import pytest

from scripts import riva_transcribe


@pytest.mark.parametrize("extension,demuxer", [(".aac", "aac"), (".opus", "ogg")])
def test_supported_audio_formats_decode_before_recognition(monkeypatch, tmp_path, extension, demuxer):
    source = tmp_path / ("recording" + extension)
    source.write_bytes(b"synthetic input; ffmpeg is mocked")
    calls = []

    def decode(command, **kwargs):
        assert command[command.index("-f") + 1] == demuxer
        assert command[command.index("-protocol_whitelist") + 1] == "file,pipe"
        assert command[command.index("-i") + 1] == str(source)
        assert kwargs["timeout"] == 120
        with wave.open(command[-1], "wb") as wav:
            wav.setnchannels(1)
            wav.setsampwidth(2)
            wav.setframerate(16000)
            wav.writeframes(b"\x00\x00" * 1600)
        calls.append("decoded")

    class Stub:
        def GetRivaSpeechRecognitionConfig(self, request, timeout):
            return SimpleNamespace(model_config=[SimpleNamespace(
                model_name="test-russian", parameters={"language_code": "ru-RU"})])

        def Recognize(self, request, timeout):
            assert calls == ["decoded"]
            calls.append("recognized")
            return SimpleNamespace(results=[SimpleNamespace(alternatives=[
                SimpleNamespace(transcript="Обсудили план.", words=[])])])

    channel = SimpleNamespace(close=lambda: calls.append("closed"))
    riva = ModuleType("riva")
    client = ModuleType("riva.client")
    client.Auth = lambda **kwargs: SimpleNamespace(channel=channel)
    client.ASRService = lambda auth: SimpleNamespace(stub=Stub())
    client.RecognitionConfig = lambda **kwargs: kwargs
    client.AudioEncoding = SimpleNamespace(LINEAR_PCM=1)
    riva.client = client
    proto = ModuleType("riva.client.proto")
    proto.riva_asr_pb2 = SimpleNamespace(
        RivaSpeechRecognitionConfigRequest=lambda: None,
        RecognizeRequest=lambda **kwargs: kwargs,
    )
    grpc = ModuleType("grpc")
    grpc.channel_ready_future = lambda channel: SimpleNamespace(result=lambda timeout: None)
    for name, module in [("riva", riva), ("riva.client", client), ("riva.client.proto", proto), ("grpc", grpc)]:
        monkeypatch.setitem(sys.modules, name, module)
    monkeypatch.setattr(riva_transcribe.shutil, "which", lambda name: "ffmpeg")
    monkeypatch.setattr(riva_transcribe.subprocess, "run", decode)
    monkeypatch.setattr(riva_transcribe.os, "umask", lambda mode: 0)

    result = riva_transcribe.run(source, "localhost:50051", "ru")
    assert result["segments"] == [{"text": "Обсудили план.", "start": 0.0, "end": 0.1}]
    assert result["diarization"] is False
    assert calls == ["decoded", "recognized", "closed"]
