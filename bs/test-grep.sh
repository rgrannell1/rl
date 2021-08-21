
go build || return 1

# read input; dynamically search against a folder with xargs
./rl --clear | xargs -L 1 | xargs -rI % grep --color=always -rH ".%" -- ~/Drive/Obsidian/*
