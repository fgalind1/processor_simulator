#! /bin/bash
rm -rf "benchmark"

# Programs and configs to execute
configprefix="renaming/"
configs=("disabled" "2" "4" "6" "8" "10" "12" "14" "16" "20" "25" "30")
programs=("inner_product" "fibonacci" "bubble_sort_forward")

# Run benchmarks
for p in ${programs[@]}; do
    for c in ${configs[@]}; do
        bin/simulator.exe run samples/programs/${p}.asm -c samples/configs/${configprefix}${c}.config -o benchmark/${p}/${configprefix}${c} --max-cycles 1800 &
    done
done