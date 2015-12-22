#! /bin/bash
rm -rf "benchmark"

# Programs and configs to execute
configprefix="superscalar/"
configs=("scalar_pipelined" "superscalar_2_way" "superscalar_4_way" "superscalar_6_way")
programs=("load_store_alu_op" "factorial" "inner_product" "fibonacci" "loop_vectors_forward" "bubble_sort_forward")

# Run benchmarks
for p in ${programs[@]}; do
    for c in ${configs[@]}; do
        bin/simulator.exe run samples/programs/${p}.asm -c samples/configs/${configprefix}${c}.config -o benchmark/${p}/${configprefix}${c} --max-cycles 3000 &
    done
done