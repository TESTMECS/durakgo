clean:
    rm debug.log && touch debug.log
build:
    go build . 
logs:
    bat debug.log
