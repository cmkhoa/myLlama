"""Real Python SDK -> myLlama binary -> managed mock runner, with no cloud calls."""
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request

from openai import OpenAI, NotFoundError, BadRequestError

ROOT = Path(__file__).resolve().parents[2]

def free_port():
    with socket.socket() as sock:
        sock.bind(('127.0.0.1', 0))
        return sock.getsockname()[1]

def wait_for(check, timeout=10):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        try:
            if check():
                return
        except (OSError, urllib.error.URLError):
            pass
        time.sleep(0.05)
    raise AssertionError('timed out waiting for local service')

def main():
    fixture = ROOT / 'tests/compat/mock_runner.py'
    fixture.chmod(0o755)
    with tempfile.TemporaryDirectory(prefix='myLlama-sdk-') as directory:
        directory = Path(directory)
        model = directory / 'fixture.gguf'
        model.write_bytes(b'GGUF')
        api_port, runner_port = free_port(), free_port()
        while runner_port == api_port:
            runner_port = free_port()
        address = f'http://127.0.0.1:{api_port}'
        def call(path, payload=None):
            request = urllib.request.Request(address + path,
                data=None if payload is None else json.dumps(payload).encode(),
                headers={'Content-Type': 'application/json'})
            with urllib.request.urlopen(request, timeout=15) as response:
                return json.load(response)
        env = {k:v for k,v in os.environ.items() if not k.startswith('MYLLAMA_')}
        env.update(MYLLAMA_LISTEN_ADDRESS=f'127.0.0.1:{api_port}',
            MYLLAMA_RUNNER_ADDRESS=f'127.0.0.1:{runner_port}',
            MYLLAMA_LLAMA_SERVER_PATH=str(fixture), MYLLAMA_IDLE_TIMEOUT='1s',
            MYLLAMA_LOG_PATH=str(directory / 'myLlama.log'))
        process = subprocess.Popen([str(ROOT / 'bin/myLlama')], env=env,
                                   stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
        child_pid = None
        try:
            wait_for(lambda: call('/v1/runners/status')['state'] == 'stopped')
            call('/v1/models/register', {'path': str(model)})
            child_pid = call('/v1/runners/start', {})['pid']
            with OpenAI(base_url=address + '/v1', api_key='local-test-only', max_retries=0, timeout=10) as client:
                assert client.models.list().data[0].id == 'fixture.gguf'
                params = dict(model='fixture.gguf', messages=[{'role':'user','content':'private test prompt'}], max_tokens=8)
                result = client.chat.completions.create(**params)
                assert result.choices[0].message.content == 'Hello locally.'
                assert result.usage.completion_tokens == 3
                with client.chat.completions.create(**params, stream=True,
                        stream_options={'include_usage':True}) as stream:
                    chunks = list(stream)
                assert ''.join(c.choices[0].delta.content or '' for c in chunks if c.choices) == 'Hello locally.'
                assert chunks[-1].usage.completion_tokens == 3
                try:
                    client.chat.completions.create(**dict(params, model='missing'))
                except NotFoundError as exc:
                    assert exc.body['code'] == 'model_not_found'
                else:
                    raise AssertionError('unknown model was accepted')
                try:
                    client.chat.completions.create(**dict(params, max_tokens=999999))
                except BadRequestError:
                    pass
                else:
                    raise AssertionError('invalid token limit was accepted')
            result = subprocess.run([str(ROOT/'bin/myLlamactl'),'--address',address,'--json','chat','hello'],
                                    capture_output=True, text=True, check=True)
            assert json.loads(result.stdout)['usage']['completion_tokens'] == 3
            wait_for(lambda: call('/v1/runners/status')['state'] == 'stopped')
            child_pid = call('/v1/runners/start', {})['pid']
            process.send_signal(signal.SIGTERM)
            assert process.wait(timeout=15) == 0
            def child_exited():
                try:
                    os.kill(child_pid, 0)
                    return False
                except ProcessLookupError:
                    return True
            wait_for(child_exited)
            log = (directory/'myLlama.log').read_text()
            assert 'private test prompt' not in log and 'Hello locally.' not in log
            assert 'idle runner stopped' in log
        finally:
            if process.poll() is None:
                process.terminate()
                try:
                    process.wait(timeout=15)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait(timeout=5)
            if child_pid:
                try:
                    os.kill(child_pid, signal.SIGTERM)
                except ProcessLookupError:
                    pass
        print('SDK, CLI JSON, idle shutdown, process cleanup and log privacy checks passed.')

if __name__ == '__main__':
    main()
