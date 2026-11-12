#!/usr/bin/env python3
"""Verify checksums/archive contents; smoke-test and install the host's binaries."""
import argparse
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import platform
import subprocess
import tarfile
import tempfile

ROOT=Path(__file__).resolve().parents[1]

def main():
    parser=argparse.ArgumentParser()
    parser.add_argument('directory',type=Path,nargs='?',default=ROOT/'dist')
    args=parser.parse_args()
    system=platform.system().lower()
    arch={'x86_64':'amd64','aarch64':'arm64','arm64':'arm64'}.get(platform.machine())
    expected_version=(ROOT/'VERSION').read_text().strip()
    entries=(args.directory/'SHA256SUMS').read_text().splitlines()
    assert len(entries)==4, 'expected four platform archives'
    verified=[]
    native=None
    for entry in entries:
        digest,name=entry.split('  ',1)
        assert PurePosixPath(name).name==name, 'unsafe checksum path'
        path=args.directory/name
        assert hashlib.sha256(path.read_bytes()).hexdigest()==digest, f'checksum mismatch: {name}'
        with tarfile.open(path) as archive:
            members=archive.getmembers()
            for member in members:
                assert member.isfile() and not PurePosixPath(member.name).is_absolute() and '..' not in PurePosixPath(member.name).parts
            names={PurePosixPath(m.name).name for m in members}
            assert {'myLlama','myLlamactl','install.sh','README.md','VERSION'} <= names
        verified.append(name)
        if name.endswith(f'_{system}_{arch}.tar.gz'):
            native=path
    assert len(set(verified))==4
    assert native, 'no archive for host platform'
    with tempfile.TemporaryDirectory(prefix='myLlama-install-') as temp:
        temp=Path(temp)
        with tarfile.open(native) as archive:
            archive.extractall(temp,filter='data')
        folder=next(p for p in temp.iterdir() if p.is_dir())
        installed=temp/'installed'
        subprocess.run(['sh',str(folder/'install.sh')],env=dict(os.environ,INSTALL_DIR=str(installed)),check=True)
        for name in ['myLlama','myLlamactl']:
            info=json.loads(subprocess.check_output([str(installed/name),'--version']))
            assert info['version']==expected_version, info
        subprocess.run([str(installed/'myLlamactl'),'init','--directory',str(temp/'config')],check=True,stdout=subprocess.DEVNULL)
    print('Verified checksums, four archives, native versions and local installation.')

if __name__=='__main__':
    main()
