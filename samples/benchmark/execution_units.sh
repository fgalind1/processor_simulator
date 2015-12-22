#! /bin/bash
rm -rf "benchmark"

# Programs and configs to execute
configprefix="execution_units/"
configs=("3" "4" "5" "6" "7" "8" "9" "11" "10" "12" "13" "14")
programs=("load_store_alu_op" "loop_vectors_forward" "inner_product" "factorial" "fibonacci" "bubble_sort_forward")

# Run benchmarks
for p in ${programs[@]}; do
    for c in ${configs[@]}; do
        bin/simulator.exe run samples/programs/${p}.asm -c samples/configs/${configprefix}${c}.config -o benchmark/${p}/${configprefix}${c} --max-cycles 2000 &
    done
done