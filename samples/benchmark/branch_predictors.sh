#! /bin/bash
rm -rf "benchmark"

# Programs and configs to execute
configprefix="branch_predictors/"
configs=("stall" "always" "never" "backward" "forward" "one_bit" "two_bit")
programs=("loop_vectors_forward" "loop_vectors_backward" "bubble_sort_forward" "bubble_sort_backward")

# Run benchmarks
for p in ${programs[@]}; do
    for c in ${configs[@]}; do
        bin/simulator.exe run samples/programs/${p}.asm -c samples/configs/${configprefix}${c}.config -o benchmark/${p}/${configprefix}${c} --max-cycles 1800 &
    done
done