#!/usr/bin/env python3
"""Deterministic test fixture, not an inference server. Accepts runner flags."""
import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import time

class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_GET(self):
        self.send_response(200 if self.path == '/health' else 404)
        self.end_headers()
        self.wfile.write(b'{}')

    def do_POST(self):
        request = json.loads(self.rfile.read(int(self.headers['Content-Length'])))
        model = request.get('model') or 'fixture.gguf'
        base = {'id': 'chatcmpl-fixture', 'created': int(time.time()), 'model': model}
        self.send_response(200)
        if not request.get('stream'):
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(dict(base, object='chat.completion', choices=[{
                'index': 0, 'message': {'role': 'assistant', 'content': 'Hello locally.'},
                'finish_reason': 'stop'}], usage={'prompt_tokens': 2, 'completion_tokens': 3, 'total_tokens': 5})).encode())
            return
        self.send_header('Content-Type', 'text/event-stream')
        self.end_headers()
        def send(payload):
            self.wfile.write(('data: ' + json.dumps(payload) + '\n\n').encode())
            self.wfile.flush()
        for content in ['Hello ', 'locally.']:
            send(dict(base, object='chat.completion.chunk', choices=[{
                'index': 0, 'delta': {'content': content}, 'finish_reason': None}]))
        send(dict(base, object='chat.completion.chunk', choices=[{
            'index': 0, 'delta': {}, 'finish_reason': 'stop'}]))
        if request.get('stream_options', {}).get('include_usage'):
            send(dict(base, object='chat.completion.chunk', choices=[], usage={
                'prompt_tokens': 2, 'completion_tokens': 3, 'total_tokens': 5}))
        self.wfile.write(b'data: [DONE]\n\n')
        self.wfile.flush()

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--host', default='127.0.0.1')
    parser.add_argument('--port', type=int, required=True)
    args, _ = parser.parse_known_args()
    try:
        ThreadingHTTPServer((args.host, args.port), Handler).serve_forever()
    except KeyboardInterrupt:
        pass
