#!/bin/bash

alice=$(checkersd keys show alice -a)
bob=$(checkersd keys show bob -a)

checkersd tx checkers create-game $alice $bob --from $alice  --dry-run
