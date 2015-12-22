#! /bin/bash
rm -rf "benchmark"

# Programs and configs to execute
configprefix="reorder_buffer_size/"
configs=("1" "5" "10" "15" "20" "25" "30" "35" "40" "45" "50" "55" "60")
programs=("inner_product" "fibonacci" "loop_vectors_forward" "bubble_sort_forward")

# Run benchmarks
for p in ${programs[@]}; do
    for c in ${configs[@]}; do
        bin/simulator.exe run samples/programs/${p}.asm -c samples/configs/${configprefix}${c}.config -o benchmark/${p}/${configprefix}${c} --max-cycles 1000 &
    done
done