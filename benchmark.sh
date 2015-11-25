#! /bin/bash

# Programs and configs to execute
programs=("loop_bw" "loop_fw", "bubble_sort_bw" "bubble_sort_fw")
configs=("nehalem_stall" "nehalem_always" "nehalem_never" "nehalem_forward" "nehalem_backward", "nehalem_one_bit")

# Run benchmarks
for p in ${programs[@]}; do
    for c in ${configs[@]}; do
        bin/simulator.exe run samples/programs/sample_${p}.txt -c samples/configs/${c}.config -o benchmark/${p}/${c} --max-cycles 2000
    done
done

# Merge output files
stats_directory="benchmark/stats"
mkdir "${stats_directory}"

for file in benchmark/*/*/output.log; do
    program="${file#*/}"
    program="${program%%/*}"

    config="${file#*/}"
    config="${config#*/}"
    config="${config%%/*}"

    target_file_name="${program}_${config}.log"
    cp "$file" "${stats_directory}/${target_file_name}"
done