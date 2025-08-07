clean:
    rm debug.log && touch debug.log
build:
    go build . 
logs:
    less debug.log
