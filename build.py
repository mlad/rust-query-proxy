import subprocess
import os

def main():
    env = dict(os.environ)
    env['GOOS'] = 'linux'
    env['GOARCH'] = '386'

    print('Building...')
    try:
        subprocess.run(('go', 'build', '-trimpath', '-ldflags=-s -w'), env=env, check=True)
    except subprocess.CalledProcessError:
        print('Build error')
        return

    print('Success')


if __name__ == '__main__':
    main()
