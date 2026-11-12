#!/usr/bin/env python3
"""Build portable release archives using only Go and the Python standard library."""
import argparse
import gzip
import hashlib
import io
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[1]
TARGETS = [('darwin','amd64'),('darwin','arm64'),('linux','amd64'),('linux','arm64')]
PATTERN = r'(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-rc\.([1-9][0-9]*))?'

def git(*args, fallback):
    try:
        return subprocess.check_output(['git',*args],cwd=ROOT,stderr=subprocess.DEVNULL,text=True).strip()
    except subprocess.CalledProcessError:
        return fallback

def main():
    parser=argparse.ArgumentParser()
    parser.add_argument('--version', default=(ROOT/'VERSION').read_text().strip())
    parser.add_argument('--output', type=Path, default=ROOT/'dist')
    args=parser.parse_args()
    version=args.version.removeprefix('v')
    if not re.fullmatch(PATTERN,version):
        parser.error('version must be X.Y.Z or X.Y.Z-rc.N')
    if version != (ROOT/'VERSION').read_text().strip():
        parser.error('version must match VERSION')
    if os.environ.get('GITHUB_REF_TYPE') == 'tag' and os.environ['GITHUB_REF_NAME'] != 'v'+version:
        parser.error('tag must equal v + VERSION')
    args.output.mkdir(parents=True,exist_ok=True)
    if any(args.output.iterdir()):
        parser.error('output directory must be empty; choose a fresh directory')
    commit=git('rev-parse','HEAD',fallback='unknown')
    epoch=int(os.environ.get('SOURCE_DATE_EPOCH',git('show','-s','--format=%ct','HEAD',fallback='0')))
    date=datetime.fromtimestamp(epoch,timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    flags=f'-s -w -X myLlama/internal/buildinfo.Version={version} -X myLlama/internal/buildinfo.Commit={commit} -X myLlama/internal/buildinfo.Date={date}'
    with tempfile.TemporaryDirectory(prefix='myLlama-release-') as temp:
        for system,arch in TARGETS:
            stem=f'myLlama_{version}_{system}_{arch}'
            archive=args.output/(stem+'.tar.gz')
            env=dict(os.environ,CGO_ENABLED='0',GOOS=system,GOARCH=arch)
            files={}
            for command in ['myLlama','myLlamactl']:
                binary=Path(temp)/command
                subprocess.run(['go','build','-trimpath','-buildvcs=false','-ldflags',flags,'-o',str(binary),'./cmd/'+command],cwd=ROOT,env=env,check=True)
                files[command]=(binary.read_bytes(),0o755)
            documents={name:name for name in ['README.md','LICENSE','CHANGELOG.md','VERSION','ROADMAP.md','CONTRIBUTING.md','DESIGN_LOG.md']}
            documents.update({str(path.relative_to(ROOT)):str(path.relative_to(ROOT)) for path in (ROOT/'docs').glob('*.md')})
            documents.update({'install.sh':'scripts/install.sh','configs/profiles.example.json':'configs/profiles.example.json'})
            for name,path in documents.items():
                files[name]=((ROOT/path).read_bytes(),0o755 if name=='install.sh' else 0o644)
            with archive.open('wb') as raw:
                with gzip.GzipFile(filename='',mode='wb',fileobj=raw,mtime=epoch) as compressed:
                    with tarfile.open(fileobj=compressed,mode='w') as tar:
                        for name,(data,mode) in sorted(files.items()):
                            info=tarfile.TarInfo(stem+'/'+name)
                            info.size,info.mode,info.mtime=len(data),mode,epoch
                            tar.addfile(info,io.BytesIO(data))
            print(archive.name,flush=True)
    lines=[hashlib.sha256(path.read_bytes()).hexdigest()+'  '+path.name for path in sorted(args.output.glob('*.tar.gz'))]
    (args.output/'SHA256SUMS').write_text('\n'.join(lines)+'\n')

if __name__=='__main__':
    main()
