#!/usr/bin/env bash

go build


# this test is fun; type quickly and check that launched subprocesses are being
# reaped
./rl -x 'echo eep && sleep 5 && spd-say -w "hello"'
