
# rl [![CI](https://github.com/rgrannell1/rl/actions/workflows/ci.yaml/badge.svg)](https://github.com/rgrannell1/rl/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/rgrannell1/rl)](https://goreportcard.com/report/github.com/rgrannell1/rl)

rl (readline) allows users to run commands like grep interactively.

[![asciicast](https://asciinema.org/a/jMkJl36C46dvv12lsMuZwWd0L.svg)](https://asciinema.org/a/jMkJl36C46dvv12lsMuZwWd0L)

For example, `rl` can interactively search files for a keyword in your notes folder, then open your final matches in VSCode.

```bash
rl -x 'grep -rl $RL_INPUT ~/Notes' | xargs -I % code %
```

## Build

```bash
bs build                      # live-build using entr
bs install                    # install to /usr/bin
```
## Installation

```bash
git clone https://github.com/rgrannell1/rl.git ~/.rl
~/.rl/bs/install.sh
```

## License

The MIT License

Copyright (c) 2021 Róisín Grannell

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
