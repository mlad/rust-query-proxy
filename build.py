import subprocess
import os
import sys

targets = [
    ('linux', '386', 'proxy_linux32'),
    ('linux', 'amd64', 'proxy_linux64'),
    ('linux', 'arm64', 'proxy_arm64'),
    ('windows', 'amd64', 'proxy_win64.exe'),
]

def main():
    if len(sys.argv) > 1 and sys.argv[1] == 'all':
        for (go_os, go_arch, out_filename) in targets:
            build(go_os, go_arch, 'bin/%s' % out_filename)
    else:
        build('linux', 'amd64', 'proxy')

    print('Success')

def build(go_os, go_arch, out_path):
    env = dict(os.environ)
    env['GOOS'] = go_os
    env['GOARCH'] = go_arch

    print('Building %s %s...' % (go_os, go_arch))
    try:
        subprocess.run(('go', 'build', '-trimpath', '-ldflags=-s -w', '-o=' + out_path), env=env, check=True)
    except subprocess.CalledProcessError:
        print('Build error')
        return

if __name__ == '__main__':
    main()
